package proxy

import (
	"nofx/proxy/health"
	"nofx/proxy/mapper"
	"nofx/proxy/provider"
	"nofx/proxy/storage"
)

// ============================================================================
// 公共API - Re-export关键类型，确保向后兼容
// ============================================================================

// Provider相关类型
type (
	// ProxyIP 代理IP信息
	ProxyIP = provider.ProxyIP
	// IPProvider IP提供者接口
	IPProvider = provider.IPProvider
)

// Storage相关类型
type (
	// Storage 存储接口
	Storage = storage.Storage
	// MappingData 映射数据
	MappingData = storage.MappingData
	// StorageConfig 存储配置
	StorageConfig = storage.Config
)

// Health相关类型
type (
	// HealthStatus 健康状态
	HealthStatus = health.HealthStatus
	// IPHealth IP健康信息
	IPHealth = health.IPHealth
	// HealthCheckConfig 健康检查配置
	HealthCheckConfig = health.HealthCheckConfig
	// HealthChecker 健康检查器
	HealthChecker = health.HealthChecker
	// IPFailureEvent IP失败事件
	IPFailureEvent = health.IPFailureEvent
)

// Mapper相关类型
type (
	// UserIPMapper 用户IP映射器
	UserIPMapper = mapper.UserIPMapper
)

// ============================================================================
// Provider工厂函数
// ============================================================================

// CreateProvider 创建IP提供者
// mode: "single", "pool", "brightdata"
func CreateProvider(mode, proxyURL string, proxyList []string, brightdataEndpoint, brightdataToken, brightdataZone string) (IPProvider, error) {
	return provider.CreateProvider(mode, proxyURL, proxyList, brightdataEndpoint, brightdataToken, brightdataZone)
}

// ============================================================================
// Storage工厂函数
// ============================================================================

// CreateStorage 创建存储实例
func CreateStorage(cfg StorageConfig) (Storage, error) {
	return storage.CreateStorage(cfg)
}

// ============================================================================
// 健康状态常量
// ============================================================================

const (
	// StatusHealthy 健康状态
	StatusHealthy = health.StatusHealthy
	// StatusTemporaryFailed 临时失败状态
	StatusTemporaryFailed = health.StatusTemporaryFailed
	// StatusBlacklisted 黑名单状态
	StatusBlacklisted = health.StatusBlacklisted
)

// ============================================================================
// Mapper工厂函数
// ============================================================================

// NewUserIPMapper 创建用户IP映射器
func NewUserIPMapper(stor Storage) *UserIPMapper {
	return mapper.NewUserIPMapper(stor)
}

// ============================================================================
// Health工厂函数
// ============================================================================

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config HealthCheckConfig) *HealthChecker {
	return health.NewHealthChecker(config)
}
