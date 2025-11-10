package health

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"nofx/proxy/provider"
	"sync"
	"time"
)

// ============ 健康检查器 ============

// HealthChecker 代理IP健康检查器
//
// 职责：
// 1. 定期检查代理IP是否可用
// 2. 维护IP健康状态
// 3. 触发IP失败/恢复事件
type HealthChecker struct {
	config HealthCheckConfig

	// IP健康状态表（ip -> health）
	healthMap map[string]*IPHealth
	healthMu  sync.RWMutex

	// 代理配置表（ip -> proxyIP）
	proxyMap map[string]provider.ProxyIP
	proxyMu  sync.RWMutex

	// 事件回调
	onIPFailure  func(IPFailureEvent)
	onIPRecovery func(IPRecoveryEvent)

	// 获取IP绑定用户的函数
	getUsersForIP func(string) []string

	// 控制
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config HealthCheckConfig) *HealthChecker {

	config = config.WithDefaults()

	return &HealthChecker{
		config:    config,
		healthMap: make(map[string]*IPHealth),
		proxyMap:  make(map[string]provider.ProxyIP),
		stopChan:  make(chan struct{}),
	}
}

// ============ 配置方法 ============

// SetFailureCallback 设置IP失败回调
func (hc *HealthChecker) SetFailureCallback(fn func(IPFailureEvent)) {
	hc.onIPFailure = fn
}

// SetRecoveryCallback 设置IP恢复回调
func (hc *HealthChecker) SetRecoveryCallback(fn func(IPRecoveryEvent)) {
	hc.onIPRecovery = fn
}

// SetGetUsersFunc 设置获取IP用户列表的函数
func (hc *HealthChecker) SetGetUsersFunc(fn func(string) []string) {
	hc.getUsersForIP = fn
}

// ============ 生命周期管理 ============

// Start 启动健康检查
func (hc *HealthChecker) Start(ips []provider.ProxyIP) {
	// 初始化IP状态
	hc.UpdateIPs(ips)

	// 启动后台检查
	hc.wg.Add(1)
	go hc.checkLoop()

	log.Printf("✓ 健康检查已启动: %d个代理IP, 间隔=%v, 失败阈值=%d次",
		len(ips), hc.config.Interval, hc.config.FailureThreshold)
}

// Stop 停止健康检查
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
	log.Printf("✓ 健康检查已停止")
}

// UpdateIPs 更新IP列表（扩容/缩容时调用）
func (hc *HealthChecker) UpdateIPs(ips []provider.ProxyIP) {
	hc.lockBothMaps()
	defer hc.unlockBothMaps()

	// 构建新IP集合
	newIPSet := make(map[string]bool)
	for _, ip := range ips {
		newIPSet[ip.IP] = true

		// 添加新IP
		if _, exists := hc.healthMap[ip.IP]; !exists {
			hc.addNewProxyIP(ip)
		}
	}

	// 清理已移除的IP
	removedCount := 0
	for ip := range hc.healthMap {
		if !newIPSet[ip] {
			delete(hc.healthMap, ip)
			delete(hc.proxyMap, ip)
			removedCount++
		}
	}

	if removedCount > 0 {
		log.Printf("✓ 健康检查IP列表已更新: %d个IP, 移除 %d个IP", len(ips), removedCount)
	}
}

// ============ 查询方法 ============

// IsIPHealthy 检查IP是否健康
func (hc *HealthChecker) IsIPHealthy(ip string) bool {
	hc.healthMu.RLock()
	defer hc.healthMu.RUnlock()

	health, exists := hc.healthMap[ip]
	if !exists {
		// 未知IP不应该被使用（启用了健康检查后，所有IP都应该在监控列表中）
		log.Printf("⚠️  检测到未知IP: %s（不在健康检查列表中）", ip)
		return false
	}

	return health.CanAssignToUser()
}

// GetHealthyIPs 获取所有健康的IP列表
func (hc *HealthChecker) GetHealthyIPs() []string {
	hc.healthMu.RLock()
	defer hc.healthMu.RUnlock()

	healthyIPs := make([]string, 0)
	for ip, health := range hc.healthMap {
		if health.CanAssignToUser() {
			healthyIPs = append(healthyIPs, ip)
		}
	}

	return healthyIPs
}

// GetAllIPs 获取所有IP列表（无论健康状态）
func (hc *HealthChecker) GetAllIPs() []string {
	hc.healthMu.RLock()
	defer hc.healthMu.RUnlock()

	allIPs := make([]string, 0, len(hc.healthMap))
	for ip := range hc.healthMap {
		allIPs = append(allIPs, ip)
	}

	return allIPs
}

// GetStatus 获取IP健康状态（返回副本）
func (hc *HealthChecker) GetStatus(ip string) *IPHealth {
	hc.healthMu.RLock()
	defer hc.healthMu.RUnlock()

	health, exists := hc.healthMap[ip]
	if !exists {
		return nil
	}

	return health.Copy()
}

// GetAllStatus 获取所有IP状态
func (hc *HealthChecker) GetAllStatus() map[string]*IPHealth {
	hc.healthMu.RLock()
	defer hc.healthMu.RUnlock()

	result := make(map[string]*IPHealth)
	for ip, health := range hc.healthMap {
		result[ip] = health.Copy()
	}
	return result
}

// ============ 黑名单管理 ============

// AddBlacklist 手动将IP加入黑名单（例如外部请求失败时调用）
func (hc *HealthChecker) AddBlacklist(ip string, reason string) {
	hc.healthMu.Lock()
	defer hc.healthMu.Unlock()

	health, exists := hc.healthMap[ip]
	if !exists {
		log.Printf("⚠️  无法加入黑名单: IP %s 不在健康检查列表中", ip)
		return
	}

	// 如果已经在黑名单中，只更新失败次数
	if health.Status == StatusBlacklisted {
		health.ConsecutiveFailures++
		log.Printf("⚠️  IP已在黑名单中: %s (失败次数: %d)", ip, health.ConsecutiveFailures)
		return
	}

	// 加入黑名单
	health.Status = StatusBlacklisted
	health.ConsecutiveFailures = hc.config.FailureThreshold
	health.LastError = reason
	health.LastCheckTime = time.Now()

	log.Printf("❌ IP已手动加入黑名单: %s, 原因: %s", ip, reason)

	// 异步触发失败回调，让mapper重新分配用户
	if hc.onIPFailure != nil {
		go hc.safeExecuteFailureCallback(ip, reason)
	}
}

// RemoveFromBlacklist 从黑名单移除IP（恢复健康状态）
func (hc *HealthChecker) RemoveFromBlacklist(ip string) {
	hc.healthMu.Lock()
	defer hc.healthMu.Unlock()

	health, exists := hc.healthMap[ip]
	if !exists {
		return
	}

	if health.Status == StatusBlacklisted {
		health.Status = StatusHealthy
		health.ConsecutiveFailures = 0
		health.LastError = ""
		log.Printf("✓ IP已从黑名单移除: %s", ip)
	}
}

// ============ 内部实现 - 主循环 ============

// checkLoop 健康检查主循环
func (hc *HealthChecker) checkLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllProxies()
		case <-hc.stopChan:
			return
		}
	}
}

// checkAllProxies 检查所有代理IP
func (hc *HealthChecker) checkAllProxies() {
	// 获取所有需要检查的IP
	hc.proxyMu.RLock()
	proxies := make([]provider.ProxyIP, 0, len(hc.proxyMap))
	for _, proxy := range hc.proxyMap {
		proxies = append(proxies, proxy)
	}
	hc.proxyMu.RUnlock()

	// 顺序检查每个代理
	for _, proxy := range proxies {
		hc.checkSingleProxyHealth(proxy)
	}
}

// checkSingleProxyHealth 检查单个代理IP健康状态
func (hc *HealthChecker) checkSingleProxyHealth(proxy provider.ProxyIP) {
	// 对每个检查URL进行测试
	var lastError error
	allSuccess := true

	for _, checkURL := range hc.config.CheckURLs {
		if err := hc.testProxyConnection(proxy, checkURL); err != nil {
			allSuccess = false
			lastError = err
			break // 任意一个URL失败即认为代理不可用
		}
	}

	// 更新状态
	hc.updateHealthStatus(proxy.IP, allSuccess, lastError)
}

// testProxyConnection 测试代理连接（通过代理访问URL）
func (hc *HealthChecker) testProxyConnection(proxy provider.ProxyIP, targetURL string) error {
	parsedProxyURL, err := url.Parse(proxy.ProxyURL())
	if err != nil {
		return fmt.Errorf("解析代理URL失败: %w", err)
	}

	// 2. 创建使用代理的HTTP客户端
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsedProxyURL),
		},
		Timeout:   hc.config.Timeout,
	}

	// 3. 创建请求
	ctx, cancel := context.WithTimeout(context.Background(), hc.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	log.Printf("🔍 检查代理 %s 访问 %s", proxy.IP, targetURL)

	// 4. 通过代理发送请求
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("代理请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 5. 检查响应状态
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return fmt.Errorf("HTTP状态码异常: %d", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		log.Printf("⚠️  代理 %s 返回服务器错误状态码: %d", proxy.IP, resp.StatusCode)
		return nil
	}

	return nil
}

// ============ 内部实现 - 状态管理 ============

// updateHealthStatus 更新IP健康状态
func (hc *HealthChecker) updateHealthStatus(ip string, success bool, lastError error) {
	hc.healthMu.Lock()
	defer hc.healthMu.Unlock()

	health, exists := hc.healthMap[ip]
	if !exists {
		return
	}

	health.LastCheckTime = time.Now()

	if success {
		hc.handleCheckSuccess(health)
	} else {
		hc.handleCheckFailure(health, lastError)
	}
}

// handleCheckSuccess 处理检查成功
func (hc *HealthChecker) handleCheckSuccess(health *IPHealth) {
	previousStatus := health.Status

	// 恢复健康
	health.Status = StatusHealthy
	health.ConsecutiveFailures = 0
	health.LastError = ""

	// 根据之前的状态，输出不同的日志
	switch previousStatus {
	case StatusBlacklisted:
		// 从黑名单恢复
		log.Printf("✅ IP已从黑名单恢复（重新加入健康名单）: %s", health.IP)
		if hc.onIPRecovery != nil {
			go hc.safeExecuteRecoveryCallback(health.IP)
		}

	case StatusTemporaryFailed:
		// 从临时失败恢复
		log.Printf("✅ IP已从临时失败恢复: %s", health.IP)
		if hc.onIPRecovery != nil {
			go hc.safeExecuteRecoveryCallback(health.IP)
		}

	case StatusHealthy:
		// 持续健康，无需特殊处理
	}
}

// handleCheckFailure 处理检查失败
func (hc *HealthChecker) handleCheckFailure(health *IPHealth, err error) {
	health.ConsecutiveFailures++
	if err != nil {
		health.LastError = err.Error()
	}

	// 如果已经在黑名单中，继续保持黑名单状态，不重复触发事件
	if health.Status == StatusBlacklisted {
		log.Printf("⚠️  黑名单IP持续失败: %s (失败次数: %d)", health.IP, health.ConsecutiveFailures)
		return
	}

	// 根据失败次数转换状态
	if health.ConsecutiveFailures == 1 {
		// 首次失败：标记为临时失败
		hc.transitionToTemporaryFailed(health, err)
	} else if health.ConsecutiveFailures >= hc.config.FailureThreshold {
		// 达到阈值：加入黑名单
		hc.transitionToBlacklisted(health)
	} else {
		// 持续失败但未达阈值
		log.Printf("⚠️  代理持续失败（%d/%d）: %s, 原因: %v",
			health.ConsecutiveFailures, hc.config.FailureThreshold, health.IP, err)
	}
}

// transitionToTemporaryFailed 转换为临时失败状态
func (hc *HealthChecker) transitionToTemporaryFailed(health *IPHealth, err error) {
	health.Status = StatusTemporaryFailed
	log.Printf("⚠️  代理首次失败（临时）: %s, 原因: %v", health.IP, err)
}

// transitionToBlacklisted 转换为黑名单状态
func (hc *HealthChecker) transitionToBlacklisted(health *IPHealth) {
	// 只在首次进入黑名单时触发事件
	if health.Status != StatusBlacklisted {
		health.Status = StatusBlacklisted
		log.Printf("❌ 代理已加入黑名单（连续失败%d次）: %s",
			health.ConsecutiveFailures, health.IP)

		// 触发失败事件
		if hc.onIPFailure != nil {
			reason := fmt.Sprintf("连续失败%d次", health.ConsecutiveFailures)
			go hc.safeExecuteFailureCallback(health.IP, reason)
		}
	}
}


// ============ 内部实现 - 事件触发 ============

// safeExecuteFailureCallback 安全执行失败回调（带panic保护）
func (hc *HealthChecker) safeExecuteFailureCallback(ip, reason string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️  失败回调panic: %v", r)
		}
	}()

	// 获取受影响的用户
	var affectedUsers []string
	if hc.getUsersForIP != nil {
		affectedUsers = hc.getUsersForIP(ip)
	}

	event := IPFailureEvent{
		IP:            ip,
		AffectedUsers: affectedUsers,
		Reason:        reason,
		OccurredAt:    time.Now(),
	}

	hc.onIPFailure(event)
}

// safeExecuteRecoveryCallback 安全执行恢复回调（带panic保护）
func (hc *HealthChecker) safeExecuteRecoveryCallback(ip string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️  恢复回调panic: %v", r)
		}
	}()

	event := IPRecoveryEvent{
		IP:         ip,
		OccurredAt: time.Now(),
	}

	hc.onIPRecovery(event)
}

// ============ 内部实现 - 辅助方法 ============

// addNewProxyIP 添加新的代理IP（需要持锁调用）
func (hc *HealthChecker) addNewProxyIP(proxy provider.ProxyIP) {
	hc.healthMap[proxy.IP] = &IPHealth{
		IP:                  proxy.IP,
		Status:              StatusHealthy,
		ConsecutiveFailures: 0,
		LastCheckTime:       time.Now(),
	}
	hc.proxyMap[proxy.IP] = proxy
	log.Printf("✓ 新增IP到健康检查: %s", proxy.IP)
}

// lockBothMaps 按固定顺序锁定两个map（避免死锁）
func (hc *HealthChecker) lockBothMaps() {
	hc.proxyMu.Lock()
	hc.healthMu.Lock()
}

// unlockBothMaps 按相反顺序解锁
func (hc *HealthChecker) unlockBothMaps() {
	hc.healthMu.Unlock()
	hc.proxyMu.Unlock()
}
