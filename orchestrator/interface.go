package orchestrator

import (
	"context"
)

// ============================================================
// Orchestrator V2 - 占位符驱动架构
//
// 核心理念：
// - Section是纯模板（包含占位符）
// - Provider提供细粒度数据（单个字段值）
// - Orchestrator统一解析和替换占位符
// ============================================================

// Orchestrator 编排器接口
type Orchestrator interface {
	// BuildPrompt 构建Prompt
	// 使用PromptBuilder来明确指定每个部分的顺序
	BuildPrompt(ctx context.Context, builder *PromptBuilder) (*BuiltPrompt, error)

	// RegisterProvider 注册Provider
	RegisterProvider(provider Provider) error
}

// PromptBuilder Prompt构建器
// 通过数组顺序明确指定最终Prompt的组装顺序
type PromptBuilder struct {
	// SystemConstraintSections 系统约束部分的Sections
	// 数组顺序就是最终SystemConstraintPrompt的组装顺序
	// 例如: [基础模板, 风控规则, 输出格式] 就是最终顺序
	SystemConstraintSections []*Section

	// AnalysisDataSections 数据分析部分的Sections
	// 数组顺序就是最终AnalysisDataPrompt的组装顺序
	// 例如: [账户信息, 市场数据, 持仓信息] 就是最终顺序
	AnalysisDataSections []*Section
}

// Section 纯模板Section（不包含Fill逻辑）
type Section struct {
	Name     string // Section名称
	Template string // 纯模板字符串（包含占位符）
}

// BuiltPrompt 构建好的Prompt
type BuiltPrompt struct {
	SystemConstraintPrompt string                 // 系统约束提示词（包括基础模板、风控规则、输出格式等）
	AnalysisDataPrompt     string                 // 分析数据提示词（包括市场数据、账户信息、持仓等）
	Metadata               map[string]interface{} // 元数据
}

// Provider 数据提供者接口（V2版本：细粒度）
type Provider interface {
	// Name 返回Provider唯一名称
	Name() string

	// Fetch 获取单个字段的值
	// ctx: 上下文（包含用户信息，如trader_id）
	// request: 字段请求（包含字段路径和参数）
	Fetch(ctx context.Context, request *FieldRequest) (interface{}, error)
}

// BatchProvider 批量数据提供者接口
// Provider可以选择实现此接口来优化批量数据获取
type BatchProvider interface {
	Provider // 嵌入基础接口

	// FetchBatch 批量获取多个字段的值
	// ctx: 上下文（包含用户信息）
	// requests: 批量字段请求
	// 返回: map[占位符原始字符串]值
	//
	// 示例:
	//   requests: [
	//     {Field: "available_balance", ...},
	//     {Field: "equity", ...}
	//   ]
	//   返回: {
	//     "{account.available_balance}": 10000.50,
	//     "{account.equity}": 9500.00
	//   }
	FetchBatch(ctx context.Context, requests []*BatchFieldRequest) (map[string]interface{}, error)
}

// FieldRequest 字段请求
type FieldRequest struct {
	// Field 字段路径
	// 示例:
	//   "available_balance"           → 简单字段
	//   "BTCUSDT.price"              → 带symbol参数
	//   "1h.kline.close"             → 带时间参数
	//   "1h.BTCUSDT.kline.close"     → 带多个参数
	Field string

	// Params 解析出的参数
	// 示例:
	//   "1h.BTCUSDT.kline.close" → {"interval": "1h", "symbol": "BTCUSDT"}
	Params map[string]string
}

// BatchFieldRequest 批量字段请求
type BatchFieldRequest struct {
	*FieldRequest        // 嵌入字段请求
	Raw           string // 原始占位符（用于返回时的key）
}

// ============================================================
// 占位符格式
// ============================================================
//
// 基本格式: {provider.field}
// 带参数格式: {provider.param1.param2.field}
//
// 示例:
//   {account.available_balance}          → AccountProvider, field="available_balance"
//   {market.BTCUSDT.price}               → MarketProvider, field="BTCUSDT.price", params={"symbol": "BTCUSDT"}
//   {market.1h.kline.close}              → MarketProvider, field="1h.kline.close", params={"interval": "1h"}
//   {market.1h.BTCUSDT.kline.close}      → MarketProvider, field="1h.BTCUSDT.kline.close", params={"interval": "1h", "symbol": "BTCUSDT"}
//   {config.strategy_type}               → ConfigProvider, field="strategy_type"
