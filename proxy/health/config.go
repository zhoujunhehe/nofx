package health

import "time"

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	CheckURLs        []string      // 探活URL列表（通过代理访问这些URL）
	Interval         time.Duration // 探活间隔（默认5秒）
	Timeout          time.Duration // 单次超时（默认3秒）
	FailureThreshold int           // 失败阈值（默认5次）
}

// WithDefaults 设置默认值
func (c *HealthCheckConfig) WithDefaults() HealthCheckConfig {
	config := *c
	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}
	if config.FailureThreshold == 0 {
		config.FailureThreshold = 5
	}
	if len(config.CheckURLs) == 0 {
		config.CheckURLs = []string{
			"https://api.binance.com/api/v3/ping",
		}
	}
	return config
}
