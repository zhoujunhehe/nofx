package api

import (
    "fmt"
    "log"
    "net/http"
    "nofx/manager"
    "os"
    "path/filepath"
    "strings"

    "github.com/gin-gonic/gin"
)

// Server HTTP API服务器
type Server struct {
	router        *gin.Engine
	traderManager *manager.TraderManager
	port          int
}

// NewServer 创建API服务器
func NewServer(traderManager *manager.TraderManager, port int) *Server {
	// 设置为Release模式（减少日志输出）
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	// 启用CORS
	router.Use(corsMiddleware())

	s := &Server{
		router:        router,
		traderManager: traderManager,
		port:          port,
	}

    // 设置路由
    s.setupRoutes()

    // 托管前端静态文件（如果存在）
    s.setupFrontend()

    return s
}

// corsMiddleware CORS中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.Any("/health", s.handleHealth)

	// API路由组
	api := s.router.Group("/api")
	{
		// 竞赛总览
		api.GET("/competition", s.handleCompetition)

		// Trader列表
		api.GET("/traders", s.handleTraderList)

		// 指定trader的数据（使用query参数 ?trader_id=xxx）
		api.GET("/status", s.handleStatus)
		api.GET("/account", s.handleAccount)
		api.GET("/positions", s.handlePositions)
		api.GET("/decisions", s.handleDecisions)
		api.GET("/decisions/latest", s.handleLatestDecisions)
		api.GET("/statistics", s.handleStatistics)
		api.GET("/equity-history", s.handleEquityHistory)
		api.GET("/performance", s.handlePerformance)
	}
}

// setupFrontend 托管 web/dist 静态资源，并提供 SPA 回退
func (s *Server) setupFrontend() {
    distDir := filepath.Join("web", "dist")
    indexPath := filepath.Join(distDir, "index.html")
    if _, err := os.Stat(indexPath); err != nil {
        log.Printf("ℹ️ 未找到前端构建产物（%s），仅启动 API 路由", indexPath)
        // 提供一个友好的占位首页，避免根路径超时
        s.router.GET("/", func(c *gin.Context) {
            c.Header("Content-Type", "text/html; charset=utf-8")
            c.String(http.StatusOK, `<!doctype html>
<html><head><meta charset="utf-8"/><title>NOFX</title></head>
<body style="font-family: -apple-system, BlinkMacSystemFont, Segoe UI, Roboto, Arial, sans-serif; padding: 24px;">
  <h1>NOFX 服务已启动</h1>
  <p>未找到前端构建产物 <code>web/dist</code>。当前仅提供 API。</p>
  <p>
    健康检查：<a href="/health">/health</a><br/>
    竞赛总览：<a href="/api/competition">/api/competition</a><br/>
    Trader 列表：<a href="/api/traders">/api/traders</a>
  </p>
  <p style="color:#888">如果你在 Railway 上部署，请确保 Docker 构建阶段成功执行了前端构建（npm ci && npm run build）。</p>
 </body></html>`)
        })
        return
    }

    // 静态资源目录（Vite 默认 assets 路径）
    s.router.Static("/assets", filepath.Join(distDir, "assets"))

    // 首页
    s.router.GET("/", func(c *gin.Context) {
        c.File(indexPath)
    })

    // SPA 回退：非 /api 和 /health 的未命中路由返回 index.html
    s.router.NoRoute(func(c *gin.Context) {
        p := c.Request.URL.Path
        if strings.HasPrefix(p, "/api") || p == "/health" {
            c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
            return
        }
        // 仅当浏览器请求 HTML 时回退
        if strings.Contains(c.GetHeader("Accept"), "text/html") {
            c.File(indexPath)
            return
        }
        c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    })

    log.Printf("🖥️ 前端已托管: %s", distDir)
}

// handleHealth 健康检查
func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   c.Request.Context().Value("time"),
	})
}

// getTraderFromQuery 从query参数获取trader
func (s *Server) getTraderFromQuery(c *gin.Context) (*manager.TraderManager, string, error) {
	traderID := c.Query("trader_id")
	if traderID == "" {
		// 如果没有指定trader_id，返回第一个trader
		ids := s.traderManager.GetTraderIDs()
		if len(ids) == 0 {
			return nil, "", fmt.Errorf("没有可用的trader")
		}
		traderID = ids[0]
	}
	return s.traderManager, traderID, nil
}

// handleCompetition 竞赛总览（对比所有trader）
func (s *Server) handleCompetition(c *gin.Context) {
	comparison, err := s.traderManager.GetComparisonData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取对比数据失败: %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, comparison)
}

// handleTraderList trader列表
func (s *Server) handleTraderList(c *gin.Context) {
	traders := s.traderManager.GetAllTraders()
	result := make([]map[string]interface{}, 0, len(traders))

	for _, t := range traders {
		result = append(result, map[string]interface{}{
			"trader_id":   t.GetID(),
			"trader_name": t.GetName(),
			"ai_model":    t.GetAIModel(),
		})
	}

	c.JSON(http.StatusOK, result)
}

// handleStatus 系统状态
func (s *Server) handleStatus(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	status := trader.GetStatus()
	c.JSON(http.StatusOK, status)
}

// handleAccount 账户信息
func (s *Server) handleAccount(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	log.Printf("📊 收到账户信息请求 [%s]", trader.GetName())
	account, err := trader.GetAccountInfo()
	if err != nil {
		log.Printf("❌ 获取账户信息失败 [%s]: %v", trader.GetName(), err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取账户信息失败: %v", err),
		})
		return
	}

	log.Printf("✓ 返回账户信息 [%s]: 净值=%.2f, 可用=%.2f, 盈亏=%.2f (%.2f%%)",
		trader.GetName(),
		account["total_equity"],
		account["available_balance"],
		account["total_pnl"],
		account["total_pnl_pct"])
	c.JSON(http.StatusOK, account)
}

// handlePositions 持仓列表
func (s *Server) handlePositions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	positions, err := trader.GetPositions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取持仓列表失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, positions)
}

// handleDecisions 决策日志列表
func (s *Server) handleDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取所有历史决策记录（无限制）
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, records)
}

// handleLatestDecisions 最新决策日志（最近5条，最新的在前）
func (s *Server) handleLatestDecisions(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	records, err := trader.GetDecisionLogger().GetLatestRecords(5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取决策日志失败: %v", err),
		})
		return
	}

	// 反转数组，让最新的在前面（用于列表显示）
	// GetLatestRecords返回的是从旧到新（用于图表），这里需要从新到旧
	for i, j := 0, len(records)-1; i < j; i, j = i+1, j-1 {
		records[i], records[j] = records[j], records[i]
	}

	c.JSON(http.StatusOK, records)
}

// handleStatistics 统计信息
func (s *Server) handleStatistics(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	stats, err := trader.GetDecisionLogger().GetStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取统计信息失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleEquityHistory 收益率历史数据
func (s *Server) handleEquityHistory(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 获取尽可能多的历史数据（几天的数据）
	// 每3分钟一个周期：10000条 = 约20天的数据
	records, err := trader.GetDecisionLogger().GetLatestRecords(10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取历史数据失败: %v", err),
		})
		return
	}

	// 构建收益率历史数据点
	type EquityPoint struct {
		Timestamp        string  `json:"timestamp"`
		TotalEquity      float64 `json:"total_equity"`      // 账户净值（wallet + unrealized）
		AvailableBalance float64 `json:"available_balance"` // 可用余额
		TotalPnL         float64 `json:"total_pnl"`         // 总盈亏（相对初始余额）
		TotalPnLPct      float64 `json:"total_pnl_pct"`     // 总盈亏百分比
		PositionCount    int     `json:"position_count"`    // 持仓数量
		MarginUsedPct    float64 `json:"margin_used_pct"`   // 保证金使用率
		CycleNumber      int     `json:"cycle_number"`
	}

	// 从AutoTrader获取初始余额（用于计算盈亏百分比）
	initialBalance := 0.0
	if status := trader.GetStatus(); status != nil {
		if ib, ok := status["initial_balance"].(float64); ok && ib > 0 {
			initialBalance = ib
		}
	}

	// 如果无法从status获取，且有历史记录，则从第一条记录获取
	if initialBalance == 0 && len(records) > 0 {
		// 第一条记录的equity作为初始余额
		initialBalance = records[0].AccountState.TotalBalance
	}

	// 如果还是无法获取，返回错误
	if initialBalance == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "无法获取初始余额",
		})
		return
	}

	var history []EquityPoint
	for _, record := range records {
		// TotalBalance字段实际存储的是TotalEquity
		totalEquity := record.AccountState.TotalBalance
		// TotalUnrealizedProfit字段实际存储的是TotalPnL（相对初始余额）
		totalPnL := record.AccountState.TotalUnrealizedProfit

		// 计算盈亏百分比
		totalPnLPct := 0.0
		if initialBalance > 0 {
			totalPnLPct = (totalPnL / initialBalance) * 100
		}

		history = append(history, EquityPoint{
			Timestamp:        record.Timestamp.Format("2006-01-02 15:04:05"),
			TotalEquity:      totalEquity,
			AvailableBalance: record.AccountState.AvailableBalance,
			TotalPnL:         totalPnL,
			TotalPnLPct:      totalPnLPct,
			PositionCount:    record.AccountState.PositionCount,
			MarginUsedPct:    record.AccountState.MarginUsedPct,
			CycleNumber:      record.CycleNumber,
		})
	}

	c.JSON(http.StatusOK, history)
}

// handlePerformance AI历史表现分析（用于展示AI学习和反思）
func (s *Server) handlePerformance(c *gin.Context) {
	_, traderID, err := s.getTraderFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trader, err := s.traderManager.GetTrader(traderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 分析最近100个周期的交易表现（避免长期持仓的交易记录丢失）
	// 假设每3分钟一个周期，100个周期 = 5小时，足够覆盖大部分交易
	performance, err := trader.GetDecisionLogger().AnalyzePerformance(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("分析历史表现失败: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, performance)
}

// Start 启动服务器
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("🌐 API服务器启动在 http://localhost%s", addr)
	log.Printf("📊 API文档:")
	log.Printf("  • GET  /api/competition      - 竞赛总览（对比所有trader）")
	log.Printf("  • GET  /api/traders          - Trader列表")
	log.Printf("  • GET  /api/status?trader_id=xxx     - 指定trader的系统状态")
	log.Printf("  • GET  /api/account?trader_id=xxx    - 指定trader的账户信息")
	log.Printf("  • GET  /api/positions?trader_id=xxx  - 指定trader的持仓列表")
	log.Printf("  • GET  /api/decisions?trader_id=xxx  - 指定trader的决策日志")
	log.Printf("  • GET  /api/decisions/latest?trader_id=xxx - 指定trader的最新决策")
	log.Printf("  • GET  /api/statistics?trader_id=xxx - 指定trader的统计信息")
	log.Printf("  • GET  /api/equity-history?trader_id=xxx - 指定trader的收益率历史数据")
	log.Printf("  • GET  /api/performance?trader_id=xxx - 指定trader的AI学习表现分析")
	log.Printf("  • GET  /health               - 健康检查")
	log.Println()

	return s.router.Run(addr)
}
