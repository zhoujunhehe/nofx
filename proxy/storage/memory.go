package storage

import "sync"

// MemoryStorage 内存存储实现（仅用于测试）
type MemoryStorage struct {
	data *MappingData
	mu   sync.RWMutex
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: &MappingData{
			UserToIP:  make(map[string]string),
			IPToUsers: make(map[string][]string),
		},
	}
}

// Load 加载数据
func (m *MemoryStorage) Load() (*MappingData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	return m.copyData(m.data), nil
}

// Save 保存数据
func (m *MemoryStorage) Save(data *MappingData) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data = m.copyData(data)
	return nil
}

// AddUserMapping 添加用户映射
func (m *MemoryStorage) AddUserMapping(userID, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新正向映射
	m.data.UserToIP[userID] = ip

	// 更新反向映射
	if m.data.IPToUsers[ip] == nil {
		m.data.IPToUsers[ip] = make([]string, 0)
	}

	// 检查是否已存在
	found := false
	for _, uid := range m.data.IPToUsers[ip] {
		if uid == userID {
			found = true
			break
		}
	}

	if !found {
		m.data.IPToUsers[ip] = append(m.data.IPToUsers[ip], userID)
	}

	return nil
}

// RemoveUserMapping 移除用户映射
func (m *MemoryStorage) RemoveUserMapping(userID, ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 移除正向映射
	delete(m.data.UserToIP, userID)

	// 移除反向映射
	if users, exists := m.data.IPToUsers[ip]; exists {
		newUsers := make([]string, 0, len(users)-1)
		for _, uid := range users {
			if uid != userID {
				newUsers = append(newUsers, uid)
			}
		}
		m.data.IPToUsers[ip] = newUsers
	}

	return nil
}

// Flush 刷新（内存存储无需刷新）
func (m *MemoryStorage) Flush() error {
	return nil
}

// Close 关闭（内存存储无需关闭）
func (m *MemoryStorage) Close() error {
	return nil
}

// copyData 深拷贝数据
func (m *MemoryStorage) copyData(data *MappingData) *MappingData {
	if data == nil {
		return &MappingData{
			UserToIP:  make(map[string]string),
			IPToUsers: make(map[string][]string),
		}
	}

	copied := &MappingData{
		UserToIP:  make(map[string]string),
		IPToUsers: make(map[string][]string),
	}

	for k, v := range data.UserToIP {
		copied.UserToIP[k] = v
	}

	for k, v := range data.IPToUsers {
		copiedUsers := make([]string, len(v))
		copy(copiedUsers, v)
		copied.IPToUsers[k] = copiedUsers
	}

	return copied
}
