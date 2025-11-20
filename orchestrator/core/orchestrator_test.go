package core

import (
	"context"
	"strings"
	"testing"

	orchestrator "nofx/orchestrator"
)

// MockProvider 模拟Provider
type MockProvider struct {
	name   string
	values map[string]interface{}
}

func NewMockProvider(name string) *MockProvider {
	return &MockProvider{
		name:   name,
		values: make(map[string]interface{}),
	}
}

func (p *MockProvider) Name() string {
	return p.name
}

func (p *MockProvider) Fetch(ctx context.Context, request *orchestrator.FieldRequest) (interface{}, error) {
	value, ok := p.values[request.Field]
	if !ok {
		return nil, nil
	}
	return value, nil
}

// FetchBatch 实现BatchProvider接口
func (p *MockProvider) FetchBatch(ctx context.Context, requests []*orchestrator.BatchFieldRequest) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, req := range requests {
		value, ok := p.values[req.Field]
		if ok {
			result[req.Raw] = value
		}
	}

	return result, nil
}

func (p *MockProvider) SetValue(field string, value interface{}) {
	p.values[field] = value
}

// TestBuildPrompt_Basic 基础测试
func TestBuildPrompt_Basic(t *testing.T) {
	// 创建orchestrator
	orch := New()

	// 创建并注册mock provider
	accountProvider := NewMockProvider("account")
	accountProvider.SetValue("available_balance", 10000.50)
	accountProvider.SetValue("equity", 9500.00)

	orch.RegisterProvider(accountProvider)

	// 使用PromptBuilder构建
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     "template",
				Template: "你是交易助手",
			},
		},
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "account",
				Template: "余额: {account.available_balance} USDT\n净值: {account.equity} USDT",
			},
		},
	}

	// 构建prompt
	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证结果
	if !strings.Contains(prompt.AnalysisDataPrompt, "10000.50") || !strings.Contains(prompt.AnalysisDataPrompt, "9500.00") {
		t.Errorf("Expected AnalysisDataPrompt to contain balance and equity\nGot:\n%s", prompt.AnalysisDataPrompt)
	}

	t.Logf("SystemConstraintPrompt: %s", prompt.SystemConstraintPrompt)
	t.Logf("AnalysisDataPrompt: %s", prompt.AnalysisDataPrompt)
}

// TestPlaceholderParser_Extract 测试占位符提取
func TestPlaceholderParser_Extract(t *testing.T) {
	parser := NewPlaceholderParser()

	templates := []string{
		"余额: {account.available_balance} USDT",
		"价格: {market.BTCUSDT.price}",
		"K线: {market.1h.BTCUSDT.kline.close}",
	}

	placeholders, err := parser.ExtractPlaceholders(templates)
	if err != nil {
		t.Fatalf("ExtractPlaceholders failed: %v", err)
	}

	if len(placeholders) != 3 {
		t.Fatalf("Expected 3 placeholders, got %d", len(placeholders))
	}

	// 验证第一个占位符
	p1 := placeholders[0]
	if p1.Provider != "account" {
		t.Errorf("Expected provider 'account', got '%s'", p1.Provider)
	}
	if p1.Field != "available_balance" {
		t.Errorf("Expected field 'available_balance', got '%s'", p1.Field)
	}

	// 验证第二个占位符
	p2 := placeholders[1]
	if p2.Provider != "market" {
		t.Errorf("Expected provider 'market', got '%s'", p2.Provider)
	}
	if p2.Field != "BTCUSDT.price" {
		t.Errorf("Expected field 'BTCUSDT.price', got '%s'", p2.Field)
	}
	if p2.Params["symbol"] != "BTCUSDT" {
		t.Errorf("Expected symbol 'BTCUSDT', got '%s'", p2.Params["symbol"])
	}

	// 验证第三个占位符（带时间参数）
	p3 := placeholders[2]
	if p3.Provider != "market" {
		t.Errorf("Expected provider 'market', got '%s'", p3.Provider)
	}
	if p3.Params["interval"] != "1h" {
		t.Errorf("Expected interval '1h', got '%s'", p3.Params["interval"])
	}
	if p3.Params["symbol"] != "BTCUSDT" {
		t.Errorf("Expected symbol 'BTCUSDT', got '%s'", p3.Params["symbol"])
	}

	t.Logf("Parsed placeholders:")
	for i, p := range placeholders {
		t.Logf("  %d: Provider=%s, Field=%s, Params=%v", i+1, p.Provider, p.Field, p.Params)
	}
}

// TestPlaceholderParser_GroupByProvider 测试按Provider分组
func TestPlaceholderParser_GroupByProvider(t *testing.T) {
	parser := NewPlaceholderParser()

	templates := []string{
		"{account.available_balance}",
		"{account.equity}",
		"{market.BTCUSDT.price}",
		"{market.ETHUSDT.price}",
	}

	placeholders, err := parser.ExtractPlaceholders(templates)
	if err != nil {
		t.Fatalf("ExtractPlaceholders failed: %v", err)
	}

	grouped := parser.GroupByProvider(placeholders)

	if len(grouped) != 2 {
		t.Fatalf("Expected 2 providers, got %d", len(grouped))
	}

	accountPlaceholders := grouped["account"]
	if len(accountPlaceholders) != 2 {
		t.Errorf("Expected 2 account placeholders, got %d", len(accountPlaceholders))
	}

	marketPlaceholders := grouped["market"]
	if len(marketPlaceholders) != 2 {
		t.Errorf("Expected 2 market placeholders, got %d", len(marketPlaceholders))
	}

	t.Logf("Grouped by provider:")
	for provider, phs := range grouped {
		t.Logf("  %s: %d placeholders", provider, len(phs))
	}
}

// TestBuildPrompt_WithMultipleProviders 测试多Provider
func TestBuildPrompt_WithMultipleProviders(t *testing.T) {
	orch := New()

	// 注册account provider
	accountProvider := NewMockProvider("account")
	accountProvider.SetValue("available_balance", 10000.50)
	orch.RegisterProvider(accountProvider)

	// 注册market provider
	marketProvider := NewMockProvider("market")
	marketProvider.SetValue("BTCUSDT.price", 51234.50)
	orch.RegisterProvider(marketProvider)

	// 使用PromptBuilder构建
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     "template",
				Template: "你是交易助手",
			},
		},
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "data",
				Template: "余额: {account.available_balance}, 价格: {market.BTCUSDT.price}",
			},
		},
	}

	// 构建prompt
	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证AnalysisDataPrompt包含期望的数据
	if !strings.Contains(prompt.AnalysisDataPrompt, "10000.50") || !strings.Contains(prompt.AnalysisDataPrompt, "51234.50") {
		t.Errorf("Expected AnalysisDataPrompt to contain balance and price\nGot:\n%s", prompt.AnalysisDataPrompt)
	}

	t.Logf("SystemConstraintPrompt: %s", prompt.SystemConstraintPrompt)
	t.Logf("AnalysisDataPrompt: %s", prompt.AnalysisDataPrompt)
}
