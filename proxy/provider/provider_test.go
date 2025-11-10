package provider

import (
	"testing"
)

// TestProxyIPProxyURL 测试代理URL生成
func TestProxyIPProxyURL(t *testing.T) {
	// 保存原始值
	originalHost := Host
	originalUsername := Username
	originalPassword := Password

	// 恢复原始值
	defer func() {
		Host = originalHost
		Username = originalUsername
		Password = originalPassword
	}()

	// 测试不使用全局Host的情况（直接返回IP）
	Host = ""
	Username = ""
	Password = ""

	tests := []struct {
		name     string
		ip       ProxyIP
		expected string
	}{
		{
			name: "HTTP代理",
			ip: ProxyIP{
				IP:       "192.168.1.1",
				Protocol: "http",
			},
			expected: "192.168.1.1",  // 没有Host时直接返回IP
		},
		{
			name: "直接URL",
			ip: ProxyIP{
				IP:       "http://proxy.example.com:8080",
				Protocol: "http",
			},
			expected: "http://proxy.example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := tt.ip.ProxyURL()
			if url != tt.expected {
				t.Errorf("期望: %s, 实际: %s", tt.expected, url)
			}
		})
	}

	// 测试使用Host和Username的情况（含%s占位符）
	Host = "proxy.example.com"
	Username = "user-%s"  // 使用%s占位符
	Password = "pass123"

	ip := ProxyIP{
		IP:       "192.168.1.1",
		Protocol: "http",
	}

	url := ip.ProxyURL()
	expected := "http://user-192.168.1.1:pass123@proxy.example.com"
	if url != expected {
		t.Errorf("使用占位符时，期望: %s, 实际: %s", expected, url)
	}
}

// TestSingleProvider 测试单代理提供者
func TestSingleProvider(t *testing.T) {
	proxyURL := "http://127.0.0.1:7890"

	provider := NewSingleProxyProvider(proxyURL)
	if provider == nil {
		t.Fatal("provider不应为nil")
	}

	// 获取IP列表
	ips, err := provider.GetIPList()
	if err != nil {
		t.Fatalf("获取IP列表失败: %v", err)
	}

	if len(ips) != 1 {
		t.Errorf("单代理模式应返回1个IP，实际: %d", len(ips))
	}

	if ips[0].IP != proxyURL {
		t.Errorf("单代理IP应为'%s'，实际: %s", proxyURL, ips[0].IP)
	}

	// 刷新IP列表
	ips2, err := provider.RefreshIPList()
	if err != nil {
		t.Fatalf("刷新IP列表失败: %v", err)
	}

	if len(ips2) != 1 {
		t.Errorf("刷新后应返回1个IP，实际: %d", len(ips2))
	}
}

// TestPoolProvider 测试代理池提供者
func TestPoolProvider(t *testing.T) {
	proxyList := []string{
		"http://127.0.0.1:7890",
		"http://127.0.0.1:7891",
		"http://127.0.0.1:7892",
	}

	provider := NewPoolProvider(proxyList)
	if provider == nil {
		t.Fatal("provider不应为nil")
	}

	// 获取IP列表
	ips, err := provider.GetIPList()
	if err != nil {
		t.Fatalf("获取IP列表失败: %v", err)
	}

	if len(ips) != 3 {
		t.Errorf("代理池应返回3个IP，实际: %d", len(ips))
	}

	// 验证每个IP的格式
	for i, ip := range ips {
		t.Logf("IP %d: %s, ProxyURL: %s", i, ip.IP, ip.ProxyURL())

		if ip.IP == "" {
			t.Errorf("IP %d 不应为空", i)
		}

		if ip.Protocol != "http" {
			t.Errorf("IP %d 协议应为http，实际: %s", i, ip.Protocol)
		}
	}

	// 刷新IP列表
	ips2, err := provider.RefreshIPList()
	if err != nil {
		t.Fatalf("刷新IP列表失败: %v", err)
	}

	if len(ips2) != 3 {
		t.Errorf("刷新后应返回3个IP，实际: %d", len(ips2))
	}
}

// TestCreateProvider 测试provider工厂
func TestCreateProvider(t *testing.T) {
	// 测试single模式
	provider, err := CreateProvider("single", "http://127.0.0.1:7890", nil, "", "", "")
	if err != nil {
		t.Fatalf("创建single provider失败: %v", err)
	}

	if provider == nil {
		t.Error("provider不应为nil")
	}

	ips, err := provider.GetIPList()
	if err != nil {
		t.Fatalf("获取IP列表失败: %v", err)
	}

	if len(ips) != 1 {
		t.Errorf("single模式应返回1个IP，实际: %d", len(ips))
	}

	// 测试pool模式
	proxyList := []string{
		"http://127.0.0.1:7890",
		"http://127.0.0.1:7891",
	}

	provider2, err := CreateProvider("pool", "", proxyList, "", "", "")
	if err != nil {
		t.Fatalf("创建pool provider失败: %v", err)
	}

	ips2, err := provider2.GetIPList()
	if err != nil {
		t.Fatalf("获取IP列表失败: %v", err)
	}

	if len(ips2) != 2 {
		t.Errorf("pool模式应返回2个IP，实际: %d", len(ips2))
	}

	// 测试未知模式
	_, err = CreateProvider("unknown", "", nil, "", "", "")
	if err == nil {
		t.Error("未知模式应返回错误")
	}
}

// TestProxyIPDefaultProtocol 测试默认协议
func TestProxyIPDefaultProtocol(t *testing.T) {
	// 清空全局变量
	Host = ""
	Username = ""
	Password = ""

	ip := ProxyIP{
		IP:       "192.168.1.1",
		Protocol: "", // 未设置协议
	}

	url := ip.ProxyURL()
	t.Logf("默认协议URL: %s", url)

	// 应该使用默认协议（取决于实现）
	if url == "" {
		t.Error("ProxyURL不应为空")
	}
}

// TestPoolProviderEmpty 测试空代理池
func TestPoolProviderEmpty(t *testing.T) {
	proxyList := []string{}

	provider := NewPoolProvider(proxyList)
	if provider == nil {
		t.Fatal("provider不应为nil")
	}

	ips, err := provider.GetIPList()
	if err != nil {
		t.Fatalf("获取IP列表失败: %v", err)
	}

	if len(ips) != 0 {
		t.Errorf("空代理池应返回0个IP，实际: %d", len(ips))
	}
}

// TestProviderGlobalVariables 测试全局变量设置
func TestProviderGlobalVariables(t *testing.T) {
	// 保存原始值
	originalHost := Host
	originalUsername := Username
	originalPassword := Password

	// 恢复原始值
	defer func() {
		Host = originalHost
		Username = originalUsername
		Password = originalPassword
	}()

	// 测试设置
	Host = "test.proxy.com"
	Username = "testuser"
	Password = "testpass"

	ip := ProxyIP{
		IP:       "192.168.1.1",
		Protocol: "http",
	}

	url := ip.ProxyURL()

	if Host != "test.proxy.com" {
		t.Errorf("Host设置错误: %s", Host)
	}

	if Username != "testuser" {
		t.Errorf("Username设置错误: %s", Username)
	}

	if Password != "testpass" {
		t.Errorf("Password设置错误: %s", Password)
	}

	t.Logf("生成的URL: %s", url)
}
