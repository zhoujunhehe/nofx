package examples

import (
	"context"
	"fmt"

	orchestrator "nofx/orchestrator"
)

// ============================================================
// AccountProvider 示例实现
// ============================================================

// AccountProvider 账户数据Provider
type AccountProvider struct {
	// 字段处理器映射：field → handler
	fieldHandlers map[string]FieldHandler

	// 可以注入实际的账户模块
	// accountModule *account.Module
}

// FieldHandler 字段处理器函数
type FieldHandler func(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error)

// NewAccountProvider 创建AccountProvider
func NewAccountProvider() *AccountProvider {
	p := &AccountProvider{
		fieldHandlers: make(map[string]FieldHandler),
	}

	// 注册所有字段处理器
	p.registerHandlers()

	return p
}

// Name 返回Provider名称
func (p *AccountProvider) Name() string {
	return "account"
}

// Fetch 获取字段值
func (p *AccountProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	// 查找字段处理器
	handler, ok := p.fieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("unsupported field: %s", request.Field)
	}

	// 调用处理器
	return handler(ctx, request)
}

// FetchBatch 批量获取字段值
// 实现BatchProvider接口，Provider可以在这里优化批量查询（比如一次性从数据库获取多个字段）
func (p *AccountProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 这里可以优化：
	// 1. 先分析所有请求需要的数据
	// 2. 一次性从数据库/API获取所有数据
	// 3. 然后填充到结果中
	//
	// 目前的简单实现：逐个调用处理器（但在实际应用中应该优化）
	for _, req := range requests {
		handler, ok := p.fieldHandlers[req.Field]
		if !ok {
			return nil, fmt.Errorf("unsupported field: %s", req.Field)
		}

		value, err := handler(ctx, req.FieldRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch %s: %w", req.Raw, err)
		}

		result[req.Raw] = value
	}

	return result, nil
}

// registerHandlers 注册所有字段处理器
func (p *AccountProvider) registerHandlers() {
	// 简单字段
	p.fieldHandlers["available_balance"] = p.getAvailableBalance
	p.fieldHandlers["equity"] = p.getEquity
	p.fieldHandlers["margin_used"] = p.getMarginUsed
	p.fieldHandlers["available_margin"] = p.getAvailableMargin
	p.fieldHandlers["unrealized_pnl"] = p.getUnrealizedPnL
	p.fieldHandlers["total_position_value"] = p.getTotalPositionValue
}

// ============================================================
// 字段处理器实现
// ============================================================

// getAvailableBalance 获取可用余额
func (p *AccountProvider) getAvailableBalance(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	// 从ctx中获取用户信息
	traderID := ctx.Value("trader_id").(string)

	// 调用实际的账户模块
	// account := p.accountModule.GetAccount(traderID)
	// return account.AvailableBalance, nil

	// 示例数据
	_ = traderID
	return 10000.50, nil
}

// getEquity 获取净值
func (p *AccountProvider) getEquity(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	traderID := ctx.Value("trader_id").(string)
	_ = traderID
	return 9500.00, nil
}

// getMarginUsed 获取已用保证金
func (p *AccountProvider) getMarginUsed(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	traderID := ctx.Value("trader_id").(string)
	_ = traderID
	return 500.00, nil
}

// getAvailableMargin 获取可用保证金
func (p *AccountProvider) getAvailableMargin(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	traderID := ctx.Value("trader_id").(string)
	_ = traderID
	return 9000.00, nil
}

// getUnrealizedPnL 获取未实现盈亏
func (p *AccountProvider) getUnrealizedPnL(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	traderID := ctx.Value("trader_id").(string)
	_ = traderID
	return 150.25, nil
}

// getTotalPositionValue 获取持仓总价值
func (p *AccountProvider) getTotalPositionValue(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	traderID := ctx.Value("trader_id").(string)
	_ = traderID
	return 5000.00, nil
}
