package data_providers

import (
	"context"
	"fmt"

	"nofx/orchestrator"
)

// CustomProvider 用户自定义数据Provider
// 提供用户自定义的策略文本等内容
type CustomProvider struct {
	userID string
	// TODO: 添加数据库连接等依赖
}

// 字段处理器类型
type customFieldHandler func(*CustomProvider, context.Context) (interface{}, error)

// 字段到方法的映射
var customFieldHandlers = map[string]customFieldHandler{
	"user_prompt": (*CustomProvider).fetchUserPrompt,
}

// NewCustomProvider 创建CustomProvider
func NewCustomProvider(userID string) *CustomProvider {
	return &CustomProvider{
		userID: userID,
	}
}

// Name 返回Provider名称
func (p *CustomProvider) Name() string {
	return "custom"
}

// Fetch 获取字段值
func (p *CustomProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	handler, ok := customFieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("custom: unsupported field: %s", request.Field)
	}
	return handler(p, ctx)
}

// FetchBatch 批量获取字段值（并发）
func (p *CustomProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现

func (p *CustomProvider) fetchUserPrompt(ctx context.Context) (interface{}, error) {
	return p.getUserPrompt(ctx)
}

// 内部数据获取方法 - TODO: 实现真实数据获取逻辑

func (p *CustomProvider) getUserPrompt(ctx context.Context) (string, error) {
	// TODO: 从数据库获取用户自定义策略文本
	return "", nil
}
