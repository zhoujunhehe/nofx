package examples

import (
	"context"
	"fmt"

	orchestrator "nofx/orchestrator"
	"nofx/orchestrator/core"
)

// ============================================================
// 完整使用示例
// ============================================================

// Example 展示如何使用orchestrator_v2
func Example() {
	// 1. 创建orchestrator
	orch := core.New()

	// 2. 注册Providers
	orch.RegisterProvider(NewAccountProvider())
	orch.RegisterProvider(NewMarketProvider())

	// 3. 使用PromptBuilder明确指定Section的组装顺序
	// 数组顺序就是最终Prompt的顺序，直接明了
	builder := &orchestrator.PromptBuilder{
		// SystemConstraintSections: 系统约束部分
		// 数组顺序: [基础模板, 风控规则] 就是最终顺序
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name: "template",
				Template: `你是一个专业的加密货币交易助手。
当前正在分析市场数据，为用户提供交易建议。`,
			},
			{
				Name: "risk_control",
				Template: `
## 风控规则
- 单次最大开仓金额: {account.available_balance} USDT的20%
- 可用保证金: {account.available_margin} USDT
- 当前持仓价值: {account.total_position_value} USDT`,
			},
		},

		// AnalysisDataSections: 数据分析部分
		// 数组顺序: [账户信息, 市场数据] 就是最终顺序
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name: "account",
				Template: `
## 账户信息
余额: {account.available_balance} USDT
净值: {account.equity} USDT
未实现盈亏: {account.unrealized_pnl} USDT
已用保证金: {account.margin_used} USDT`,
			},
			{
				Name: "market",
				Template: `
## 市场数据
### BTCUSDT
当前价格: {market.BTCUSDT.price}
24h涨跌: {market.BTCUSDT.change_24h}%
RSI(14): {market.BTCUSDT.rsi_14}

### 1小时K线
收盘价: {market.1h.BTCUSDT.kline.close}
开盘价: {market.1h.BTCUSDT.kline.open}
最高价: {market.1h.BTCUSDT.kline.high}
最低价: {market.1h.BTCUSDT.kline.low}`,
			},
		},
	}

	// 4. 构建Prompt（用户信息通过ctx传入）
	ctx := context.WithValue(context.Background(), "trader_id", "trader_001")
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 5. 输出结果
	fmt.Println("=== SystemConstraintPrompt ===")
	fmt.Println(prompt.SystemConstraintPrompt)
	fmt.Println()
	fmt.Println("=== AnalysisDataPrompt ===")
	fmt.Println(prompt.AnalysisDataPrompt)
	fmt.Println()
	fmt.Println("=== Metadata ===")
	fmt.Printf("Build time: %v ms\n", prompt.Metadata["build_time_ms"])
	fmt.Printf("Placeholders: %v\n", prompt.Metadata["placeholders_count"])
	fmt.Printf("Providers used: %v\n", prompt.Metadata["providers_used"])
}

// ExampleOutput 预期输出
/*
=== SystemConstraintPrompt ===
你是一个专业的加密货币交易助手。
当前正在分析市场数据，为用户提供交易建议。

## 风控规则
- 单次最大开仓金额: 10000.50 USDT的20%
- 可用保证金: 9000.00 USDT
- 当前持仓价值: 5000.00 USDT

=== AnalysisDataPrompt ===

## 账户信息
余额: 10000.50 USDT
净值: 9500.00 USDT
未实现盈亏: 150.25 USDT
已用保证金: 500.00 USDT

## 市场数据
### BTCUSDT
当前价格: 51234.50
24h涨跌: 2.35%
RSI(14): 65.50

### 1小时K线
收盘价: 51200.00
开盘价: 51100.00
最高价: 51300.00
最低价: 51000.00

=== Metadata ===
Build time: 15 ms
Placeholders: 13
Providers used: 2
*/
