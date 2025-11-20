package data_providers

import (
	"context"
	"fmt"
	"strings"

	"nofx/orchestrator"
)

// MarketProvider 市场数据Provider
// 提供币种市场数据（价格、涨跌幅、技术指标）
// 支持通过Params["symbol"]指定币种
type MarketProvider struct {
	userID string
	// TODO: 添加市场数据服务、数据库连接等依赖
}

// 字段处理器类型（需要symbol参数）
type marketFieldHandler func(*MarketProvider, context.Context, string) (interface{}, error)

// 字段到方法的映射
var marketFieldHandlers = map[string]marketFieldHandler{
	"price":     (*MarketProvider).fetchPrice,
	"change_1h": (*MarketProvider).fetchChange1h,
	"change_4h": (*MarketProvider).fetchChange4h,
	"macd":      (*MarketProvider).fetchMACD,
	"rsi":       (*MarketProvider).fetchRSI,
}

// NewMarketProvider 创建MarketProvider
func NewMarketProvider(userID string) *MarketProvider {
	return &MarketProvider{
		userID: userID,
	}
}

// Name 返回Provider名称
func (p *MarketProvider) Name() string {
	return "market"
}

// Fetch 获取字段值
// 需要通过request.Params["symbol"]指定币种
func (p *MarketProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	// 从Params获取symbol
	symbol, ok := request.Params["symbol"]
	if !ok || symbol == "" {
		return nil, fmt.Errorf("market: symbol parameter is required")
	}

	// 从Field中提取实际字段名
	// Field可能是 "price" 或 "BTCUSDT.price"
	field := request.Field
	if idx := strings.LastIndex(field, "."); idx != -1 {
		field = field[idx+1:]
	}

	// 查找字段处理器
	handler, ok := marketFieldHandlers[field]
	if !ok {
		return nil, fmt.Errorf("market: unsupported field: %s", request.Field)
	}

	return handler(p, ctx, symbol)
}

// FetchBatch 批量获取字段值（并发）
func (p *MarketProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现

func (p *MarketProvider) fetchPrice(ctx context.Context, symbol string) (interface{}, error) {
	price, err := p.getPrice(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", price), nil
}

func (p *MarketProvider) fetchChange1h(ctx context.Context, symbol string) (interface{}, error) {
	change, err := p.getChange1h(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", change), nil
}

func (p *MarketProvider) fetchChange4h(ctx context.Context, symbol string) (interface{}, error) {
	change, err := p.getChange4h(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", change), nil
}

func (p *MarketProvider) fetchMACD(ctx context.Context, symbol string) (interface{}, error) {
	macd, err := p.getMACD(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.4f", macd), nil
}

func (p *MarketProvider) fetchRSI(ctx context.Context, symbol string) (interface{}, error) {
	rsi, err := p.getRSI(ctx, symbol)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", rsi), nil
}

// 内部数据获取方法 - TODO: 实现真实数据获取逻辑

func (p *MarketProvider) getPrice(ctx context.Context, symbol string) (float64, error) {
	// TODO: 从市场数据服务获取价格
	return 50000.0, nil
}

func (p *MarketProvider) getChange1h(ctx context.Context, symbol string) (float64, error) {
	// TODO: 从市场数据服务获取1小时涨跌幅
	return 1.5, nil
}

func (p *MarketProvider) getChange4h(ctx context.Context, symbol string) (float64, error) {
	// TODO: 从市场数据服务获取4小时涨跌幅
	return 3.2, nil
}

func (p *MarketProvider) getMACD(ctx context.Context, symbol string) (float64, error) {
	// TODO: 从市场数据服务获取MACD
	return 0.0012, nil
}

func (p *MarketProvider) getRSI(ctx context.Context, symbol string) (float64, error) {
	// TODO: 从市场数据服务获取RSI
	return 55.0, nil
}
