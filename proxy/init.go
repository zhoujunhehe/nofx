package proxy

import (
	"fmt"
	"log"
	"net/http"
	"nofx/bootstrap"
	"nofx/hook"

	"github.com/adshao/go-binance/v2/futures"
)

func init() {
	// 注册 Proxy 模块到 bootstrap 系统
	bootstrap.Register("Proxy", bootstrap.PriorityCore, initProxy).
		OnError(bootstrap.WarnOnError) // 代理失败不应该阻止整个系统启动
}

// initProxy Proxy 模块的 bootstrap 初始化函数
func initProxy(ctx *bootstrap.Context) error {
	// 1. 从 bootstrap context 获取配置
	configFile, err := LoadConfig("config.json")
	if err != nil {
		log.Printf("⚠️  加载代理配置失败: %v", err)
		return err
	}

	config := configFile.ConvertToConfig()
	if !config.Enabled {
		log.Printf("ℹ️  代理功能未启用，跳过注册钩子")
		return nil
	}

	// 2. 初始化全局代理管理器
	err = InitGlobalProxyManager(config)
	if err != nil {
		return fmt.Errorf("初始化全局代理管理器失败: %w", err)
	}

	// 3. 注册业务钩子
	registerHooks()

	log.Printf("✓ Proxy 模块初始化完成")
	return nil
}

// convertBootstrapConfig 将 bootstrap config 转换为 proxy Config
func convertBootstrapConfig(cfg interface{}) *Config {
	// 如果已经是 *Config 类型，直接返回
	if proxyConfig, ok := cfg.(*Config); ok {
		return proxyConfig
	}

	// 尝试从 map 或其他类型转换
	// 这里可以根据实际的 config 结构进行调整
	log.Printf("⚠️  Proxy 配置类型不匹配，使用默认配置")
	return &Config{Enabled: false}
}

// registerHooks 注册 Proxy 相关的业务钩子
func registerHooks() {
	// NEW_BINANCE_TRADER: 为用户的 Binance Trader 绑定代理IP
	hook.RegisterHook(hook.NEW_BINANCE_TRADER, func(args ...any) any {
		if len(args) < 2 {
			return &hook.NewBinanceTraderResult{
				Err: fmt.Errorf("insufficient arguments for NewBinanceTrader hook"),
			}
		}

		userId, ok := args[0].(string)
		if !ok {
			return &hook.NewBinanceTraderResult{
				Err: fmt.Errorf("invalid argument for NewBinanceTrader hook"),
			}
		}

		client, ok := args[1].(*futures.Client)
		if !ok {
			return &hook.NewBinanceTraderResult{
				Err: fmt.Errorf("failed to create Binance futures client"),
			}
		}

		proxyMgr := GetGlobalProxyManager()
		// 为用户绑定代理IP
		httpClient, err := proxyMgr.GetProxyClientForUser(userId)
		if err != nil {
			log.Printf("⚠️  获取代理HTTP客户端失败，使用直连: %v", err)
		} else {
			client.HTTPClient = httpClient.Client
			log.Printf("✓ 用户 %s 绑定代理IP: %s", userId, httpClient.IP)
		}

		return &hook.NewBinanceTraderResult{
			Client: client,
		}
	})

	// SET_HTTP_CLIENT: 为 HTTP Client 设置代理
	hook.RegisterHook(hook.SET_HTTP_CLIENT, func(args ...any) any {
		if len(args) < 1 {
			return &hook.SetHttpClientResult{
				Err: fmt.Errorf("insufficient arguments for SetHttpClient hook"),
			}
		}

		client, ok := args[0].(*http.Client)
		if !ok {
			return &hook.SetHttpClientResult{
				Err:    fmt.Errorf("invalid argument for SetHttpClient hook"),
				Client: client,
			}
		}

		proxyClient := GetGlobalProxyManager()
		httpClient, err := proxyClient.GetProxyClient()
		if err != nil {
			log.Printf("⚠️  获取代理HTTP客户端失败，使用直连: %v", err)
			return &hook.SetHttpClientResult{
				Err:    err,
				Client: client,
			}
		}

		client.Transport = httpClient.Client.Transport
		return &hook.SetHttpClientResult{
			Client: client,
		}
	})

	// GETIP: 获取用户绑定的IP
	hook.RegisterHook(hook.GETIP, func(args ...any) any {
		if len(args) < 1 {
			return hook.IpResult{
				Err: fmt.Errorf("insufficient arguments for GETIP hook"),
			}
		}

		userId, ok := args[0].(string)
		if !ok {
			return &hook.IpResult{
				Err: fmt.Errorf("invalid argument for GETIP hook"),
			}
		}

		proxyMgr := GetGlobalProxyManager()
		ip, err := proxyMgr.GetUserIPMapper().GetIPForUser(userId)
		if err != nil {
			return &hook.IpResult{
				Err: err,
			}
		}

		return &hook.IpResult{
			IP: ip,
		}
	})

	log.Printf("✓ Proxy 钩子注册完成")
}
