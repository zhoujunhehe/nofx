package data_providers

import (
	"context"
	"fmt"

	"nofx/orchestrator"
)

// AccountProvider 账户数据Provider
// 提供账户状态（净值、余额、盈亏、保证金使用率等）
type AccountProvider struct {
	userID string
	// TODO: 添加交易所API客户端、数据库连接等依赖
}

// 字段处理器类型
type accountFieldHandler func(*AccountProvider, context.Context) (interface{}, error)

// 字段到方法的映射
var accountFieldHandlers = map[string]accountFieldHandler{
	"total_equity":        (*AccountProvider).fetchTotalEquity,
	"available_balance":   (*AccountProvider).fetchAvailableBalance,
	"available_balance_pct": (*AccountProvider).fetchAvailableBalancePct,
	"total_pnl_pct":       (*AccountProvider).fetchTotalPnLPct,
	"margin_used_pct":     (*AccountProvider).fetchMarginUsedPct,
	"position_count":      (*AccountProvider).fetchPositionCount,
}

// NewAccountProvider 创建AccountProvider
func NewAccountProvider(userID string) *AccountProvider {
	return &AccountProvider{
		userID: userID,
	}
}

// Name 返回Provider名称
func (p *AccountProvider) Name() string {
	return "account"
}

// Fetch 获取字段值
func (p *AccountProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	handler, ok := accountFieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("account: unsupported field: %s", request.Field)
	}
	return handler(p, ctx)
}

// FetchBatch 批量获取字段值（并发）
func (p *AccountProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现

func (p *AccountProvider) fetchTotalEquity(ctx context.Context) (interface{}, error) {
	equity, err := p.getTotalEquity(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", equity), nil
}

func (p *AccountProvider) fetchAvailableBalance(ctx context.Context) (interface{}, error) {
	balance, err := p.getAvailableBalance(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", balance), nil
}

func (p *AccountProvider) fetchAvailableBalancePct(ctx context.Context) (interface{}, error) {
	equity, err := p.getTotalEquity(ctx)
	if err != nil {
		return nil, err
	}
	balance, err := p.getAvailableBalance(ctx)
	if err != nil {
		return nil, err
	}
	if equity > 0 {
		pct := (balance / equity) * 100
		return fmt.Sprintf("%.1f", pct), nil
	}
	return "0.0", nil
}

func (p *AccountProvider) fetchTotalPnLPct(ctx context.Context) (interface{}, error) {
	pnl, err := p.getTotalPnLPct(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.2f", pnl), nil
}

func (p *AccountProvider) fetchMarginUsedPct(ctx context.Context) (interface{}, error) {
	margin, err := p.getMarginUsedPct(ctx)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%.1f", margin), nil
}

func (p *AccountProvider) fetchPositionCount(ctx context.Context) (interface{}, error) {
	return p.getPositionCount(ctx)
}

// 内部数据获取方法 - TODO: 实现真实数据获取逻辑

func (p *AccountProvider) getTotalEquity(ctx context.Context) (float64, error) {
	// TODO: 从交易所API获取账户净值
	return 1000.0, nil
}

func (p *AccountProvider) getAvailableBalance(ctx context.Context) (float64, error) {
	// TODO: 从交易所API获取可用余额
	return 800.0, nil
}

func (p *AccountProvider) getTotalPnLPct(ctx context.Context) (float64, error) {
	// TODO: 计算总盈亏百分比
	return 5.5, nil
}

func (p *AccountProvider) getMarginUsedPct(ctx context.Context) (float64, error) {
	// TODO: 计算保证金使用率
	return 20.0, nil
}

func (p *AccountProvider) getPositionCount(ctx context.Context) (int, error) {
	// TODO: 获取当前持仓数量
	return 3, nil
}
