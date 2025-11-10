package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStorageConfig 文件存储配置
type FileStorageConfig struct {
	FilePath      string        // 文件路径
	FlushInterval time.Duration // 刷新间隔（0表示每次立即写入）
	BufferSize    int           // 缓冲区大小（并发保存请求队列）
}

// FileStorage 文件存储实现
type FileStorage struct {
	config FileStorageConfig

	// 批量写入控制
	dirty       bool          // 脏标记
	pendingSave chan struct{} // 保存请求通道
	stopFlush   chan struct{} // 停止定时刷新
	wg          sync.WaitGroup

	// 最新数据缓存（避免重复读文件）
	lastData *MappingData
	dataMu   sync.RWMutex
}

// NewFileStorage 创建文件存储
func NewFileStorage(config FileStorageConfig) *FileStorage {
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second // 默认5秒刷新
	}
	if config.BufferSize == 0 {
		config.BufferSize = 100 // 默认缓冲100个保存请求
	}

	fs := &FileStorage{
		config:      config,
		pendingSave: make(chan struct{}, config.BufferSize),
		stopFlush:   make(chan struct{}),
	}

	// 启动异步刷新协程
	fs.wg.Add(1)
	go fs.flushLoop()

	return fs
}

// Load 加载数据
func (fs *FileStorage) Load() (*MappingData, error) {
	fs.dataMu.RLock()
	if fs.lastData != nil {
		// 返回缓存数据
		defer fs.dataMu.RUnlock()
		return fs.copyMappingData(fs.lastData), nil
	}
	fs.dataMu.RUnlock()

	// 从文件加载
	data, err := fs.loadFromFile()
	if err != nil {
		return nil, err
	}

	// 缓存数据
	fs.dataMu.Lock()
	fs.lastData = data
	fs.dataMu.Unlock()

	return fs.copyMappingData(data), nil
}

// Save 保存数据（异步，全量保存）
func (fs *FileStorage) Save(data *MappingData) error {
	// 更新缓存
	fs.dataMu.Lock()
	fs.lastData = fs.copyMappingData(data)
	fs.dirty = true
	fs.dataMu.Unlock()

	// 发送保存请求（非阻塞）
	select {
	case fs.pendingSave <- struct{}{}:
		// 发送成功
	default:
		// 缓冲区满，跳过（定时刷新会处理）
		log.Printf("⚠️  保存请求缓冲区满，将在下次定时刷新时保存")
	}

	return nil
}

// AddUserMapping 增量添加用户映射
func (fs *FileStorage) AddUserMapping(userID, ip string) error {
	fs.dataMu.Lock()
	defer fs.dataMu.Unlock()

	// 确保数据已初始化
	if fs.lastData == nil {
		fs.lastData = &MappingData{
			UserToIP:  make(map[string]string),
			IPToUsers: make(map[string][]string),
		}
	}

	// 更新正向映射
	fs.lastData.UserToIP[userID] = ip

	// 更新反向映射
	if fs.lastData.IPToUsers[ip] == nil {
		fs.lastData.IPToUsers[ip] = make([]string, 0)
	}
	// 检查是否已存在（避免重复）
	found := false
	for _, uid := range fs.lastData.IPToUsers[ip] {
		if uid == userID {
			found = true
			break
		}
	}
	if !found {
		fs.lastData.IPToUsers[ip] = append(fs.lastData.IPToUsers[ip], userID)
	}

	// 标记脏数据
	fs.dirty = true

	// 触发异步保存
	select {
	case fs.pendingSave <- struct{}{}:
	default:
	}

	return nil
}

// RemoveUserMapping 增量删除用户映射
func (fs *FileStorage) RemoveUserMapping(userID, ip string) error {
	fs.dataMu.Lock()
	defer fs.dataMu.Unlock()

	if fs.lastData == nil {
		return nil // 没有数据，无需删除
	}

	// 删除正向映射
	delete(fs.lastData.UserToIP, userID)

	// 删除反向映射
	if users, exists := fs.lastData.IPToUsers[ip]; exists {
		newUsers := make([]string, 0, len(users)-1)
		for _, uid := range users {
			if uid != userID {
				newUsers = append(newUsers, uid)
			}
		}
		fs.lastData.IPToUsers[ip] = newUsers
	}

	// 标记脏数据
	fs.dirty = true

	// 触发异步保存
	select {
	case fs.pendingSave <- struct{}{}:
	default:
	}

	return nil
}

// Flush 强制刷新
func (fs *FileStorage) Flush() error {
	fs.dataMu.Lock()
	if !fs.dirty || fs.lastData == nil {
		fs.dataMu.Unlock()
		return nil
	}

	data := fs.copyMappingData(fs.lastData)
	fs.dirty = false
	fs.dataMu.Unlock()

	return fs.saveToFile(data)
}

// Close 关闭存储
func (fs *FileStorage) Close() error {
	// 停止定时刷新
	close(fs.stopFlush)
	fs.wg.Wait()

	// 最后一次刷新
	return fs.Flush()
}

// flushLoop 异步刷新循环
func (fs *FileStorage) flushLoop() {
	defer fs.wg.Done()

	ticker := time.NewTicker(fs.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fs.pendingSave:
			// 收到保存请求，立即刷新
			if err := fs.Flush(); err != nil {
				log.Printf("⚠️  保存映射失败: %v", err)
			}

		case <-ticker.C:
			// 定时刷新
			if err := fs.Flush(); err != nil {
				log.Printf("⚠️  定时刷新映射失败: %v", err)
			}

		case <-fs.stopFlush:
			return
		}
	}
}

// loadFromFile 从文件加载
func (fs *FileStorage) loadFromFile() (*MappingData, error) {
	data, err := os.ReadFile(fs.config.FilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// 文件不存在，返回空数据
			return &MappingData{
				UserToIP:  make(map[string]string),
				IPToUsers: make(map[string][]string),
			}, nil
		}
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var mappingData MappingData
	if err := json.Unmarshal(data, &mappingData); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	log.Printf("✓ 从文件加载映射: %d个用户, %d个IP",
		len(mappingData.UserToIP), len(mappingData.IPToUsers))

	return &mappingData, nil
}

// saveToFile 保存到文件
func (fs *FileStorage) saveToFile(data *MappingData) error {
	log.Printf("💾 保存映射到文件: %d个用户, %d个IP", len(data.UserToIP), len(data.IPToUsers))
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化JSON失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(fs.config.FilePath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	// 原子写入（先写临时文件，再重命名）
	tempFile := fs.config.FilePath + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tempFile, fs.config.FilePath); err != nil {
		return fmt.Errorf("重命名文件失败: %w", err)
	}
	log.Printf("✓ 映射保存成功: %s", fs.config.FilePath)

	return nil
}

// copyMappingData 深拷贝映射数据
func (fs *FileStorage) copyMappingData(data *MappingData) *MappingData {
	if data == nil {
		return &MappingData{
			UserToIP:  make(map[string]string),
			IPToUsers: make(map[string][]string),
		}
	}

	result := &MappingData{
		UserToIP:  make(map[string]string, len(data.UserToIP)),
		IPToUsers: make(map[string][]string, len(data.IPToUsers)),
	}

	for k, v := range data.UserToIP {
		result.UserToIP[k] = v
	}

	for k, v := range data.IPToUsers {
		users := make([]string, len(v))
		copy(users, v)
		result.IPToUsers[k] = users
	}

	return result
}
