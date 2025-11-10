package mapper

import (
	"fmt"
	"log"
	"nofx/proxy/health"
	"nofx/proxy/provider"
	"nofx/proxy/storage"
	"sync"
)

// ============ 用户IP映射器 ============

// UserIPMapper 用户-IP映射管理器
//
// 职责：
// 1. 管理用户和代理IP的绑定关系
// 2. 负载均衡分配IP
// 3. IP失败时自动重新分配
// 4. 持久化映射关系
type UserIPMapper struct {
	// userID -> IP 映射
	userToIP map[string]string

	// IP -> []userID 映射（反向索引）
	ipToUsers map[string][]string

	// IP -> 用户数量（快速查询）
	ipUserCount map[string]int

	// 存储接口（支持File/Redis/DB）
	storage storage.Storage

	// 健康检查器（负责管理所有IP及其健康状态）
	healthChecker *health.HealthChecker

	// IP重新分配通知回调
	onIPReassignCallback func(userID, oldIP, newIP string)

	mutex sync.RWMutex
}

// NewUserIPMapper 创建用户IP映射器
func NewUserIPMapper(storage storage.Storage) *UserIPMapper {
	mapper := &UserIPMapper{
		userToIP:    make(map[string]string),
		ipToUsers:   make(map[string][]string),
		ipUserCount: make(map[string]int),
		storage:     storage,
	}

	// 从存储加载已有映射
	if storage != nil {
		if err := mapper.loadMappingsFromStorage(); err != nil {
			log.Printf("⚠️  加载映射失败（将创建新映射）: %v", err)
		}
	}

	return mapper
}

// ============ 公共API ============

// UpdateAvailableIPs 更新可用IP列表（扩容/缩容时调用）
func (m *UserIPMapper) UpdateAvailableIPs(ips []provider.ProxyIP) {
	// 更新健康检查器的IP列表（health负责管理所有IP）
	if m.healthChecker != nil {
		// 如果healthChecker还没启动，先启动
		if len(m.healthChecker.GetAllIPs()) == 0 && len(ips) > 0 {
			m.healthChecker.Start(ips)
			log.Printf("✓ 健康检查器已启动，管理 %d 个IP", len(ips))
		} else {
			m.healthChecker.UpdateIPs(ips)
		}
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 构建新IP集合用于检查
	newIPSet := make(map[string]bool, len(ips))
	for _, ip := range ips {
		newIPSet[ip.IP] = true

		// 初始化反向索引（如果不存在）
		if _, exists := m.ipToUsers[ip.IP]; !exists {
			m.ipToUsers[ip.IP] = make([]string, 0)
			m.ipUserCount[ip.IP] = 0
		}
	}

	// 清理已移除的IP数据（防止内存泄漏）
	removedCount := 0
	for ip := range m.ipToUsers {
		if !newIPSet[ip] {
			// IP已被移除，清理该IP上的用户绑定
			if users := m.ipToUsers[ip]; len(users) > 0 {
				log.Printf("⚠️  IP %s 已移除，清理 %d 个用户的绑定（下次请求时重新分配）", ip, len(users))

				// 清除用户的IP绑定
				for _, userID := range users {
					delete(m.userToIP, userID)
					// 异步从存储中删除
					m.asyncRemoveFromStorage(userID, ip)
				}
			}
			delete(m.ipToUsers, ip)
			delete(m.ipUserCount, ip)
			removedCount++
		}
	}

	if removedCount > 0 {
		log.Printf("✓ Mapper已清理 %d 个已移除IP的用户绑定", removedCount)
	}
}

// GetIPForUser 获取用户对应的IP（不存在则分配）
func (m *UserIPMapper) GetIPForUser(userID string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果用户已有IP映射
	if ip, exists := m.userToIP[userID]; exists {
		// 检查IP是否健康（healthChecker负责管理所有IP）
		if m.healthChecker != nil && !m.healthChecker.IsIPHealthy(ip) {
			// IP不健康或不存在，重新分配
			log.Printf("⚠️  用户 %s 的原IP %s 不健康或不可用，重新分配", userID, ip)
			m.unbindUserFromIP(userID, ip)
			m.asyncRemoveFromStorage(userID, ip)
		} else {
			// IP健康，直接返回
			return ip, nil
		}
	}

	// 为用户分配新IP
	ip, err := m.allocateIPForUser(userID, nil)
	if err != nil {
		return "", err
	}

	log.Printf("✓ 用户 %s 分配IP: %s", userID, ip)
	return ip, nil
}

// GetUsersForIP 获取绑定到指定IP的所有用户
func (m *UserIPMapper) GetUsersForIP(ip string) []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if users, exists := m.ipToUsers[ip]; exists {
		// 返回副本，避免外部修改
		result := make([]string, len(users))
		copy(result, users)
		return result
	}

	return []string{}
}

// GetAllMappings 获取所有映射关系（调试用）
func (m *UserIPMapper) GetAllMappings() map[string][]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 返回副本
	result := make(map[string][]string)
	for ip, users := range m.ipToUsers {
		usersCopy := make([]string, len(users))
		copy(usersCopy, users)
		result[ip] = usersCopy
	}

	return result
}

// GetStats 获取统计信息
func (m *UserIPMapper) GetStats() (totalUsers int, totalIPs int, avgUsersPerIP float64) {
	m.mutex.RLock()
	totalUsers = len(m.userToIP)

	// 从healthChecker获取总IP数，如果未启用则从ipToUsers获取
	if m.healthChecker != nil {
		totalIPs = len(m.healthChecker.GetAllIPs())
	} else {
		totalIPs = len(m.ipToUsers)
	}
	m.mutex.RUnlock()

	if totalIPs > 0 {
		avgUsersPerIP = float64(totalUsers) / float64(totalIPs)
	}

	return
}

// RemoveUserMapping 移除用户映射
func (m *UserIPMapper) RemoveUserMapping(userID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	ip, exists := m.userToIP[userID]
	if !exists {
		return fmt.Errorf("用户 %s 不存在映射", userID)
	}

	m.unbindUserFromIP(userID, ip)
	m.asyncRemoveFromStorage(userID, ip)

	log.Printf("✓ 移除用户映射: %s -> %s", userID, ip)
	return nil
}

// Flush 强制刷新到存储
func (m *UserIPMapper) Flush() error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.storage == nil {
		return nil
	}

	return m.storage.Flush()
}

// Close 关闭映射器（清理资源）
func (m *UserIPMapper) Close() error {
	// 停止健康检查
	if m.healthChecker != nil {
		m.healthChecker.Stop()
	}

	// 关闭存储
	if m.storage != nil {
		return m.storage.Close()
	}
	return nil
}

// ============ 健康检查相关 ============

// IsIPHealthy 检查IP是否健康（如果启用了健康检查）
func (m *UserIPMapper) IsIPHealthy(ip string) bool {
	if m.healthChecker == nil {
		return true // 未启用健康检查，默认都健康
	}
	return m.healthChecker.IsIPHealthy(ip)
}

// AddBlacklist 手动将IP加入黑名单
func (m *UserIPMapper) AddBlacklist(ip string, reason string) {
	if m.healthChecker == nil {
		return
	}
	m.healthChecker.AddBlacklist(ip, reason)
}

// EnableHealthCheck 启用健康检查
func (m *UserIPMapper) EnableHealthCheck(config health.HealthCheckConfig) {
	m.healthChecker = health.NewHealthChecker(config)
	if m.healthChecker == nil {
		return
	}

	// 设置获取IP绑定用户的函数
	m.healthChecker.SetGetUsersFunc(func(ip string) []string {
		return m.GetUsersForIP(ip)
	})

	// 设置IP失败回调
	m.healthChecker.SetFailureCallback(func(event health.IPFailureEvent) {
		m.OnIPFailedAutoReassign(event)
	})

	// 注意：healthChecker.Start() 将在 UpdateAvailableIPs() 中调用
	// 此时还没有IP列表，需要等待UpdateAvailableIPs()初始化IP后启动
}

// SetReassignCallback 设置IP重新分配回调
func (m *UserIPMapper) SetReassignCallback(callback func(userID, oldIP, newIP string)) {
	m.onIPReassignCallback = callback
}

// OnIPFailedAutoReassign 当IP失败时自动重新分配用户
func (m *UserIPMapper) OnIPFailedAutoReassign(event health.IPFailureEvent) {
	log.Printf("🔧 处理IP失败事件: %s, 受影响用户: %d个", event.IP, len(event.AffectedUsers))

	// 为每个受影响的用户重新分配IP
	for _, userID := range event.AffectedUsers {
		newIP, err := m.reassignFailedUserToHealthyIP(userID, event.IP)
		if err != nil {
			log.Printf("❌ 用户 %s 重新分配IP失败: %v", userID, err)
			continue
		}

		log.Printf("✓ 用户 %s: %s -> %s (原IP失败)", userID, event.IP, newIP)

		// 触发回调通知
		m.notifyIPReassigned(userID, event.IP, newIP)
	}
}

// reassignFailedUserToHealthyIP 为用户重新分配健康的IP（排除失败的IP）
func (m *UserIPMapper) reassignFailedUserToHealthyIP(userID, failedIP string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 验证用户确实绑定到失败的IP
	currentIP, exists := m.userToIP[userID]
	if !exists {
		return "", fmt.Errorf("用户不存在")
	}
	if currentIP != failedIP {
		// 用户已经被重新分配了
		return currentIP, nil
	}

	// 移除旧映射
	m.unbindUserFromIP(userID, failedIP)
	m.asyncRemoveFromStorage(userID, failedIP)

	// 分配新IP（排除失败的IP）
	excludeIPs := []string{failedIP}
	newIP, err := m.allocateIPForUser(userID, excludeIPs)
	if err != nil {
		return "", err
	}

	// 保存新映射
	m.asyncSaveToStorage(userID, newIP)

	return newIP, nil
}

// ============ 核心分配逻辑 ============

// allocateIPForUser 为用户分配IP（核心分配逻辑）
//
// 参数：
//   - userID: 用户ID
//   - excludeIPs: 需要排除的IP列表（用于重新分配场景）
//
// 返回：
//   - 分配的IP地址
//   - 错误信息
func (m *UserIPMapper) allocateIPForUser(userID string, excludeIPs []string) (string, error) {
	// 1. 获取候选IP列表
	candidateIPs := m.getCandidateIPs(excludeIPs)
	if len(candidateIPs) == 0 {
		return "", fmt.Errorf("没有可用的IP")
	}

	// 2. 选择负载最少的IP
	selectedIP := m.selectLeastLoadedIP(candidateIPs)

	// 3. 建立绑定关系
	m.bindUserToIP(userID, selectedIP)

	// 4. 异步保存到存储
	m.asyncSaveToStorage(userID, selectedIP)

	return selectedIP, nil
}

// getCandidateIPs 获取候选IP列表（优先健康IP，可排除指定IP）
func (m *UserIPMapper) getCandidateIPs(excludeIPs []string) []string {
	// 1. 从healthChecker获取候选IP
	var candidateIPs []string
	if m.healthChecker != nil {
		// 优先使用健康的IP
		candidateIPs = m.healthChecker.GetHealthyIPs()
		if len(candidateIPs) == 0 {
			// 没有健康IP，使用所有IP（降级策略）
			log.Printf("⚠️  没有健康IP，使用所有IP作为候选")
			candidateIPs = m.healthChecker.GetAllIPs()
		}
	} else {
		// 未启用健康检查时，从ipToUsers获取所有IP（向后兼容，支持测试）
		candidateIPs = make([]string, 0, len(m.ipToUsers))
		for ip := range m.ipToUsers {
			candidateIPs = append(candidateIPs, ip)
		}
	}

	// 2. 过滤掉需要排除的IP
	if len(excludeIPs) == 0 {
		return candidateIPs
	}

	excludeSet := make(map[string]bool)
	for _, ip := range excludeIPs {
		excludeSet[ip] = true
	}

	filtered := make([]string, 0, len(candidateIPs))
	for _, ip := range candidateIPs {
		if !excludeSet[ip] {
			filtered = append(filtered, ip)
		}
	}

	// 3. 如果过滤后为空，但原本有候选IP，使用原候选列表（二次降级策略）
	// 这种情况发生在：所有IP都在excludeIPs中，但系统必须分配一个IP
	if len(filtered) == 0 && len(candidateIPs) > 0 {
		log.Printf("⚠️  过滤后无可用IP（已排除%d个），降级使用所有候选IP", len(excludeIPs))
		// 返回未过滤的候选IP列表，让系统至少有IP可用
		return candidateIPs
	}

	return filtered
}

// selectLeastLoadedIP 选择负载最少的IP（用户数最少）
func (m *UserIPMapper) selectLeastLoadedIP(ips []string) string {
	var selectedIP string
	minUserCount := -1

	for _, ip := range ips {
		count := m.ipUserCount[ip]
		if minUserCount == -1 || count < minUserCount {
			minUserCount = count
			selectedIP = ip
		}
	}

	return selectedIP
}

// ============ 映射关系管理 ============

// bindUserToIP 建立用户到IP的绑定关系（需要持锁调用）
func (m *UserIPMapper) bindUserToIP(userID, ip string) {
	m.userToIP[userID] = ip
	m.ipToUsers[ip] = append(m.ipToUsers[ip], userID)
	m.ipUserCount[ip]++
}

// unbindUserFromIP 解除用户到IP的绑定关系（需要持锁调用）
func (m *UserIPMapper) unbindUserFromIP(userID, ip string) {
	// 删除正向映射
	delete(m.userToIP, userID)

	// 删除反向映射
	if users, exists := m.ipToUsers[ip]; exists {
		newUsers := make([]string, 0, len(users)-1)
		for _, uid := range users {
			if uid != userID {
				newUsers = append(newUsers, uid)
			}
		}
		m.ipToUsers[ip] = newUsers
		m.ipUserCount[ip] = len(newUsers)
	}
}

// ============ 存储操作 ============

// asyncSaveToStorage 异步保存映射到存储
func (m *UserIPMapper) asyncSaveToStorage(userID, ip string) {
	if m.storage == nil {
		return
	}

	go func() {
		if err := m.storage.AddUserMapping(userID, ip); err != nil {
			log.Printf("⚠️  保存用户映射失败: %v", err)
		}
	}()
}

// asyncRemoveFromStorage 异步从存储删除映射
func (m *UserIPMapper) asyncRemoveFromStorage(userID, ip string) {
	if m.storage == nil {
		return
	}

	go func() {
		if err := m.storage.RemoveUserMapping(userID, ip); err != nil {
			log.Printf("⚠️  删除用户映射失败: %v", err)
		}
	}()
}

// loadMappingsFromStorage 从存储加载映射关系
func (m *UserIPMapper) loadMappingsFromStorage() error {
	if m.storage == nil {
		return nil
	}

	data, err := m.storage.Load()
	if err != nil {
		return err
	}

	if data == nil {
		return nil
	}

	m.userToIP = data.UserToIP
	m.ipToUsers = data.IPToUsers

	// 重建用户计数
	m.ipUserCount = make(map[string]int)
	for ip, users := range m.ipToUsers {
		m.ipUserCount[ip] = len(users)
	}

	return nil
}

// ============ 通知回调 ============

// notifyIPReassigned 通知IP重新分配（异步调用，带panic保护）
func (m *UserIPMapper) notifyIPReassigned(userID, oldIP, newIP string) {
	if m.onIPReassignCallback == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("⚠️  重新分配回调panic: %v", r)
			}
		}()
		m.onIPReassignCallback(userID, oldIP, newIP)
	}()
}

// ============ 辅助方法 ============
