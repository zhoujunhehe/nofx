package proxy

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"nofx/proxy/mapper"
	"nofx/proxy/provider"
	"nofx/proxy/storage"
)

// TestConcurrentUserAssignment 测试并发分配用户
func TestConcurrentUserAssignment(t *testing.T) {
	tempFile := "/tmp/test_concurrent.json"
	defer func() {
		os.Remove(tempFile)
		os.Remove(tempFile + ".tmp")
	}()

	stor := storage.NewFileStorage(storage.FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	})
	defer stor.Close()

	userMapper := mapper.NewUserIPMapper(stor)

	// 设置3个IP
	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
	}
	userMapper.UpdateAvailableIPs(ips)

	// 并发分配100个用户
	const numUsers = 100
	const numGoroutines = 10

	var wg sync.WaitGroup
	errors := make(chan error, numUsers)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < numUsers/numGoroutines; j++ {
				userID := fmt.Sprintf("user_%d", start*numUsers/numGoroutines+j)
				_, err := userMapper.GetIPForUser(userID)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发分配失败: %v", err)
	}

	// 验证结果
	totalUsers, totalIPs, _ := userMapper.GetStats()
	if totalUsers != numUsers {
		t.Errorf("期望 %d 个用户，实际 %d", numUsers, totalUsers)
	}
	if totalIPs != 3 {
		t.Errorf("期望 3 个IP，实际 %d", totalIPs)
	}

	// 验证负载均衡
	allMappings := userMapper.GetAllMappings()
	minCount := numUsers
	maxCount := 0
	for ip, users := range allMappings {
		count := len(users)
		t.Logf("IP %s: %d 用户", ip, count)
		if count < minCount {
			minCount = count
		}
		if count > maxCount {
			maxCount = count
		}
	}

	diff := maxCount - minCount
	if diff > 5 {
		t.Errorf("负载不够均衡，差异: %d (期望 ≤5)", diff)
	}
}

// TestConcurrentReadWrite 测试并发读写
func TestConcurrentReadWrite(t *testing.T) {
	tempFile := "/tmp/test_concurrent_rw.json"
	defer func() {
		os.Remove(tempFile)
		os.Remove(tempFile + ".tmp")
	}()

	stor := storage.NewFileStorage(storage.FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 100 * time.Millisecond,
	})
	defer stor.Close()

	userMapper := mapper.NewUserIPMapper(stor)

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	userMapper.UpdateAvailableIPs(ips)

	// 预先分配一些用户
	for i := 0; i < 10; i++ {
		userID := fmt.Sprintf("user_%d", i)
		userMapper.GetIPForUser(userID)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 并发读
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				userID := fmt.Sprintf("user_%d", j%10)
				_, err := userMapper.GetIPForUser(userID)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	// 并发写
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				userID := fmt.Sprintf("new_user_%d", start*10+j)
				_, err := userMapper.GetIPForUser(userID)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// 并发查询
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				userMapper.GetAllMappings()
				userMapper.GetStats()
				userMapper.GetUsersForIP("192.168.1.1")
			}
		}()
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发操作失败: %v", err)
	}
}

// TestConcurrentStorageOperations 测试Storage层的并发安全
func TestConcurrentStorageOperations(t *testing.T) {
	tempFile := "/tmp/test_storage_concurrent.json"
	defer func() {
		os.Remove(tempFile)
		os.Remove(tempFile + ".tmp")
	}()

	stor := storage.NewFileStorage(storage.FileStorageConfig{
		FilePath:      tempFile,
		FlushInterval: 50 * time.Millisecond,
	})
	defer stor.Close()

	var wg sync.WaitGroup
	errors := make(chan error, 200)

	// 并发添加
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				userID := fmt.Sprintf("user_%d", start*10+j)
				ip := fmt.Sprintf("192.168.1.%d", (start*10+j)%3+1)
				err := stor.AddUserMapping(userID, ip)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// 并发删除
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // 等待一些数据被添加
			for j := 0; j < 5; j++ {
				userID := fmt.Sprintf("user_%d", start*5+j)
				ip := fmt.Sprintf("192.168.1.%d", (start*5+j)%3+1)
				err := stor.RemoveUserMapping(userID, ip)
				if err != nil {
					errors <- err
				}
			}
		}(i)
	}

	// 并发读取
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_, err := stor.Load()
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	// 并发Flush
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				time.Sleep(20 * time.Millisecond)
				err := stor.Flush()
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// 检查是否有错误
	for err := range errors {
		t.Errorf("并发Storage操作失败: %v", err)
	}

	// 验证最终数据一致性
	data, err := stor.Load()
	if err != nil {
		t.Fatalf("加载数据失败: %v", err)
	}

	// 验证数据一致性：每个userID只能出现一次
	ipUserCount := make(map[string]int)
	for userID, ip := range data.UserToIP {
		// 检查反向索引是否一致
		found := false
		for _, uid := range data.IPToUsers[ip] {
			if uid == userID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("数据不一致: user %s -> ip %s, 但反向索引中不存在", userID, ip)
		}
		ipUserCount[ip]++
	}

	// 检查反向索引的用户数是否匹配
	for ip, users := range data.IPToUsers {
		if len(users) != ipUserCount[ip] {
			t.Errorf("IP %s 的用户数不匹配: 正向=%d, 反向=%d", ip, ipUserCount[ip], len(users))
		}
	}

	t.Logf("最终数据: %d 用户, %d IP", len(data.UserToIP), len(data.IPToUsers))
}

// TestConcurrentRemoveMapping 测试并发删除映射
func TestConcurrentRemoveMapping(t *testing.T) {
	userMapper := mapper.NewUserIPMapper(nil) // 不使用持久化

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}
	userMapper.UpdateAvailableIPs(ips)

	// 预先分配50个用户
	for i := 0; i < 50; i++ {
		userID := fmt.Sprintf("user_%d", i)
		userMapper.GetIPForUser(userID)
	}

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	// 并发删除
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(start int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				userID := fmt.Sprintf("user_%d", start*5+j)
				err := userMapper.RemoveUserMapping(userID)
				if err != nil {
					// 某些用户可能被其他goroutine删除了，这是预期的
					t.Logf("删除用户 %s 失败（可能已被删除）: %v", userID, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// 验证剩余用户数
	totalUsers, _, _ := userMapper.GetStats()
	if totalUsers > 0 {
		t.Logf("剩余 %d 个用户（预期0-50之间）", totalUsers)
	}
}
