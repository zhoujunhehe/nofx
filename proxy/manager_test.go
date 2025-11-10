package proxy

import (
	"testing"
	"time"
)

// TestNewProxyManagerDisabled 测试禁用代理时的管理器创建
func TestNewProxyManagerDisabled(t *testing.T) {
	config := &Config{
		Enabled: false,
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建禁用的代理管理器失败: %v", err)
	}

	if manager == nil {
		t.Fatal("管理器不应为nil")
	}

	// 测试获取直连客户端
	client, err := manager.GetProxyClient()
	if err != nil {
		t.Fatalf("获取直连客户端失败: %v", err)
	}

	if client.ProxyID != -1 {
		t.Errorf("禁用代理时ProxyID应为-1，实际: %d", client.ProxyID)
	}

	if client.IP != "direct" {
		t.Errorf("禁用代理时IP应为'direct'，实际: %s", client.IP)
	}
}

// TestNewProxyManagerSingleMode 测试单代理模式
func TestNewProxyManagerSingleMode(t *testing.T) {
	config := &Config{
		Enabled:  true,
		Mode:     "single",
		ProxyURL: "http://127.0.0.1:7890",
		Timeout:  30 * time.Second,
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建单代理管理器失败: %v", err)
	}

	if manager.provider == nil {
		t.Fatal("provider不应为nil")
	}

	// 测试获取代理客户端
	client, err := manager.GetProxyClient()
	if err != nil {
		t.Fatalf("获取代理客户端失败: %v", err)
	}

	if client.ProxyID == -1 {
		t.Error("单代理模式ProxyID不应为-1")
	}
}

// TestNewProxyManagerPoolMode 测试代理池模式
func TestNewProxyManagerPoolMode(t *testing.T) {
	config := &Config{
		Enabled: true,
		Mode:    "pool",
		ProxyList: []string{
			"http://127.0.0.1:7890",
			"http://127.0.0.1:7891",
			"http://127.0.0.1:7892",
		},
		Timeout: 30 * time.Second,
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建代理池管理器失败: %v", err)
	}

	if len(manager.ipList) != 3 {
		t.Errorf("期望3个代理IP，实际: %d", len(manager.ipList))
	}
}

// TestGetProxyClientForUser 测试用户绑定代理
func TestGetProxyClientForUser(t *testing.T) {
	config := &Config{
		Enabled:       true,
		Mode:          "pool",
		ProxyList:     []string{"http://127.0.0.1:7890", "http://127.0.0.1:7891"},
		Timeout:       30 * time.Second,
		ProxyHost:     "127.0.0.1:8080",  // 设置代理主机
		ProxyUser:     "user-%s",          // 使用%s占位符
		ProxyPassword: "testpass",         // 设置密码
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 测试同一用户多次获取应返回相同IP
	client1, err := manager.GetProxyClientForUser("user1")
	if err != nil {
		t.Fatalf("获取用户代理失败: %v", err)
	}

	client2, err := manager.GetProxyClientForUser("user1")
	if err != nil {
		t.Fatalf("再次获取用户代理失败: %v", err)
	}

	if client1.IP != client2.IP {
		t.Errorf("同一用户应获得相同IP，实际: %s != %s", client1.IP, client2.IP)
	}

	// 测试不同用户可能获得不同IP
	client3, err := manager.GetProxyClientForUser("user2")
	if err != nil {
		t.Fatalf("获取用户2代理失败: %v", err)
	}

	t.Logf("user1 IP: %s, user2 IP: %s", client1.IP, client3.IP)
}

// TestGetBlacklistStatus 测试黑名单状态获取
func TestGetBlacklistStatus(t *testing.T) {
	config := &Config{
		Enabled: true,
		Mode:    "pool",
		ProxyList: []string{
			"http://127.0.0.1:7890",
			"http://127.0.0.1:7891",
			"http://127.0.0.1:7892",
		},
		Timeout: 30 * time.Second,
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	total, blacklisted, available := manager.GetBlacklistStatus()

	if total != 3 {
		t.Errorf("期望3个总IP，实际: %d", total)
	}

	if available != 3 {
		t.Errorf("期望3个可用IP，实际: %d", available)
	}

	if blacklisted != 0 {
		t.Errorf("期望0个黑名单IP，实际: %d", blacklisted)
	}
}

// TestStopAutoRefresh 测试停止自动刷新（防止重复关闭channel panic）
func TestStopAutoRefresh(t *testing.T) {
	config := &Config{
		Enabled:         true,
		Mode:            "pool",
		ProxyList:       []string{"http://127.0.0.1:7890"},
		Timeout:         30 * time.Second,
		RefreshInterval: 1 * time.Hour, // 设置较长时间，避免自动刷新
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 多次调用StopAutoRefresh不应panic
	manager.StopAutoRefresh()
	manager.StopAutoRefresh()
	manager.StopAutoRefresh()

	t.Log("多次停止刷新成功，无panic")
}

// TestInitGlobalProxyManagerSuccess 测试全局管理器初始化成功
func TestInitGlobalProxyManagerSuccess(t *testing.T) {
	// 注意：这个测试会修改全局状态，需要小心
	config := &Config{
		Enabled: false, // 使用禁用模式避免依赖外部服务
	}

	err := InitGlobalProxyManager(config)
	if err != nil {
		t.Fatalf("初始化全局管理器失败: %v", err)
	}

	manager := GetGlobalProxyManager()
	if manager == nil {
		t.Fatal("全局管理器不应为nil")
	}
}

// TestGetProxyClientConcurrent 测试并发获取代理客户端
func TestGetProxyClientConcurrent(t *testing.T) {
	config := &Config{
		Enabled:       true,
		Mode:          "pool",
		ProxyList:     []string{"http://127.0.0.1:7890", "http://127.0.0.1:7891", "http://127.0.0.1:7892"},
		Timeout:       30 * time.Second,
		ProxyHost:     "127.0.0.1:8080",
		ProxyUser:     "user-%s",  // 使用%s占位符
		ProxyPassword: "testpass",
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 并发测试
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < 100; j++ {
				_, err := manager.GetProxyClient()
				if err != nil {
					t.Errorf("goroutine %d 获取代理失败: %v", id, err)
					return
				}
			}
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("并发测试完成")
}

// TestRefreshIPList 测试刷新IP列表
func TestRefreshIPList(t *testing.T) {
	config := &Config{
		Enabled: true,
		Mode:    "pool",
		ProxyList: []string{
			"http://127.0.0.1:7890",
		},
		Timeout: 30 * time.Second,
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 刷新IP列表
	err = manager.RefreshIPList()
	if err != nil {
		t.Fatalf("刷新IP列表失败: %v", err)
	}

	if len(manager.ipList) == 0 {
		t.Error("刷新后IP列表不应为空")
	}
}

// TestMappingPersistence 测试映射持久化
func TestMappingPersistence(t *testing.T) {
	config := &Config{
		Enabled:       true,
		Mode:          "pool",
		ProxyList:     []string{"http://127.0.0.1:7890", "http://127.0.0.1:7891"},
		Timeout:       30 * time.Second,
		MappingFile:   "/tmp/test_mapping_persistence.json",
		ProxyHost:     "127.0.0.1:8080",
		ProxyUser:     "user-%s",  // 使用%s占位符
		ProxyPassword: "testpass",
	}

	manager, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建管理器失败: %v", err)
	}

	// 为用户分配IP
	client, err := manager.GetProxyClientForUser("test_user")
	if err != nil {
		t.Fatalf("获取用户代理失败: %v", err)
	}

	originalIP := client.IP

	// 强制保存
	err = manager.SaveMappings()
	if err != nil {
		t.Fatalf("保存映射失败: %v", err)
	}

	// 创建新管理器，应该能加载已有映射
	manager2, err := NewProxyManager(config)
	if err != nil {
		t.Fatalf("创建第二个管理器失败: %v", err)
	}

	client2, err := manager2.GetProxyClientForUser("test_user")
	if err != nil {
		t.Fatalf("从第二个管理器获取用户代理失败: %v", err)
	}

	if client2.IP != originalIP {
		t.Errorf("持久化后IP应保持不变，期望: %s，实际: %s", originalIP, client2.IP)
	}

	t.Logf("持久化测试通过，IP保持: %s", originalIP)
}
