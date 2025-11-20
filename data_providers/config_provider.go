package data_providers

import (
	"context"
	"fmt"

	"nofx/orchestrator"
)

// ConfigProvider 配置数据Provider
// 提供交易配置参数（杠杆、仓位限制等）
type ConfigProvider struct {
	userID string
	// TODO: 添加数据库连接、配置服务等依赖
}

// 字段处理器类型
type configFieldHandler func(*ConfigProvider, context.Context) (interface{}, error)

// 字段到方法的映射
var configFieldHandlers = map[string]configFieldHandler{
	"altcoin_min_position":      (*ConfigProvider).fetchAltcoinMinPosition,
	"altcoin_max_position":      (*ConfigProvider).fetchAltcoinMaxPosition,
	"btc_eth_min_position":      (*ConfigProvider).fetchBTCETHMinPosition,
	"btc_eth_max_position":      (*ConfigProvider).fetchBTCETHMaxPosition,
	"altcoin_leverage":          (*ConfigProvider).fetchAltcoinLeverage,
	"btc_eth_leverage":          (*ConfigProvider).fetchBTCETHLeverage,
	"min_position_size_general": (*ConfigProvider).fetchMinPositionSizeGeneral,
	"min_position_size_btc_eth": (*ConfigProvider).fetchMinPositionSizeBTCETH,
	"example_btc_position":      (*ConfigProvider).fetchExampleBTCPosition,
}

// NewConfigProvider 创建ConfigProvider
func NewConfigProvider(userID string) *ConfigProvider {
	return &ConfigProvider{
		userID: userID,
	}
}

// Name 返回Provider名称
func (p *ConfigProvider) Name() string {
	return "config"
}

// Fetch 获取字段值
func (p *ConfigProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	handler, ok := configFieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("config: unsupported field: %s", request.Field)
	}
	return handler(p, ctx)
}

// FetchBatch 批量获取字段值（并发）
func (p *ConfigProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现 - 内部直接获取数据

func (p *ConfigProvider) fetchAltcoinMinPosition(ctx context.Context) (interface{}, error) {
	// TODO: 从数据库/配置中心获取用户账户净值
	equity, err := p.getAccountEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.0f", equity*0.8), nil
}

func (p *ConfigProvider) fetchAltcoinMaxPosition(ctx context.Context) (interface{}, error) {
	equity, err := p.getAccountEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.0f", equity*1.5), nil
}

func (p *ConfigProvider) fetchBTCETHMinPosition(ctx context.Context) (interface{}, error) {
	equity, err := p.getAccountEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.0f", equity*5), nil
}

func (p *ConfigProvider) fetchBTCETHMaxPosition(ctx context.Context) (interface{}, error) {
	equity, err := p.getAccountEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.0f", equity*10), nil
}

func (p *ConfigProvider) fetchAltcoinLeverage(ctx context.Context) (interface{}, error) {
	// TODO: 从用户配置中获取
	return 10, nil
}

func (p *ConfigProvider) fetchBTCETHLeverage(ctx context.Context) (interface{}, error) {
	// TODO: 从用户配置中获取
	return 20, nil
}

func (p *ConfigProvider) fetchMinPositionSizeGeneral(ctx context.Context) (interface{}, error) {
	// 修改为 10 USDT：
	// - Binance: 官方建议 >=10 USDT (之前硬编码 12 过于保守)
	// - Hyperliquid: 支持 >=5 USDT (10 USDT 完全可用)
	return 10, nil
}

func (p *ConfigProvider) fetchMinPositionSizeBTCETH(ctx context.Context) (interface{}, error) {
	return 60, nil // BTC/ETH最小60 USDT
}

func (p *ConfigProvider) fetchExampleBTCPosition(ctx context.Context) (interface{}, error) {
	equity, err := p.getAccountEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.0f", equity*5), nil
}

// 内部数据获取方法

func (p *ConfigProvider) getAccountEquity(ctx context.Context) (float64, error) {
	// TODO: 实现从交易所API或数据库获取账户净值
	// 示例返回
	return 1000.0, nil
}
