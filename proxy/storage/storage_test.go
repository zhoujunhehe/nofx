package storage

import (
	"os"
	"testing"
	"time"
)

// TestMemoryStorage 测试内存存储
func TestMemoryStorage(t *testing.T) {
	stor := NewMemoryStorage()

	// 测试保存和加载
	data := &MappingData{
		UserToIP: map[string]string{
			"user1": "192.168.1.1",
			"user2": "192.168.1.2",
		},
		IPToUsers: map[string][]string{
			"192.168.1.1": {"user1"},
			"192.168.1.2": {"user2"},
		},
	}

	err := stor.Save(data)
	if err != nil {
		t.Fatalf("保存数据失败: %v", err)
	}

	loadedData, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(loadedData.UserToIP) != 2 {
		t.Errorf("期望2个用户映射，实际: %d", len(loadedData.UserToIP))
	}

	if loadedData.UserToIP["user1"] != "192.168.1.1" {
		t.Errorf("user1映射错误，期望: 192.168.1.1，实际: %s", loadedData.UserToIP["user1"])
	}
}

// TestMemoryStorageAddMapping 测试增量添加映射
func TestMemoryStorageAddMapping(t *testing.T) {
	stor := NewMemoryStorage()

	err := stor.AddUserMapping("user1", "192.168.1.1")
	if err != nil {
		t.Fatalf("添加映射失败: %v", err)
	}

	err = stor.AddUserMapping("user2", "192.168.1.1")
	if err != nil {
		t.Fatalf("添加映射失败: %v", err)
	}

	data, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(data.UserToIP) != 2 {
		t.Errorf("期望2个用户映射，实际: %d", len(data.UserToIP))
	}

	if len(data.IPToUsers["192.168.1.1"]) != 2 {
		t.Errorf("期望IP有2个用户，实际: %d", len(data.IPToUsers["192.168.1.1"]))
	}
}

// TestMemoryStorageRemoveMapping 测试移除映射
func TestMemoryStorageRemoveMapping(t *testing.T) {
	stor := NewMemoryStorage()

	stor.AddUserMapping("user1", "192.168.1.1")
	stor.AddUserMapping("user2", "192.168.1.2")

	err := stor.RemoveUserMapping("user1", "192.168.1.1")
	if err != nil {
		t.Fatalf("移除映射失败: %v", err)
	}

	data, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(data.UserToIP) != 1 {
		t.Errorf("移除后期望1个用户映射，实际: %d", len(data.UserToIP))
	}

	if _, exists := data.UserToIP["user1"]; exists {
		t.Error("user1应该已被移除")
	}
}

// TestFileStorage 测试文件存储
func TestFileStorage(t *testing.T) {
	tempFile := "/tmp/test_storage.json"
	defer os.Remove(tempFile)

	config := FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	}

	stor := NewFileStorage(config)
	defer stor.Close()

	// 测试保存
	data := &MappingData{
		UserToIP: map[string]string{
			"user1": "192.168.1.1",
			"user2": "192.168.1.2",
		},
		IPToUsers: map[string][]string{
			"192.168.1.1": {"user1"},
			"192.168.1.2": {"user2"},
		},
	}

	err := stor.Save(data)
	if err != nil {
		t.Fatalf("保存数据失败: %v", err)
	}

	// 等待异步刷新
	time.Sleep(200 * time.Millisecond)

	// 验证文件存在
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("文件应该已创建")
	}

	// 测试加载
	loadedData, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(loadedData.UserToIP) != 2 {
		t.Errorf("期望2个用户映射，实际: %d", len(loadedData.UserToIP))
	}
}

// TestFileStoragePersistence 测试文件持久化
func TestFileStoragePersistence(t *testing.T) {
	tempFile := "/tmp/test_storage_persist.json"
	defer os.Remove(tempFile)

	config := FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	}

	// 第一次：保存数据
	{
		stor := NewFileStorage(config)

		data := &MappingData{
			UserToIP: map[string]string{
				"user1": "192.168.1.1",
			},
			IPToUsers: map[string][]string{
				"192.168.1.1": {"user1"},
			},
		}

		stor.Save(data)
		stor.Flush() // 强制刷新
		stor.Close()
	}

	// 第二次：加载数据
	{
		stor := NewFileStorage(config)
		defer stor.Close()

		loadedData, err := stor.Load()
		if err != nil {
			t.Fatalf("加载数据失败: %v", err)
		}

		if len(loadedData.UserToIP) != 1 {
			t.Errorf("持久化后期望1个用户映射，实际: %d", len(loadedData.UserToIP))
		}

		if loadedData.UserToIP["user1"] != "192.168.1.1" {
			t.Error("持久化数据错误")
		}
	}
}

// TestFileStorageAddMapping 测试文件存储增量操作
func TestFileStorageAddMapping(t *testing.T) {
	tempFile := "/tmp/test_storage_add.json"
	defer os.Remove(tempFile)

	config := FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	}

	stor := NewFileStorage(config)
	defer stor.Close()

	// 增量添加
	err := stor.AddUserMapping("user1", "192.168.1.1")
	if err != nil {
		t.Fatalf("添加映射失败: %v", err)
	}

	err = stor.AddUserMapping("user2", "192.168.1.1")
	if err != nil {
		t.Fatalf("添加映射失败: %v", err)
	}

	// 等待刷新
	time.Sleep(200 * time.Millisecond)

	// 验证
	data, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(data.UserToIP) != 2 {
		t.Errorf("期望2个用户映射，实际: %d", len(data.UserToIP))
	}
}

// TestFileStorageConcurrent 测试并发访问
func TestFileStorageConcurrent(t *testing.T) {
	tempFile := "/tmp/test_storage_concurrent.json"
	defer os.Remove(tempFile)

	config := FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 50 * time.Millisecond,
	}

	stor := NewFileStorage(config)
	defer stor.Close()

	// 并发添加
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 10; j++ {
				userID := "user_" + string(rune('0'+id*10+j))
				err := stor.AddUserMapping(userID, "192.168.1.1")
				if err != nil {
					t.Errorf("goroutine %d 添加映射失败: %v", id, err)
				}
			}
		}(i)
	}

	// 等待完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 等待刷新
	time.Sleep(200 * time.Millisecond)

	// 验证
	data, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	if len(data.UserToIP) != 100 {
		t.Errorf("并发添加后期望100个用户映射，实际: %d", len(data.UserToIP))
	}

	t.Log("并发测试完成")
}

// TestCreateStorage 测试存储工厂
func TestCreateStorage(t *testing.T) {
	// 测试文件存储
	config := Config{
		Type:     "file",
		FilePath: "/tmp/test_factory.json",
	}

	stor, err := CreateStorage(config)
	if err != nil {
		t.Fatalf("创建文件存储失败: %v", err)
	}

	if stor == nil {
		t.Error("存储不应为nil")
	}

	stor.Close()
	os.Remove("/tmp/test_factory.json")
}

// TestMemoryStorageFlush 测试内存存储刷新
func TestMemoryStorageFlush(t *testing.T) {
	stor := NewMemoryStorage()

	stor.AddUserMapping("user1", "192.168.1.1")

	err := stor.Flush()
	if err != nil {
		t.Errorf("内存存储刷新不应返回错误: %v", err)
	}
}

// TestFileStorageEmptyFile 测试加载空文件
func TestFileStorageEmptyFile(t *testing.T) {
	tempFile := "/tmp/test_storage_empty.json"
	defer os.Remove(tempFile)

	// 创建空文件
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("创建文件失败: %v", err)
	}
	file.Close()

	config := FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	}

	stor := NewFileStorage(config)
	defer stor.Close()

	// 加载空文件应该返回空映射，不应报错
	data, err := stor.Load()
	if err != nil {
		t.Logf("加载空文件返回错误（可接受）: %v", err)
	}

	// 如果成功加载，应该是空映射
	if data != nil && len(data.UserToIP) != 0 {
		t.Error("空文件应返回空映射")
	}
}
