# Prompt Sections 模块使用说明

## 概述

这个模块包含所有 Prompt Section 的实现，用于等价替代 `decision/engine.go` 中的硬编码Prompt生成逻辑。

## 架构设计

**核心思想**: Section = 内容（Content） + 名称（Name）

- 每个Section是一个独立的Go结构体
- 包含默认content（在NewXxxSection()中硬编码）
- 支持SetContent()动态修改
- 支持LoadTemplate(id)从数据库/文件加载

## Section列表

### SystemConstraint 部分（System Prompt）

用于构建发送给AI的system prompt，定义交易规则和输出格式：

| Section | 文件 | 说明 | 对应engine.go |
|---------|------|------|-------------|
| **base_strategy** | `base_strategy_section.go` | 基础交易策略（从prompts/加载） | 行313-334 |
| **risk_control** | `risk_control_section.go` | 硬约束风险控制 | 行336-344 |
| **output_format** | `output_format_section.go` | 输出格式要求 | 行346-367 |
| **custom_strategy** | `custom_section.go` | 个性化策略（可选） | buildSystemPromptWithCustom |

### AnalysisData 部分（User Prompt）

用于构建发送给AI的user prompt，包含实时交易数据：

| Section | 文件 | 说明 | 对应engine.go |
|---------|------|------|-------------|
| **system_status** | `system_status_section.go` | 系统状态（时间、周期） | 行376-378 |
| **btc_market** | `btc_market_section.go` | BTC市场概况 | 行380-385 |
| **account_summary** | `account_section.go` | 账户摘要 | 行387-394 |
| **current_positions** | `positions_section.go` | 当前持仓详情 | 行396-430 |
| **candidate_coins** | `candidate_coins_section.go` | 候选币种详情 | 行432-454 |
| **performance_metrics** | `performance_metrics_section.go` | 夏普比率（可选） | 行456-468 |

## 使用示例

```go
package main

import (
	"nofx/prompt/sections"
)

func main() {
	// 1. 创建SystemConstraint sections
	baseStrategy := sections.NewBaseStrategySection("default")
	riskControl := sections.NewRiskControlSection()
	outputFormat := sections.NewOutputFormatSection()
	customStrategy := sections.NewCustomStrategySection() // 可选

	// 2. 创建AnalysisData sections
	systemStatus := sections.NewSystemStatusSection()
	btcMarket := sections.NewBTCMarketSection()
	account := sections.NewAccountSection()
	positions := sections.NewPositionsSection()
	candidates := sections.NewCandidateCoinsSection()
	performance := sections.NewPerformanceMetricsSection() // 可选

	// 3. 组装System Prompt
	systemPrompt := baseStrategy.GetContent() + "\n\n" +
		riskControl.GetContent() +
		outputFormat.GetContent()

	// 如果有自定义策略，添加到system prompt
	if hasCustom {
		systemPrompt += customStrategy.GetContent()
	}

	// 4. 组装User Prompt
	userPrompt := systemStatus.GetContent() +
		btcMarket.GetContent() +
		account.GetContent() +
		positions.GetContent() +
		candidates.GetContent()

	// 如果有历史表现数据，添加到user prompt
	if hasPerformance {
		userPrompt += performance.GetContent()
	}

	// 5. 使用Prompts调用AI
	aiResponse := callAI(systemPrompt, userPrompt)
}
```

## 占位符说明

所有Section模板中的占位符采用 `{provider.field}` 格式：

### config Provider（8个字段）
- `{config.altcoin_min_position}` - 山寨币最小仓位
- `{config.altcoin_max_position}` - 山寨币最大仓位
- `{config.btc_eth_min_position}` - BTC/ETH最小仓位
- `{config.btc_eth_max_position}` - BTC/ETH最大仓位
- `{config.altcoin_leverage}` - 山寨币杠杆倍数
- `{config.btc_eth_leverage}` - BTC/ETH杠杆倍数
- `{config.min_position_size_general}` - 一般币种最小开仓金额
- `{config.example_btc_position}` - BTC示例仓位（用于output_format）

### system Provider（3个字段）
- `{system.current_time}` - 当前时间
- `{system.call_count}` - 调用周期数
- `{system.runtime_minutes}` - 运行时长（分钟）

### market Provider（5个字段，支持symbol参数）
- `{market.BTCUSDT.price}` - BTC当前价格
- `{market.BTCUSDT.change_1h}` - 1小时涨跌幅
- `{market.BTCUSDT.change_4h}` - 4小时涨跌幅
- `{market.BTCUSDT.macd}` - MACD值
- `{market.BTCUSDT.rsi}` - RSI值

### account Provider（6个字段）
- `{account.total_equity}` - 账户净值
- `{account.available_balance}` - 可用余额
- `{account.available_balance_pct}` - 可用余额百分比
- `{account.total_pnl_pct}` - 总盈亏百分比
- `{account.margin_used_pct}` - 保证金使用率
- `{account.position_count}` - 持仓数量

### position Provider（1个字段）
- `{position.list}` - 持仓列表（完整格式化字符串）

### candidate Provider（2个字段）
- `{candidate.count}` - 候选币种数量
- `{candidate.list}` - 候选币种列表（完整格式化字符串）

### performance Provider（1个字段，可选）
- `{performance.sharpe_ratio}` - 夏普比率

### custom Provider（1个字段，可选）
- `{custom.user_prompt}` - 用户自定义策略文本

## 未来扩展

### 1. SetContent() - 运行时动态修改内容

```go
section := sections.NewRiskControlSection()
section.SetContent(`自定义风险控制内容...`)
```

### 2. LoadTemplate(id) - 从数据库/文件加载模板

```go
section := sections.NewRiskControlSection()
section.LoadTemplate(12345) // 从数据库加载ID为12345的模板
```

### 3. default/ 目录结构（预留）

```
default/
├── base_sections/              # 基础策略模板
├── data_analysis_sections/     # 数据分析模板
├── output_format_sections/     # 输出格式模板
└── risk_control_sections/      # 风控规则模板
```

## 与decision/engine.go的对应关系

| engine.go函数 | 对应Sections |
|--------------|-------------|
| `buildSystemPrompt()` | base_strategy + risk_control + output_format |
| `buildSystemPromptWithCustom()` | base_strategy + risk_control + output_format + custom_strategy |
| `buildUserPrompt()` | system_status + btc_market + account_summary + current_positions + candidate_coins + performance_metrics |

## 优势

1. **解耦**: Section与业务逻辑分离，易于维护
2. **灵活**: 支持运行时修改模板
3. **可测试**: 每个Section可独立测试
4. **可复用**: 不同策略可组合使用相同的Sections
5. **类型安全**: Go编译时检查
6. **零IO**: 默认模板编译到二进制（可选文件加载）

## 测试

```bash
# 运行所有tests
go test nofx/prompt/sections -v

# 测试特定section
go test nofx/prompt/sections -run TestRiskControl -v
```
