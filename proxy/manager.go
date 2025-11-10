package proxy

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"nofx/proxy/health"
	"nofx/proxy/mapper"
	"nofx/proxy/provider"
	"nofx/proxy/storage"
	"sync"
	"time"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	config   *Config
	provider provider.IPProvider

	// IP池管理
	ipList []provider.ProxyIP
	mutex  sync.RWMutex // 读写锁，保证线程安全

	// 用户IP映射器（持久化映射关系）
	userIPMapper *mapper.UserIPMapper

	// 刷新控制
	stopRefresh     chan struct{}
	stopRefreshOnce sync.Once // 确保stopRefresh只关闭一次
}

var (
	globalProxyManager *ProxyManager
	initOnce           sync.Once
	initErr            error
)

// InitGlobalProxyManager 初始化全局代理管理器
func InitGlobalProxyManager(config *Config) error {
	initOnce.Do(func() {
		var manager *ProxyManager
		manager, initErr = NewProxyManager(config)
		if initErr != nil {
			log.Printf("❌ 代理管理器初始化失败: %v", initErr)
			return
		}

		// 初始化成功，设置全局变量
		globalProxyManager = manager

		// 启动自动刷新
		if config.Enabled && config.RefreshInterval > 0 {
			globalProxyManager.StartAutoRefresh()
		}
	})

	// 如果初始化失败，返回错误
	if initErr != nil {
		return fmt.Errorf("代理管理器初始化失败: %w", initErr)
	}

	return nil
}

// GetGlobalProxyManager 获取全局代理管理器
func GetGlobalProxyManager() *ProxyManager {
	if globalProxyManager == nil {
		// 如果未初始化，使用默认配置（禁用代理）
		_ = InitGlobalProxyManager(&Config{Enabled: false})
	}
	return globalProxyManager
}

// NewProxyManager 创建代理管理器
func NewProxyManager(config *Config) (*ProxyManager, error) {
	if config == nil {
		config = &Config{Enabled: false}
	}

	// 设置默认值
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RefreshInterval == 0 && config.Mode == "brightdata" {
		config.RefreshInterval = 30 * time.Minute // 默认 30 分钟刷新一次
	}

	// 创建存储（根据配置选择类型）
	stor, err := storage.CreateStorage(storage.Config{
		Type:          config.StorageType,
		FilePath:      config.MappingFile,
		FlushInterval: config.FlushInterval,
		RedisAddr:     config.RedisAddr,
		RedisPassword: config.RedisPassword,
		RedisDB:       config.RedisDB,
		DBDriver:      config.DBDriver,
		DBDSN:         config.DBDSN,
	})
	if err != nil {
		return nil, fmt.Errorf("创建存储失败: %w", err)
	}

	userIPMapper := mapper.NewUserIPMapper(stor)
	// 启用用户IP映射器的健康检查（如果配置了）
	userIPMapper.EnableHealthCheck((&health.HealthCheckConfig{
		CheckURLs:        config.HealthCheckURLs,
		Interval:         config.HealthCheckInterval,
		Timeout:          config.HealthCheckTimeout,
		FailureThreshold: config.HealthCheckFailureThreshold,
	}).WithDefaults())

	m := &ProxyManager{
		config:       config,
		userIPMapper: userIPMapper,
		stopRefresh:  make(chan struct{}),
	}

	// 如果未启用代理，直接返回
	if !config.Enabled {
		log.Printf("🌐 HTTP 代理未启用，使用直连")
		return m, nil
	}

	// 创建IP提供者
	if err := m.CreateProvider(); err != nil {
		return nil, fmt.Errorf("创建IP提供者失败: %w", err)
	}

	// 初始化IP列表
	if err := m.RefreshIPList(); err != nil {
		return nil, fmt.Errorf("初始化IP列表失败: %w", err)
	}

	return m, nil
}

func (m *ProxyManager) CreateProvider() error {
	var err error
	// 使用provider工厂函数创建
	provider.Username = m.config.ProxyUser
	provider.Password = m.config.ProxyPassword
	provider.Host = m.config.ProxyHost

	m.provider, err = provider.CreateProvider(
		m.config.Mode,
		m.config.ProxyURL,
		m.config.ProxyList,
		m.config.BrightDataEndpoint,
		m.config.BrightDataToken,
		m.config.BrightDataZone,
	)
	return err
}

// RefreshIPList 刷新IP列表（线程安全）
func (m *ProxyManager) RefreshIPList() error {
	if m.provider == nil {
		return nil
	}

	ips, err := m.provider.RefreshIPList()
	if err != nil {
		return err
	}

	m.mutex.Lock()
	m.ipList = ips
	m.mutex.Unlock()

	// 更新 mapper 的可用IP列表
	m.userIPMapper.UpdateAvailableIPs(ips)

	log.Printf("✓ 刷新代理IP列表: %d个", len(ips))

	return nil
}

// StartAutoRefresh 启动自动刷新
func (m *ProxyManager) StartAutoRefresh() {
	if m.config.RefreshInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(m.config.RefreshInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := m.RefreshIPList(); err != nil {
					log.Printf("⚠️  自动刷新IP列表失败: %v", err)
				}
			case <-m.stopRefresh:
				return
			}
		}
	}()

	log.Printf("✓ 已启动代理IP自动刷新 (间隔: %v)", m.config.RefreshInterval)
}

// StopAutoRefresh 停止自动刷新
func (m *ProxyManager) StopAutoRefresh() {
	m.stopRefreshOnce.Do(func() {
		close(m.stopRefresh)
	})
}

// getRandomProxy 随机获取一个可用代理（线程安全 - 读锁，优先选择健康IP）
func (m *ProxyManager) getRandomProxy() (int, *provider.ProxyIP, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if len(m.ipList) == 0 {
		return -1, nil, fmt.Errorf("代理IP列表为空")
	}

	// 找到所有健康IP的索引
	healthyIndices := make([]int, 0, len(m.ipList))
	for i := range m.ipList {
		if m.userIPMapper.IsIPHealthy(m.ipList[i].IP) {
			healthyIndices = append(healthyIndices, i)
		}
	}

	// 如果有健康IP，从健康IP中随机选择
	if len(healthyIndices) > 0 {
		randomIdx := healthyIndices[rand.Intn(len(healthyIndices))]
		return randomIdx, &m.ipList[randomIdx], nil
	}

	// 如果所有IP都不健康，从所有IP中随机选择（降级策略）
	log.Printf("⚠️  所有代理IP都不健康，使用降级策略")
	randomIdx := rand.Intn(len(m.ipList))
	return randomIdx, &m.ipList[randomIdx], nil
}

// createDirectClient 创建直连客户端（未启用代理时使用）
func (m *ProxyManager) createDirectClient() *ProxyClient {
	return &ProxyClient{
		ProxyID: -1, // -1 表示未使用代理
		IP:      "direct",
		Client: &http.Client{
			Timeout: m.config.Timeout,
		},
	}
}

// findProxyByIP 在ipList中查找IP对应的ProxyIP和索引（线程安全 - 需要调用方持有读锁）
func (m *ProxyManager) findProxyByIP(ip string) (int, *provider.ProxyIP) {
	for i, proxyIP := range m.ipList {
		if proxyIP.IP == ip {
			return i, &m.ipList[i]
		}
	}
	return -1, nil
}

// createProxyClient 根据ProxyIP创建代理客户端
func (m *ProxyManager) createProxyClient(proxyID int, proxyIP *provider.ProxyIP) (*ProxyClient, error) {
	// 解析代理URL
	proxyURL, err := url.Parse(proxyIP.ProxyURL())
	if err != nil {
		return nil, fmt.Errorf("解析代理URL失败: %w", err)
	}

	// 创建Transport
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &ProxyClient{
		ProxyID: proxyID,
		IP:      proxyIP.IP,
		Client: &http.Client{
			Transport: transport,
			Timeout:   m.config.Timeout,
		},
	}, nil
}

// reassignUserIP 重新为用户分配健康的IP（当原IP在黑名单时调用）
func (m *ProxyManager) reassignUserIP(userID, oldIP string) (int, *provider.ProxyIP, error) {
	log.Printf("⚠️  用户 %s 绑定的代理IP %s 在黑名单中，重新分配健康IP", userID, oldIP)

	// 通过mapper重新获取IP（会自动选择负载最低的健康IP）
	newIP, err := m.userIPMapper.GetIPForUser(userID)
	if err != nil {
		return -1, nil, fmt.Errorf("为用户 %s 重新分配IP失败: %w", userID, err)
	}

	log.Printf("✓ 用户 %s 已重新分配: %s -> %s", userID, oldIP, newIP)

	// 在ipList中查找新IP的ProxyIP
	m.mutex.RLock()
	proxyID, proxyIP := m.findProxyByIP(newIP)
	m.mutex.RUnlock()

	if proxyIP == nil {
		return -1, nil, fmt.Errorf("找不到新分配的IP %s 在IP列表中", newIP)
	}

	return proxyID, proxyIP, nil
}

// GetProxyClient 获取代理客户端（线程安全）
func (m *ProxyManager) GetProxyClient() (*ProxyClient, error) {
	if !m.config.Enabled {
		return m.createDirectClient(), nil
	}

	// 获取随机代理（使用读锁，确保不越界）
	proxyID, proxyIP, err := m.getRandomProxy()
	if err != nil {
		return nil, err
	}

	return m.createProxyClient(proxyID, proxyIP)
}

// GetProxyClientForUser 根据userID获取绑定的代理客户端（持久化映射）
func (m *ProxyManager) GetProxyClientForUser(userID string) (*ProxyClient, error) {
	if !m.config.Enabled {
		return m.createDirectClient(), nil
	}

	// 从映射器获取用户绑定的IP
	boundIP, err := m.userIPMapper.GetIPForUser(userID)
	if err != nil {
		return nil, err
	}

	// 在ipList中查找对应的ProxyIP和索引
	m.mutex.RLock()
	proxyID, proxyIP := m.findProxyByIP(boundIP)
	m.mutex.RUnlock()

	if proxyIP == nil {
		return nil, fmt.Errorf("找不到IP %s 在IP列表中", boundIP)
	}

	// 检查IP是否健康（如果启用了健康检查）
	if !m.userIPMapper.IsIPHealthy(boundIP) {
		// IP不健康，重新分配健康的IP
		proxyID, proxyIP, err = m.reassignUserIP(userID, boundIP)
		if err != nil {
			return nil, err
		}
	}

	return m.createProxyClient(proxyID, proxyIP)
}

// AddBlacklist 将代理IP添加到黑名单（转发到HealthChecker）
func (m *ProxyManager) AddBlacklist(ip string) {
	// 调用 mapper 的 AddBlacklist，它会转发到 HealthChecker
	m.userIPMapper.AddBlacklist(ip, "手动加入黑名单")
}

// GetBlacklistStatus 获取黑名单状态（通过HealthChecker）
func (m *ProxyManager) GetBlacklistStatus() (total int, blacklisted int, available int) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	total = len(m.ipList)
	blacklisted = 0

	// 统计不健康的IP数量
	for i := 0; i < total; i++ {
		ip := m.ipList[i].IP
		if !m.userIPMapper.IsIPHealthy(ip) {
			blacklisted++
		}
	}

	available = total - blacklisted
	return
}

// IsEnabled 检查代理是否启用
func IsEnabled() bool {
	return GetGlobalProxyManager().config.Enabled
}

// RefreshIPList 刷新全局代理IP列表
func RefreshIPList() error {
	return GetGlobalProxyManager().RefreshIPList()
}

// AddBlacklist 将代理IP添加到全局黑名单
func AddBlacklist(ip string) {
	GetGlobalProxyManager().AddBlacklist(ip)
}

// GetUsersForIP 获取绑定到指定IP的所有用户
func (m *ProxyManager) GetUsersForIP(ip string) []string {
	return m.userIPMapper.GetUsersForIP(ip)
}

// GetAllUserIPMappings 获取所有用户IP映射关系
func (m *ProxyManager) GetAllUserIPMappings() map[string][]string {
	return m.userIPMapper.GetAllMappings()
}

// GetMappingStats 获取映射统计信息
func (m *ProxyManager) GetMappingStats() (totalUsers int, totalIPs int, avgUsersPerIP float64) {
	return m.userIPMapper.GetStats()
}

// RemoveUserMapping 移除用户映射
func (m *ProxyManager) RemoveUserMapping(userID string) error {
	return m.userIPMapper.RemoveUserMapping(userID)
}

// SaveMappings 强制保存映射到文件
func (m *ProxyManager) SaveMappings() error {
	return m.userIPMapper.Flush()
}

// GetUserIPMapper 获取用户IP映射器
func (m *ProxyManager) GetUserIPMapper() *mapper.UserIPMapper {
	return m.userIPMapper
}
