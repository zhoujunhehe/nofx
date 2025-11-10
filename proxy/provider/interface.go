package provider

import "fmt"

var (
	Host     string
	Username string
	Password string
)

// ProxyIP 代理IP信息
type ProxyIP struct {
	IP       string                 `json:"ip"`       // IP地址
	Port     string                 `json:"port"`     // 端口（可选）
	Protocol string                 `json:"protocol"` // 协议: http, https, socks5
	Ext      map[string]interface{} `json:"ext"`      // 扩展信息
}

func (p ProxyIP) ProxyURL() string {
	if Host != "" && Username != "" {
		// 使用配置的代理主机和认证信息
		user := Username
		if Username != "" && p.IP != "" {
			// 支持%s占位符替换IP
			user = fmt.Sprintf(Username, p.IP)
		}

		protocol := p.Protocol
		if protocol == "" {
			protocol = "http"
		}

		if Password != "" {
			return fmt.Sprintf("%s://%s:%s@%s", protocol, user, Password, Host)
		}
		return fmt.Sprintf("%s://%s@%s", protocol, user, Host)
	}

	// 直接使用IP信息
	return p.IP
}

// IPProvider IP提供者接口
type IPProvider interface {
	// GetIPList 获取IP列表
	GetIPList() ([]ProxyIP, error)

	// RefreshIPList 刷新IP列表（可选实现）
	RefreshIPList() ([]ProxyIP, error)
}
