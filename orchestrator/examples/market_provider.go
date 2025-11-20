package examples

import (
	"context"
	"fmt"
	"strings"

	orchestrator "nofx/orchestrator"
)

// ============================================================
// MarketProvider 示例实现
// 展示如何处理带参数的占位符
// ============================================================

// MarketProvider 市场数据Provider
type MarketProvider struct {
	// 可以注入实际的市场数据模块
	// marketModule *market.Module
}

// NewMarketProvider 创建MarketProvider
func NewMarketProvider() *MarketProvider {
	return &MarketProvider{}
}

// Name 返回Provider名称
func (p *MarketProvider) Name() string {
	return "market"
}

// Fetch 获取字段值
func (p *MarketProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	// 根据字段路径和参数路由到不同的处理器
	// Field示例:
	//   "BTCUSDT.price"           → 实时价格
	//   "1h.kline.close"          → K线数据
	//   "1h.BTCUSDT.kline.close"  → 指定symbol的K线数据

	// 提取参数
	interval, hasInterval := request.Params["interval"]
	symbol, hasSymbol := request.Params["symbol"]

	// 判断字段类型
	field := request.Field

	// 场景1: 实时价格 {market.BTCUSDT.price}
	if strings.HasSuffix(field, ".price") && hasSymbol {
		return p.getCurrentPrice(ctx, symbol)
	}

	// 场景2: K线数据 {market.1h.BTCUSDT.kline.close}
	if strings.Contains(field, "kline") && hasInterval {
		return p.getKlineData(ctx, interval, symbol, field)
	}

	// 场景3: 技术指标 {market.BTCUSDT.rsi_14}
	if hasSymbol && p.isTechnicalIndicator(field) {
		return p.getTechnicalIndicator(ctx, symbol, field)
	}

	// 场景4: 24h统计 {market.BTCUSDT.change_24h}
	if hasSymbol && strings.HasSuffix(field, ".change_24h") {
		return p.get24hChange(ctx, symbol)
	}

	return nil, fmt.Errorf("unsupported field pattern: %s", field)
}

// FetchBatch 批量获取字段值
// 实现BatchProvider接口，Provider可以在这里优化批量查询
func (p *MarketProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 这里可以优化：
	// 1. 先分析所有请求，按类型分组（价格、K线、指标等）
	// 2. 批量调用API（比如一次性获取多个symbol的价格）
	// 3. 然后填充到结果中
	//
	// 示例优化：
	// - 收集所有需要价格的symbols，一次性批量获取
	// - 收集所有需要K线的(symbol, interval)组合，批量获取
	//
	// 目前的简单实现：复用Fetch方法（但实际应用应该优化）
	for _, req := range requests {
		value, err := p.Fetch(ctx, req.FieldRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", req.Raw, err)
		}

		result[req.Raw] = value
	}

	return result, nil
}

// ============================================================
// 字段处理器实现
// ============================================================

// getCurrentPrice 获取实时价格
// 占位符: {market.BTCUSDT.price}
// Params: {"symbol": "BTCUSDT"}
func (p *MarketProvider) getCurrentPrice(ctx context.Context, symbol string) (interface{}, error) {
	// 调用实际的市场数据模块
	// data := p.marketModule.GetTicker(symbol)
	// return data.LastPrice, nil

	// 示例数据
	prices := map[string]float64{
		"BTCUSDT": 51234.50,
		"ETHUSDT": 3456.78,
	}

	price, ok := prices[symbol]
	if !ok {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	return price, nil
}

// getKlineData 获取K线数据
// 占位符: {market.1h.BTCUSDT.kline.close}
// Params: {"interval": "1h", "symbol": "BTCUSDT"}
func (p *MarketProvider) getKlineData(ctx context.Context, interval, symbol, field string) (interface{}, error) {
	// 调用实际的市场数据模块
	// klines := p.marketModule.GetKlines(symbol, interval, 1)
	// if len(klines) == 0 {
	//     return nil, fmt.Errorf("no kline data")
	// }
	// kline := klines[0]

	// 根据field返回对应字段
	// field可能是: "1h.BTCUSDT.kline.close" 或 "1h.BTCUSDT.kline.open"
	if strings.HasSuffix(field, "close") {
		return 51200.00, nil
	}
	if strings.HasSuffix(field, "open") {
		return 51100.00, nil
	}
	if strings.HasSuffix(field, "high") {
		return 51300.00, nil
	}
	if strings.HasSuffix(field, "low") {
		return 51000.00, nil
	}
	if strings.HasSuffix(field, "volume") {
		return 12345.67, nil
	}

	return nil, fmt.Errorf("unsupported kline field: %s", field)
}

// getTechnicalIndicator 获取技术指标
// 占位符: {market.BTCUSDT.rsi_14}
// Params: {"symbol": "BTCUSDT"}
func (p *MarketProvider) getTechnicalIndicator(ctx context.Context, symbol, field string) (interface{}, error) {
	// field示例: "BTCUSDT.rsi_14", "BTCUSDT.macd_signal"

	// 提取指标名称
	parts := strings.Split(field, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid field format: %s", field)
	}
	indicatorName := parts[1] // rsi_14, macd_signal等

	// 调用实际的技术指标模块
	// indicator := p.marketModule.GetIndicator(symbol, indicatorName)
	// return indicator.Value, nil

	// 示例数据
	indicators := map[string]float64{
		"rsi_14":      65.5,
		"macd_signal": 0.12,
		"ema_20":      51150.0,
	}

	value, ok := indicators[indicatorName]
	if !ok {
		return nil, fmt.Errorf("indicator not found: %s", indicatorName)
	}

	return value, nil
}

// get24hChange 获取24h涨跌幅
// 占位符: {market.BTCUSDT.change_24h}
// Params: {"symbol": "BTCUSDT"}
func (p *MarketProvider) get24hChange(ctx context.Context, symbol string) (interface{}, error) {
	// 调用实际的市场数据模块
	// ticker := p.marketModule.Get24hTicker(symbol)
	// return ticker.PriceChangePercent, nil

	// 示例数据
	changes := map[string]float64{
		"BTCUSDT": 2.35,
		"ETHUSDT": -1.28,
	}

	change, ok := changes[symbol]
	if !ok {
		return nil, fmt.Errorf("symbol not found: %s", symbol)
	}

	return change, nil
}

// isTechnicalIndicator 判断是否是技术指标字段
func (p *MarketProvider) isTechnicalIndicator(field string) bool {
	indicators := []string{"rsi", "macd", "ema", "sma", "boll", "kdj"}
	for _, indicator := range indicators {
		if strings.Contains(field, indicator) {
			return true
		}
	}
	return false
}
