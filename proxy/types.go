package proxy

import (
	"net/http"
)

// ProxyClient 代理客户端
type ProxyClient struct {
	ProxyID      int    // IP池中的代理ID（索引）
	IP           string // 使用的IP地址
	*http.Client        // HTTP客户端
}
