package data_providers

import (
	"context"
	"fmt"
	"time"

	"nofx/orchestrator"
)

// SystemProvider 系统状态Provider
// 提供系统运行状态（时间、周期数、运行时长）
type SystemProvider struct {
	userID    string
	startTime time.Time
	// TODO: 添加其他依赖
}

// 字段处理器类型
type systemFieldHandler func(*SystemProvider, context.Context) (interface{}, error)

// 字段到方法的映射
var systemFieldHandlers = map[string]systemFieldHandler{
	"current_time":    (*SystemProvider).fetchCurrentTime,
	"call_count":      (*SystemProvider).fetchCallCount,
	"runtime_minutes": (*SystemProvider).fetchRuntimeMinutes,
}

// NewSystemProvider 创建SystemProvider
func NewSystemProvider(userID string) *SystemProvider {
	return &SystemProvider{
		userID:    userID,
		startTime: time.Now(),
	}
}

// Name 返回Provider名称
func (p *SystemProvider) Name() string {
	return "system"
}

// Fetch 获取字段值
func (p *SystemProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	handler, ok := systemFieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("system: unsupported field: %s", request.Field)
	}
	return handler(p, ctx)
}

// FetchBatch 批量获取字段值（并发）
func (p *SystemProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现

func (p *SystemProvider) fetchCurrentTime(ctx context.Context) (interface{}, error) {
	return time.Now().Format("2006-01-02 15:04:05"), nil
}

func (p *SystemProvider) fetchCallCount(ctx context.Context) (interface{}, error) {
	// TODO: 从数据库或计数器获取调用次数
	return p.getCallCount(ctx)
}

func (p *SystemProvider) fetchRuntimeMinutes(ctx context.Context) (interface{}, error) {
	minutes := int(time.Since(p.startTime).Minutes())
	return minutes, nil
}

// 内部数据获取方法

func (p *SystemProvider) getCallCount(ctx context.Context) (int, error) {
	// TODO: 实现从数据库获取调用次数
	return 0, nil
}
