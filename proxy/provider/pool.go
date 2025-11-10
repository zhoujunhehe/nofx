package provider

import "strings"

// PoolProvider 固定IP列表提供者（代理池）
type PoolProvider struct {
	ips []ProxyIP
}

// NewPoolProvider 创建固定IP列表提供者
func NewPoolProvider(proxyURLs []string) *PoolProvider {
	ips := make([]ProxyIP, 0, len(proxyURLs))
	for _, proxyURL := range proxyURLs {
		// 简单解析代理URL
		// 格式: http://ip:port 或 socks5://user:pass@ip:port
		protocol := "http"
		if strings.HasPrefix(proxyURL, "socks5://") {
			protocol = "socks5"
			proxyURL = strings.TrimPrefix(proxyURL, "socks5://")
		} else if strings.HasPrefix(proxyURL, "http://") {
			proxyURL = strings.TrimPrefix(proxyURL, "http://")
		} else if strings.HasPrefix(proxyURL, "https://") {
			protocol = "https"
			proxyURL = strings.TrimPrefix(proxyURL, "https://")
		}

		ips = append(ips, ProxyIP{
			IP:       proxyURL,
			Protocol: protocol,
		})
	}

	return &PoolProvider{ips: ips}
}

func (p *PoolProvider) GetIPList() ([]ProxyIP, error) {
	return p.ips, nil
}

func (p *PoolProvider) RefreshIPList() ([]ProxyIP, error) {
	return p.ips, nil
}
