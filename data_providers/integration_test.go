package data_providers

import (
	"context"
	"strings"
	"testing"

	"nofx/orchestrator"
	"nofx/orchestrator/core"
	"nofx/prompt/sections"
)

// TestIntegration_SectionsProvidersOrchestrator 集成测试
// 测试 Sections + Providers + Orchestrator 完整流程
func TestIntegration_SectionsProvidersOrchestrator(t *testing.T) {
	ctx := context.Background()

	// 1. 创建Orchestrator
	orch := core.New()

	// 2. 创建并注册所有Providers
	userID := "test-user-001"

	configProvider := NewConfigProvider(userID)
	systemProvider := NewSystemProvider(userID)
	accountProvider := NewAccountProvider(userID)
	marketProvider := NewMarketProvider(userID)
	positionProvider := NewPositionProvider(userID)
	candidateProvider := NewCandidateProvider(userID)
	performanceProvider := NewPerformanceProvider(userID)

	providers := []orchestrator.Provider{
		configProvider,
		systemProvider,
		accountProvider,
		marketProvider,
		positionProvider,
		candidateProvider,
		performanceProvider,
	}

	for _, p := range providers {
		if err := orch.RegisterProvider(p); err != nil {
			t.Fatalf("Failed to register provider %s: %v", p.Name(), err)
		}
	}

	// 3. 创建Sections
	riskControlSection := sections.NewRiskControlSection()
	outputFormatSection := sections.NewOutputFormatSection()
	systemStatusSection := sections.NewSystemStatusSection()
	accountSection := sections.NewAccountSection()
	performanceSection := sections.NewPerformanceMetricsSection()

	// 4. 构建PromptBuilder
	// OutputFormatSection包含JSON示例，测试严格的占位符正则能正确区分JSON和占位符
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     riskControlSection.GetName(),
				Template: riskControlSection.GetContent(),
			},
			{
				Name:     outputFormatSection.GetName(),
				Template: outputFormatSection.GetContent(),
			},
		},
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     systemStatusSection.GetName(),
				Template: systemStatusSection.GetContent(),
			},
			{
				Name:     accountSection.GetName(),
				Template: accountSection.GetContent(),
			},
			{
				Name:     performanceSection.GetName(),
				Template: performanceSection.GetContent(),
			},
		},
	}

	// 5. 构建Prompt
	result, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 6. 验证结果
	t.Run("SystemConstraintPrompt不为空", func(t *testing.T) {
		if result.SystemConstraintPrompt == "" {
			t.Error("SystemConstraintPrompt should not be empty")
		}
	})

	t.Run("AnalysisDataPrompt不为空", func(t *testing.T) {
		if result.AnalysisDataPrompt == "" {
			t.Error("AnalysisDataPrompt should not be empty")
		}
	})

	t.Run("占位符已被替换", func(t *testing.T) {
		// 检查SystemConstraintPrompt中不包含未替换的占位符
		if strings.Contains(result.SystemConstraintPrompt, "{config.") {
			t.Error("SystemConstraintPrompt still contains unreplaced config placeholders")
		}

		// 检查AnalysisDataPrompt中不包含未替换的占位符
		if strings.Contains(result.AnalysisDataPrompt, "{system.") {
			t.Error("AnalysisDataPrompt still contains unreplaced system placeholders")
		}
		if strings.Contains(result.AnalysisDataPrompt, "{account.") {
			t.Error("AnalysisDataPrompt still contains unreplaced account placeholders")
		}
		if strings.Contains(result.AnalysisDataPrompt, "{performance.") {
			t.Error("AnalysisDataPrompt still contains unreplaced performance placeholders")
		}
	})

	t.Run("包含预期的数据值", func(t *testing.T) {
		// 检查账户数据（来自AccountProvider的默认值）
		if !strings.Contains(result.AnalysisDataPrompt, "1000.00") {
			t.Error("AnalysisDataPrompt should contain account equity value")
		}

		// 检查performance数据
		if !strings.Contains(result.AnalysisDataPrompt, "1.50") {
			t.Error("AnalysisDataPrompt should contain sharpe ratio value")
		}
	})

	t.Run("Metadata包含构建信息", func(t *testing.T) {
		if result.Metadata == nil {
			t.Error("Metadata should not be nil")
		}
		if _, ok := result.Metadata["build_time_ms"]; !ok {
			t.Error("Metadata should contain build_time_ms")
		}
		if _, ok := result.Metadata["placeholders_count"]; !ok {
			t.Error("Metadata should contain placeholders_count")
		}
	})

	// 打印结果供调试
	t.Logf("SystemConstraintPrompt length: %d", len(result.SystemConstraintPrompt))
	t.Logf("AnalysisDataPrompt length: %d", len(result.AnalysisDataPrompt))
	t.Logf("Metadata: %+v", result.Metadata)
}

// TestProvider_Fetch 测试单个Provider的Fetch功能
func TestProvider_Fetch(t *testing.T) {
	ctx := context.Background()
	userID := "test-user"

	t.Run("ConfigProvider", func(t *testing.T) {
		p := NewConfigProvider(userID)

		// 测试altcoin_min_position
		req := &orchestrator.FieldRequest{Field: "altcoin_min_position"}
		val, err := p.Fetch(ctx, req)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		if val == nil {
			t.Error("Value should not be nil")
		}
		t.Logf("altcoin_min_position: %v", val)
	})

	t.Run("SystemProvider", func(t *testing.T) {
		p := NewSystemProvider(userID)

		// 测试current_time
		req := &orchestrator.FieldRequest{Field: "current_time"}
		val, err := p.Fetch(ctx, req)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		if val == nil || val == "" {
			t.Error("current_time should not be empty")
		}
		t.Logf("current_time: %v", val)
	})

	t.Run("AccountProvider", func(t *testing.T) {
		p := NewAccountProvider(userID)

		// 测试total_equity
		req := &orchestrator.FieldRequest{Field: "total_equity"}
		val, err := p.Fetch(ctx, req)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		if val != "1000.00" {
			t.Errorf("Expected 1000.00, got %v", val)
		}
	})

	t.Run("MarketProvider_with_symbol", func(t *testing.T) {
		p := NewMarketProvider(userID)

		// 测试price (需要symbol参数)
		req := &orchestrator.FieldRequest{
			Field:  "price",
			Params: map[string]string{"symbol": "BTCUSDT"},
		}
		val, err := p.Fetch(ctx, req)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		t.Logf("BTCUSDT price: %v", val)
	})

	t.Run("MarketProvider_without_symbol", func(t *testing.T) {
		p := NewMarketProvider(userID)

		// 测试没有symbol参数时应该报错
		req := &orchestrator.FieldRequest{Field: "price"}
		_, err := p.Fetch(ctx, req)
		if err == nil {
			t.Error("Should return error when symbol is missing")
		}
	})

	t.Run("PositionProvider", func(t *testing.T) {
		p := NewPositionProvider(userID)

		// 测试list
		req := &orchestrator.FieldRequest{Field: "list"}
		val, err := p.Fetch(ctx, req)
		if err != nil {
			t.Fatalf("Fetch failed: %v", err)
		}
		// 默认返回空持仓
		if val != "当前无持仓" {
			t.Errorf("Expected '当前无持仓', got %v", val)
		}
	})

	t.Run("UnsupportedField", func(t *testing.T) {
		p := NewConfigProvider(userID)

		req := &orchestrator.FieldRequest{Field: "unknown_field"}
		_, err := p.Fetch(ctx, req)
		if err == nil {
			t.Error("Should return error for unsupported field")
		}
	})
}

// TestProvider_FetchBatch 测试Provider的批量获取功能
func TestProvider_FetchBatch(t *testing.T) {
	ctx := context.Background()
	userID := "test-user"

	t.Run("AccountProvider_FetchBatch", func(t *testing.T) {
		p := NewAccountProvider(userID)

		requests := []*orchestrator.BatchFieldRequest{
			{
				FieldRequest: &orchestrator.FieldRequest{Field: "total_equity"},
				Raw:          "{account.total_equity}",
			},
			{
				FieldRequest: &orchestrator.FieldRequest{Field: "available_balance"},
				Raw:          "{account.available_balance}",
			},
			{
				FieldRequest: &orchestrator.FieldRequest{Field: "position_count"},
				Raw:          "{account.position_count}",
			},
		}

		result, err := p.FetchBatch(ctx, requests)
		if err != nil {
			t.Fatalf("FetchBatch failed: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 results, got %d", len(result))
		}

		// 验证结果
		if result["{account.total_equity}"] != "1000.00" {
			t.Errorf("Unexpected total_equity: %v", result["{account.total_equity}"])
		}
		if result["{account.available_balance}"] != "800.00" {
			t.Errorf("Unexpected available_balance: %v", result["{account.available_balance}"])
		}

		t.Logf("FetchBatch result: %+v", result)
	})

	t.Run("MarketProvider_FetchBatch_with_different_symbols", func(t *testing.T) {
		p := NewMarketProvider(userID)

		requests := []*orchestrator.BatchFieldRequest{
			{
				FieldRequest: &orchestrator.FieldRequest{
					Field:  "price",
					Params: map[string]string{"symbol": "BTCUSDT"},
				},
				Raw: "{market.BTCUSDT.price}",
			},
			{
				FieldRequest: &orchestrator.FieldRequest{
					Field:  "rsi",
					Params: map[string]string{"symbol": "BTCUSDT"},
				},
				Raw: "{market.BTCUSDT.rsi}",
			},
		}

		result, err := p.FetchBatch(ctx, requests)
		if err != nil {
			t.Fatalf("FetchBatch failed: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}

		t.Logf("MarketProvider FetchBatch result: %+v", result)
	})
}

// TestSection_GetContent 测试Section的内容获取
func TestSection_GetContent(t *testing.T) {
	t.Run("RiskControlSection", func(t *testing.T) {
		s := sections.NewRiskControlSection()
		content := s.GetContent()

		if content == "" {
			t.Error("Content should not be empty")
		}
		if s.GetName() != "risk_control" {
			t.Errorf("Expected name 'risk_control', got '%s'", s.GetName())
		}

		// 检查包含占位符
		if !strings.Contains(content, "{config.") {
			t.Error("RiskControlSection should contain config placeholders")
		}
	})

	t.Run("SystemStatusSection", func(t *testing.T) {
		s := sections.NewSystemStatusSection()
		content := s.GetContent()

		if !strings.Contains(content, "{system.current_time}") {
			t.Error("SystemStatusSection should contain {system.current_time}")
		}
	})

	t.Run("AccountSection", func(t *testing.T) {
		s := sections.NewAccountSection()
		content := s.GetContent()

		if !strings.Contains(content, "{account.") {
			t.Error("AccountSection should contain account placeholders")
		}
	})

	t.Run("SetContent", func(t *testing.T) {
		s := sections.NewRiskControlSection()
		customContent := "Custom content with {config.test}"
		s.SetContent(customContent)

		if s.GetContent() != customContent {
			t.Error("SetContent should update the content")
		}
	})
}

// TestOrchestrator_ProviderRegistration 测试Provider注册
func TestOrchestrator_ProviderRegistration(t *testing.T) {
	orch := core.New()

	t.Run("RegisterProvider_Success", func(t *testing.T) {
		p := NewConfigProvider("test")
		err := orch.RegisterProvider(p)
		if err != nil {
			t.Errorf("RegisterProvider failed: %v", err)
		}
	})

	t.Run("RegisterProvider_Duplicate", func(t *testing.T) {
		// 尝试重复注册
		p := NewConfigProvider("test")
		err := orch.RegisterProvider(p)
		if err == nil {
			t.Error("Should return error for duplicate registration")
		}
	})

	t.Run("RegisterProvider_Nil", func(t *testing.T) {
		err := orch.RegisterProvider(nil)
		if err == nil {
			t.Error("Should return error for nil provider")
		}
	})
}

// TestOrchestrator_BuildPrompt_WithMarketData 测试包含市场数据的完整流程
func TestOrchestrator_BuildPrompt_WithMarketData(t *testing.T) {
	ctx := context.Background()
	orch := core.New()

	// 注册providers
	orch.RegisterProvider(NewMarketProvider("test"))
	orch.RegisterProvider(NewSystemProvider("test"))

	// 创建包含市场占位符的section
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{},
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "market_info",
				Template: "BTC价格: ${market.BTCUSDT.price}\nRSI: {market.BTCUSDT.rsi}",
			},
			{
				Name:     "system_info",
				Template: "当前时间: {system.current_time}",
			},
		},
	}

	result, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证市场数据被替换
	if strings.Contains(result.AnalysisDataPrompt, "{market.") {
		t.Error("Market placeholders should be replaced")
	}

	// 验证包含具体值
	if !strings.Contains(result.AnalysisDataPrompt, "50000.00") {
		t.Error("Should contain BTC price value")
	}

	t.Logf("AnalysisDataPrompt:\n%s", result.AnalysisDataPrompt)
}

// TestFullPromptGeneration 完整的端到端测试
// 验证所有占位符都被正确替换为期望的值
func TestFullPromptGeneration(t *testing.T) {
	ctx := context.Background()

	// 1. 创建Orchestrator
	orch := core.New()

	// 2. 注册所有Providers
	userID := "test-user-001"
	providers := []orchestrator.Provider{
		NewConfigProvider(userID),
		NewSystemProvider(userID),
		NewAccountProvider(userID),
		NewMarketProvider(userID),
		NewPositionProvider(userID),
		NewCandidateProvider(userID),
		NewPerformanceProvider(userID),
	}

	for _, p := range providers {
		if err := orch.RegisterProvider(p); err != nil {
			t.Fatalf("Failed to register provider %s: %v", p.Name(), err)
		}
	}

	// 3. 创建自定义模板进行精确测试
	systemTemplate := `## 系统约束

### 仓位限制
- 山寨币仓位: {config.altcoin_min_position} - {config.altcoin_max_position} USDT
- BTC/ETH仓位: {config.btc_eth_min_position} - {config.btc_eth_max_position} USDT

### 杠杆限制
- 山寨币杠杆: {config.altcoin_leverage}x
- BTC/ETH杠杆: {config.btc_eth_leverage}x

### 最小开仓
- 一般币种: {config.min_position_size_general} USDT
- BTC/ETH: {config.min_position_size_btc_eth} USDT`

	dataTemplate := `## 实时数据

### 系统状态
- 当前时间: {system.current_time}
- 调用周期: {system.call_count}
- 运行时长: {system.runtime_minutes} 分钟

### 账户信息
- 账户净值: {account.total_equity} USDT
- 可用余额: {account.available_balance} USDT ({account.available_balance_pct}%)
- 总盈亏: {account.total_pnl_pct}%
- 保证金使用: {account.margin_used_pct}%
- 持仓数量: {account.position_count}

### BTC市场
- 价格: ${market.BTCUSDT.price}
- 1h涨跌: {market.BTCUSDT.change_1h}%
- 4h涨跌: {market.BTCUSDT.change_4h}%
- RSI: {market.BTCUSDT.rsi}
- MACD: {market.BTCUSDT.macd}

### 当前持仓
{position.list}

### 候选币种
共 {candidate.count} 个候选
{candidate.list}

### 历史表现
- 夏普比率: {performance.sharpe_ratio}`

	// 4. 构建PromptBuilder
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{Name: "system_constraint", Template: systemTemplate},
		},
		AnalysisDataSections: []*orchestrator.Section{
			{Name: "analysis_data", Template: dataTemplate},
		},
	}

	// 5. 构建Prompt
	result, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 6. 验证SystemConstraintPrompt中的值
	t.Run("验证系统约束内容", func(t *testing.T) {
		prompt := result.SystemConstraintPrompt

		// 验证仓位限制（基于AccountEquity=1000的默认值）
		expectations := map[string]string{
			"800":  "altcoin_min_position (1000*0.8)",
			"1500": "altcoin_max_position (1000*1.5)",
			"5000": "btc_eth_min_position (1000*5)",
			"10000": "btc_eth_max_position (1000*10)",
			"10x":   "altcoin_leverage",
			"20x":   "btc_eth_leverage",
			"12 USDT": "min_position_size_general",
			"60 USDT": "min_position_size_btc_eth",
		}

		for expected, desc := range expectations {
			if !strings.Contains(prompt, expected) {
				t.Errorf("SystemConstraintPrompt应包含 %s (%s)", expected, desc)
			}
		}

		// 验证没有未替换的占位符
		if strings.Contains(prompt, "{config.") {
			t.Error("SystemConstraintPrompt仍包含未替换的config占位符")
		}

		t.Logf("SystemConstraintPrompt:\n%s", prompt)
	})

	// 7. 验证AnalysisDataPrompt中的值
	t.Run("验证数据分析内容", func(t *testing.T) {
		prompt := result.AnalysisDataPrompt

		// 验证账户数据（来自AccountProvider默认值）
		accountExpectations := map[string]string{
			"1000.00":  "total_equity",
			"800.00":   "available_balance",
			"80.0":     "available_balance_pct",
			"5.50":     "total_pnl_pct",
			"20.0":     "margin_used_pct",
		}

		for expected, desc := range accountExpectations {
			if !strings.Contains(prompt, expected) {
				t.Errorf("AnalysisDataPrompt应包含 %s (%s)", expected, desc)
			}
		}

		// 验证市场数据（来自MarketProvider默认值）
		marketExpectations := map[string]string{
			"50000.00": "BTCUSDT price",
			"1.50":     "change_1h",
			"3.20":     "change_4h",
			"55.00":    "rsi",
			"0.0012":   "macd",
		}

		for expected, desc := range marketExpectations {
			if !strings.Contains(prompt, expected) {
				t.Errorf("AnalysisDataPrompt应包含 %s (%s)", expected, desc)
			}
		}

		// 验证performance数据
		if !strings.Contains(prompt, "1.50") {
			t.Error("AnalysisDataPrompt应包含夏普比率 1.50")
		}

		// 验证持仓和候选（默认为空）
		if !strings.Contains(prompt, "当前无持仓") {
			t.Error("AnalysisDataPrompt应包含'当前无持仓'")
		}
		if !strings.Contains(prompt, "当前无候选币种") {
			t.Error("AnalysisDataPrompt应包含'当前无候选币种'")
		}

		// 验证没有未替换的占位符
		placeholderPrefixes := []string{"{system.", "{account.", "{market.", "{position.", "{candidate.", "{performance."}
		for _, prefix := range placeholderPrefixes {
			if strings.Contains(prompt, prefix) {
				t.Errorf("AnalysisDataPrompt仍包含未替换的占位符: %s", prefix)
			}
		}

		t.Logf("AnalysisDataPrompt:\n%s", prompt)
	})

	// 8. 验证Metadata
	t.Run("验证元数据", func(t *testing.T) {
		if result.Metadata == nil {
			t.Fatal("Metadata不应为nil")
		}

		// 检查占位符数量
		count, ok := result.Metadata["placeholders_count"].(int)
		if !ok {
			t.Error("placeholders_count应为int类型")
		} else if count == 0 {
			t.Error("placeholders_count不应为0")
		} else {
			t.Logf("占位符数量: %d", count)
		}

		// 检查使用的provider数量
		providerCount, ok := result.Metadata["providers_used"].(int)
		if !ok {
			t.Error("providers_used应为int类型")
		} else if providerCount < 5 {
			t.Errorf("providers_used应至少为5，实际为%d", providerCount)
		} else {
			t.Logf("使用的Provider数量: %d", providerCount)
		}

		t.Logf("Metadata: %+v", result.Metadata)
	})
}

// TestFullPromptGeneration_WithJSONContent 测试包含JSON内容的模板
func TestFullPromptGeneration_WithJSONContent(t *testing.T) {
	ctx := context.Background()
	orch := core.New()

	// 注册config provider
	orch.RegisterProvider(NewConfigProvider("test"))

	// 包含JSON示例的模板（模拟OutputFormatSection）
	templateWithJSON := `## 输出格式

请以JSON格式输出，示例：
` + "```json" + `
[
  {"symbol": "BTCUSDT", "action": "open_long", "leverage": 20, "position_size": 5000},
  {"symbol": "ETHUSDT", "action": "close_short", "reasoning": "止盈离场"}
]
` + "```" + `

### 配置参数
- 杠杆: {config.btc_eth_leverage}x
- 最小仓位: {config.min_position_size_btc_eth} USDT`

	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{Name: "output_format", Template: templateWithJSON},
		},
	}

	result, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	prompt := result.SystemConstraintPrompt

	// 验证JSON内容保持不变
	if !strings.Contains(prompt, `"symbol": "BTCUSDT"`) {
		t.Error("JSON内容应保持不变")
	}
	if !strings.Contains(prompt, `"action": "open_long"`) {
		t.Error("JSON内容应保持不变")
	}

	// 验证占位符被正确替换
	if !strings.Contains(prompt, "20x") {
		t.Error("应包含替换后的杠杆值 20x")
	}
	if !strings.Contains(prompt, "60 USDT") {
		t.Error("应包含替换后的最小仓位值 60 USDT")
	}

	// 验证没有未替换的占位符
	if strings.Contains(prompt, "{config.") {
		t.Error("仍包含未替换的config占位符")
	}

	t.Logf("包含JSON的Prompt:\n%s", prompt)
}

// BenchmarkFetchBatch 性能测试
func BenchmarkFetchBatch(b *testing.B) {
	ctx := context.Background()
	p := NewAccountProvider("bench-user")

	requests := []*orchestrator.BatchFieldRequest{
		{FieldRequest: &orchestrator.FieldRequest{Field: "total_equity"}, Raw: "{account.total_equity}"},
		{FieldRequest: &orchestrator.FieldRequest{Field: "available_balance"}, Raw: "{account.available_balance}"},
		{FieldRequest: &orchestrator.FieldRequest{Field: "total_pnl_pct"}, Raw: "{account.total_pnl_pct}"},
		{FieldRequest: &orchestrator.FieldRequest{Field: "margin_used_pct"}, Raw: "{account.margin_used_pct}"},
		{FieldRequest: &orchestrator.FieldRequest{Field: "position_count"}, Raw: "{account.position_count}"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.FetchBatch(ctx, requests)
		if err != nil {
			b.Fatal(err)
		}
	}
}
