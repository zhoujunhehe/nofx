package core

import (
	"context"
	"strings"
	"testing"

	orchestrator "nofx/orchestrator"
)

// TestValidation_InvalidPlaceholderFormat 测试无效占位符格式
// 注意：使用严格正则后，{account}（无field部分）不会被识别为占位符
// 这是预期行为，可以避免与JSON内容冲突
func TestValidation_InvalidPlaceholderFormat(t *testing.T) {
	orch := New()

	// 注册provider
	accountProvider := NewMockProvider("account")
	orch.RegisterProvider(accountProvider)

	// {account} 不会被识别为占位符（缺少.field部分）
	builder := &orchestrator.PromptBuilder{
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "invalid",
				Template: "余额: {account} USDT", // 不会被解析为占位符
			},
		},
	}

	// 构建prompt应该成功（因为没有检测到占位符）
	ctx := context.Background()
	result, err := orch.BuildPrompt(ctx, builder)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证{account}保持原样（没有被替换）
	if !strings.Contains(result.AnalysisDataPrompt, "{account}") {
		t.Error("Expected {account} to remain unchanged (not recognized as placeholder)")
	}

	t.Logf("Correctly ignored invalid format: %s", result.AnalysisDataPrompt)
}

// TestValidation_UnregisteredProvider 测试未注册的Provider
func TestValidation_UnregisteredProvider(t *testing.T) {
	orch := New()

	// 不注册任何provider

	// 使用未注册的provider
	builder := &orchestrator.PromptBuilder{
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "market",
				Template: "价格: {market.BTCUSDT.price}",
			},
		},
	}

	// 构建prompt应该失败
	ctx := context.Background()
	_, err := orch.BuildPrompt(ctx, builder)

	if err == nil {
		t.Fatal("Expected validation error for unregistered provider")
	}

	if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("Expected 'not registered' error, got: %v", err)
	}

	t.Logf("Got expected error: %v", err)
}

// TestValidation_EmptyPlaceholder 测试空占位符
// 注意：使用严格正则后，{}不会被识别为占位符
func TestValidation_EmptyPlaceholder(t *testing.T) {
	orch := New()

	builder := &orchestrator.PromptBuilder{
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "empty",
				Template: "余额: {} USDT", // 不会被识别为占位符
			},
		},
	}

	// 构建prompt应该成功（因为没有检测到占位符）
	ctx := context.Background()
	result, err := orch.BuildPrompt(ctx, builder)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// 验证{}保持原样
	if !strings.Contains(result.AnalysisDataPrompt, "{}") {
		t.Error("Expected {} to remain unchanged")
	}

	t.Logf("Correctly ignored empty braces: %s", result.AnalysisDataPrompt)
}

// TestValidation_MultipleErrors 测试多个校验错误
func TestValidation_MultipleErrors(t *testing.T) {
	orch := New()

	// 只注册account provider
	accountProvider := NewMockProvider("account")
	orch.RegisterProvider(accountProvider)

	builder := &orchestrator.PromptBuilder{
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name: "section1",
				Template: `
				余额: {account.balance}
				价格: {market.price}
			`, // market未注册
			},
			{
				Name: "section2",
				Template: `
				未注册: {unknown.field}
			`, // unknown未注册
			},
		},
	}

	// 构建prompt应该失败
	ctx := context.Background()
	_, err := orch.BuildPrompt(ctx, builder)

	if err == nil {
		t.Fatal("Expected validation errors")
	}

	// 应该包含多个错误
	errMsg := err.Error()
	if !strings.Contains(errMsg, "market") {
		t.Errorf("Expected error about 'market' provider, got: %v", err)
	}
	if !strings.Contains(errMsg, "unknown") {
		t.Errorf("Expected error about 'unknown' provider, got: %v", err)
	}

	t.Logf("Got expected errors: %v", err)
}

// TestValidation_ValidPrompt 测试有效的Prompt通过校验
func TestValidation_ValidPrompt(t *testing.T) {
	orch := New()

	// 注册providers
	accountProvider := NewMockProvider("account")
	accountProvider.SetValue("balance", 10000.50)
	orch.RegisterProvider(accountProvider)

	marketProvider := NewMockProvider("market")
	marketProvider.SetValue("BTCUSDT.price", 51234.50)
	orch.RegisterProvider(marketProvider)

	// 有效的sections
	builder := &orchestrator.PromptBuilder{
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "account",
				Template: "余额: {account.balance} USDT",
			},
			{
				Name:     "market",
				Template: "价格: {market.BTCUSDT.price}",
			},
		},
	}

	// 构建prompt应该成功
	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)

	if err != nil {
		t.Fatalf("Expected validation to pass, got error: %v", err)
	}

	if prompt == nil {
		t.Fatal("Expected prompt to be built")
	}

	t.Logf("Validation passed, prompt built successfully")
}
