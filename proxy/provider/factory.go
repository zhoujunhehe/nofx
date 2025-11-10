package provider

import "fmt"

const (
	// Provider modes
	ProviderModeSingle    = "single"
	ProviderModePool      = "pool"
	ProviderModeBrightData = "brightdata"
)

// CreateProvider 根据配置创建IP提供者
func CreateProvider(mode, proxyURL string, proxyList []string, brightdataEndpoint, brightdataToken, brightdataZone string) (IPProvider, error) {
	switch mode {
	case ProviderModeSingle:
		// 单个代理模式
		if proxyURL == "" {
			return nil, fmt.Errorf("single模式下必须配置proxy_url")
		}
		return NewSingleProxyProvider(proxyURL), nil

	case ProviderModePool:
		// 代理池模式（固定列表）
		if len(proxyList) == 0 {
			return nil, fmt.Errorf("pool模式下必须配置proxy_list")
		}
		return NewPoolProvider(proxyList), nil

	case ProviderModeBrightData:
		// Bright Data动态获取模式
		if brightdataEndpoint == "" {
			return nil, fmt.Errorf("brightdata模式下必须配置brightdata_endpoint")
		}
		return NewBrightDataProvider(brightdataEndpoint, brightdataToken, brightdataZone), nil

	default:
		// 默认使用single模式
		if proxyURL == "" {
			return nil, fmt.Errorf("未知的proxy模式: %s", mode)
		}
		return NewSingleProxyProvider(proxyURL), nil
	}
}
