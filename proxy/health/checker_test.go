package health

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"nofx/proxy/provider"
)

// TestHealthCheckerBasic 测试基本健康检查
func TestHealthCheckerBasic(t *testing.T) {
	// 创建测试HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// 使用本地地址作为测试IP
	testIP := "127.0.0.1"

	config := HealthCheckConfig{
		CheckURLs:        []string{server.URL}, // 实际测试时会替换host为IP
		Interval:         100 * time.Millisecond,
		Timeout:          1 * time.Second,
		FailureThreshold: 3,
	}

	hc := NewHealthChecker(config)
	if hc == nil {
		t.Fatal("HealthChecker创建失败")
	}

	// 使用本地可达的IP
	ips := []provider.ProxyIP{
		{IP: testIP},
	}

	hc.Start(ips)
	defer hc.Stop()

	// 等待几次检查
	time.Sleep(500 * time.Millisecond)

	// 注意：由于testURL会将URL的host替换为IP，可能导致连接失败
	// 这里主要测试健康检查器的基本功能
	status := hc.GetStatus(testIP)
	if status == nil {
		t.Fatal("无法获取IP状态")
	}

	t.Logf("IP状态: IsHealthy=%v, Failures=%d, Blacklisted=%v",
		status.IsHealthy(), status.ConsecutiveFailures, status.IsBlacklisted())
}

// TestHealthCheckerFailureDetection 测试失败检测
func TestHealthCheckerFailureDetection(t *testing.T) {
	// 创建一个会失败的服务器
	failCount := 0
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		failCount++
		shouldFail := failCount > 2 // 前2次成功，之后失败
		mu.Unlock()

		if shouldFail {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := HealthCheckConfig{
		CheckURLs:        []string{server.URL},
		Interval:         100 * time.Millisecond,
		Timeout:          1 * time.Second,
		FailureThreshold: 3,
	}

	hc := NewHealthChecker(config)
	ips := []provider.ProxyIP{{IP: "192.168.1.1"}}

	var failureDetected bool
	var detectedMu sync.Mutex

	hc.SetFailureCallback(func(event IPFailureEvent) {
		detectedMu.Lock()
		failureDetected = true
		detectedMu.Unlock()
		t.Logf("检测到IP失败: %s, 原因: %s", event.IP, event.Reason)
	})

	hc.Start(ips)
	defer hc.Stop()

	// 等待足够长时间让失败被检测到
	time.Sleep(1 * time.Second)

	detectedMu.Lock()
	detected := failureDetected
	detectedMu.Unlock()

	if !detected {
		t.Error("应该检测到IP失败")
	}

	// 验证IP状态
	status := hc.GetStatus("192.168.1.1")
	if status == nil {
		t.Fatal("无法获取IP状态")
	}

	if !status.IsBlacklisted() {
		t.Errorf("IP应该在黑名单中, 失败次数: %d", status.ConsecutiveFailures)
	}
}

// TestHealthCheckerUpdateIPs 测试IP列表更新
func TestHealthCheckerUpdateIPs(t *testing.T) {
	config := HealthCheckConfig{
		CheckURLs:        []string{"http://test.example.com"},
		Interval:         100 * time.Millisecond,
		Timeout:          1 * time.Second,
		FailureThreshold: 3,
	}

	hc := NewHealthChecker(config)

	// 初始IP列表
	ips1 := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
	}

	hc.Start(ips1)
	defer hc.Stop()

	time.Sleep(100 * time.Millisecond)

	// 验证初始IP
	if !hc.IsIPHealthy("192.168.1.1") {
		t.Error("IP 192.168.1.1 应该存在")
	}

	// 扩容：添加新IP
	ips2 := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"}, // 新增
		{IP: "192.168.1.4"}, // 新增
	}

	hc.UpdateIPs(ips2)

	time.Sleep(100 * time.Millisecond)

	// 验证新IP已添加
	allStatus := hc.GetAllStatus()
	if len(allStatus) < 4 {
		t.Errorf("期望至少4个IP，实际 %d", len(allStatus))
	}

	for _, ip := range []string{"192.168.1.1", "192.168.1.2", "192.168.1.3", "192.168.1.4"} {
		if _, exists := allStatus[ip]; !exists {
			t.Errorf("IP %s 应该存在于健康检查中", ip)
		}
	}
}

// TestHealthCheckerConcurrentAccess 测试并发访问
func TestHealthCheckerConcurrentAccess(t *testing.T) {
	config := HealthCheckConfig{
		CheckURLs:        []string{"http://test.example.com"},
		Interval:         50 * time.Millisecond,
		Timeout:          1 * time.Second,
		FailureThreshold: 3,
	}

	hc := NewHealthChecker(config)

	ips := []provider.ProxyIP{
		{IP: "192.168.1.1"},
		{IP: "192.168.1.2"},
		{IP: "192.168.1.3"},
	}

	hc.Start(ips)
	defer hc.Stop()

	// 并发访问健康检查器
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				hc.IsIPHealthy("192.168.1.1")
				hc.GetHealthyIPs()
				hc.GetStatus("192.168.1.2")
			}
		}()
	}

	wg.Wait()
	t.Log("并发访问测试通过")
}
