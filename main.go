package main

import (
    "context"
    "fmt"
    "io"
    "log"
    "net"
    "net/http"
    "nofx/api"
    "nofx/config"
    "nofx/manager"
    "nofx/pool"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║    🏆 AI模型交易竞赛系统 - Qwen vs DeepSeek               ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// 将标准日志输出重定向到 stdout，避免在 Railway 等平台被标记为 error（stderr）
	log.SetOutput(os.Stdout)

	// 加载配置文件
	configFile := "config.json"
	if len(os.Args) > 1 {
		configFile = os.Args[1]
	}

    log.Printf("📋 加载配置文件: %s", configFile)
    cfg, err := config.LoadConfig(configFile)
    if err != nil {
        log.Fatalf("❌ 加载配置失败: %v", err)
    }

    log.Printf("✓ 配置加载成功，共%d个trader参赛", len(cfg.Traders))
    fmt.Println()

    // Railway/Nixpacks: 如果存在环境变量 PORT，则覆盖配置文件中的端口
    if p := os.Getenv("PORT"); p != "" {
        if port, err := strconv.Atoi(p); err == nil && port > 0 {
            if port != cfg.APIServerPort {
                log.Printf("🔧 检测到环境变量 PORT=%d，覆盖 api_server_port=%d", port, cfg.APIServerPort)
            }
            cfg.APIServerPort = port
        } else {
            log.Printf("⚠️  环境变量 PORT='%s' 非法，继续使用配置端口 %d", p, cfg.APIServerPort)
        }
    }

    // 打印当前主机出口 IP（最佳努力，超时快速返回）
    if ip := detectPublicIP(); ip != "" {
        log.Printf("🌐 当前主机出口IP: %s", ip)
    } else {
        log.Printf("🌐 当前主机出口IP: 未能获取（可能无外网或服务超时）")
    }

	// 设置默认主流币种列表
	pool.SetDefaultCoins(cfg.DefaultCoins)

	// 设置是否使用默认主流币种
	pool.SetUseDefaultCoins(cfg.UseDefaultCoins)
	if cfg.UseDefaultCoins {
		log.Printf("✓ 已启用默认主流币种列表（共%d个币种）: %v", len(cfg.DefaultCoins), cfg.DefaultCoins)
	}

	// 设置币种池API URL
	if cfg.CoinPoolAPIURL != "" {
		pool.SetCoinPoolAPI(cfg.CoinPoolAPIURL)
		log.Printf("✓ 已配置AI500币种池API")
	}
	if cfg.OITopAPIURL != "" {
		pool.SetOITopAPI(cfg.OITopAPIURL)
		log.Printf("✓ 已配置OI Top API")
	}

	// 创建TraderManager
	traderManager := manager.NewTraderManager()

	// 添加所有启用的trader
	enabledCount := 0
	for i, traderCfg := range cfg.Traders {
		// 跳过未启用的trader
		if !traderCfg.Enabled {
			log.Printf("⏭️  [%d/%d] 跳过未启用的 %s", i+1, len(cfg.Traders), traderCfg.Name)
			continue
		}

		enabledCount++
		log.Printf("📦 [%d/%d] 初始化 %s (%s模型)...",
			i+1, len(cfg.Traders), traderCfg.Name, strings.ToUpper(traderCfg.AIModel))

		err := traderManager.AddTrader(
			traderCfg,
			cfg.CoinPoolAPIURL,
			cfg.MaxDailyLoss,
			cfg.MaxDrawdown,
			cfg.StopTradingMinutes,
			cfg.Leverage, // 传递杠杆配置
		)
		if err != nil {
			log.Fatalf("❌ 初始化trader失败: %v", err)
		}
	}

	// 检查是否至少有一个启用的trader
	if enabledCount == 0 {
		log.Fatalf("❌ 没有启用的trader，请在config.json中设置至少一个trader的enabled=true")
	}

	fmt.Println()
	fmt.Println("🏁 竞赛参赛者:")
	for _, traderCfg := range cfg.Traders {
		// 只显示启用的trader
		if !traderCfg.Enabled {
			continue
		}
		fmt.Printf("  • %s (%s) - 初始资金: %.0f USDT\n",
			traderCfg.Name, strings.ToUpper(traderCfg.AIModel), traderCfg.InitialBalance)
	}

	fmt.Println()
	fmt.Println("🤖 AI全权决策模式:")
	fmt.Printf("  • AI将自主决定每笔交易的杠杆倍数（山寨币最高%d倍，BTC/ETH最高%d倍）\n",
		cfg.Leverage.AltcoinLeverage, cfg.Leverage.BTCETHLeverage)
	fmt.Println("  • AI将自主决定每笔交易的仓位大小")
	fmt.Println("  • AI将自主设置止损和止盈价格")
	fmt.Println("  • AI将基于市场数据、技术指标、账户状态做出全面分析")
	fmt.Println()
	fmt.Println("⚠️  风险提示: AI自动交易有风险，建议小额资金测试！")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止运行")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// 创建并启动API服务器
	apiServer := api.NewServer(traderManager, cfg.APIServerPort)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("❌ API服务器错误: %v", err)
		}
	}()

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 启动所有trader
	traderManager.StartAll()

	// 等待退出信号
	<-sigChan
	fmt.Println()
	fmt.Println()
	log.Println("📛 收到退出信号，正在停止所有trader...")
	traderManager.StopAll()

	fmt.Println()
	fmt.Println("👋 感谢使用AI交易竞赛系统！")
}

// detectPublicIP 尝试通过多个公共服务获取当前主机的出口 IP。
// 返回空字符串表示未获取到。
func detectPublicIP() string {
    endpoints := []string{
        "https://api.ipify.org?format=text",
        "https://ifconfig.me/ip",
        "https://ipinfo.io/ip",
        "https://checkip.amazonaws.com",
    }

    client := &http.Client{Timeout: 3 * time.Second}

    for _, url := range endpoints {
        // 为每次请求设置最短超时与取消控制
        ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
        req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
        if err != nil {
            cancel()
            continue
        }
        // 简单标识
        req.Header.Set("User-Agent", "nofx-egress-ip-check/1.0")

        resp, err := client.Do(req)
        if err != nil {
            cancel()
            continue
        }
        body, _ := io.ReadAll(resp.Body)
        resp.Body.Close()
        cancel()

        ipStr := strings.TrimSpace(string(body))
        if ip := net.ParseIP(ipStr); ip != nil {
            return ipStr
        }
    }
    return ""
}
