package data_providers

import (
	"context"
	"fmt"
	"strings"

	"nofx/orchestrator"
)

// CandidateData 单个候选币种数据
type CandidateData struct {
	Symbol    string
	Price     float64
	Change1h  float64
	Change4h  float64
	RSI       float64
	MACD      float64
	Volume24h float64 // 24小时交易量
	Score     float64 // 筛选评分（可选）
}

// CandidateProvider 候选币种数据Provider
// 提供候选币种列表的格式化输出
type CandidateProvider struct {
	userID string
	// TODO: 添加筛选服务、市场数据服务等依赖
}

// 字段处理器类型
type candidateFieldHandler func(*CandidateProvider, context.Context) (interface{}, error)

// 字段到方法的映射
var candidateFieldHandlers = map[string]candidateFieldHandler{
	"count": (*CandidateProvider).fetchCount,
	"list":  (*CandidateProvider).fetchList,
}

// NewCandidateProvider 创建CandidateProvider
func NewCandidateProvider(userID string) *CandidateProvider {
	return &CandidateProvider{
		userID: userID,
	}
}

// Name 返回Provider名称
func (p *CandidateProvider) Name() string {
	return "candidate"
}

// Fetch 获取字段值
func (p *CandidateProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	handler, ok := candidateFieldHandlers[request.Field]
	if !ok {
		return nil, fmt.Errorf("candidate: unsupported field: %s", request.Field)
	}
	return handler(p, ctx)
}

// FetchBatch 批量获取字段值（并发）
func (p *CandidateProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	return FetchBatchConcurrent(p.Fetch, ctx, requests)
}

// 字段方法实现

func (p *CandidateProvider) fetchCount(ctx context.Context) (interface{}, error) {
	candidates, err := p.getCandidates(ctx)
	if err != nil {
		return nil, err
	}
	return len(candidates), nil
}

func (p *CandidateProvider) fetchList(ctx context.Context) (interface{}, error) {
	candidates, err := p.getCandidates(ctx)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return "当前无候选币种", nil
	}

	var sb strings.Builder
	for i, c := range candidates {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		// 基本信息
		sb.WriteString(fmt.Sprintf("### %s\n", c.Symbol))
		sb.WriteString(fmt.Sprintf("价格: $%.4f | 24h量: $%.0f\n", c.Price, c.Volume24h))
		sb.WriteString(fmt.Sprintf("1h: %.2f%% | 4h: %.2f%% | RSI: %.1f | MACD: %.4f", c.Change1h, c.Change4h, c.RSI, c.MACD))

		// 如果有评分，显示评分
		if c.Score > 0 {
			sb.WriteString(fmt.Sprintf(" | 评分: %.2f", c.Score))
		}
	}

	return sb.String(), nil
}

// 内部数据获取方法 - TODO: 实现真实数据获取逻辑

func (p *CandidateProvider) getCandidates(ctx context.Context) ([]*CandidateData, error) {
	// TODO: 从筛选服务获取候选币种
	return []*CandidateData{}, nil
}
