package core

import (
	"context"
	"strings"
	"testing"

	orchestrator "nofx/orchestrator"
)

// TestSectionOrdering 测试Section排序功能
func TestSectionOrdering(t *testing.T) {
	orch := New()

	// 注册provider
	accountProvider := NewMockProvider("account")
	accountProvider.SetValue("balance", 10000.0)
	orch.RegisterProvider(accountProvider)

	// 使用PromptBuilder明确指定顺序
	// SystemConstraintSections数组的顺序就是最终顺序
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     "base_template",
				Template: "## 基础模板",
			},
			{
				Name:     "risk_control",
				Template: "## 风控规则",
			},
			{
				Name:     "output_format",
				Template: "## 输出格式",
			},
		},
		AnalysisDataSections: []*orchestrator.Section{
			{
				Name:     "account_info",
				Template: "## 账户信息: {account.balance}",
			},
			{
				Name:     "market_data",
				Template: "## 市场数据",
			},
			{
				Name:     "position_info",
				Template: "## 持仓信息",
			},
		},
	}

	// 构建prompt
	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证SystemConstraintPrompt的顺序
	expectedSystemOrder := []string{
		"## 基础模板",
		"## 风控规则",
		"## 输出格式",
	}

	systemLines := strings.Split(strings.TrimSpace(prompt.SystemConstraintPrompt), "\n\n")
	if len(systemLines) != 3 {
		t.Errorf("Expected 3 system sections, got %d", len(systemLines))
	}

	for i, expected := range expectedSystemOrder {
		if i >= len(systemLines) {
			t.Errorf("Missing section at index %d", i)
			continue
		}
		if !strings.Contains(systemLines[i], expected) {
			t.Errorf("SystemConstraint[%d]: expected to contain '%s', got '%s'", i, expected, systemLines[i])
		}
	}

	// 验证AnalysisDataPrompt的顺序
	expectedDataOrder := []string{
		"## 账户信息",
		"## 市场数据",
		"## 持仓信息",
	}

	dataLines := strings.Split(strings.TrimSpace(prompt.AnalysisDataPrompt), "\n\n")
	if len(dataLines) != 3 {
		t.Errorf("Expected 3 data sections, got %d", len(dataLines))
	}

	for i, expected := range expectedDataOrder {
		if i >= len(dataLines) {
			t.Errorf("Missing section at index %d", i)
			continue
		}
		if !strings.Contains(dataLines[i], expected) {
			t.Errorf("AnalysisData[%d]: expected to contain '%s', got '%s'", i, expected, dataLines[i])
		}
	}

	t.Logf("SystemConstraintPrompt:\n%s\n", prompt.SystemConstraintPrompt)
	t.Logf("AnalysisDataPrompt:\n%s\n", prompt.AnalysisDataPrompt)
}

// TestSectionOrdering_ArrayOrder 测试数组顺序即为最终顺序
func TestSectionOrdering_ArrayOrder(t *testing.T) {
	orch := New()

	// 数组顺序就是最终Prompt的顺序
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     "section1",
				Template: "Section 1",
			},
			{
				Name:     "section2",
				Template: "Section 2",
			},
			{
				Name:     "section3",
				Template: "Section 3",
			},
		},
	}

	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证保持数组顺序
	expected := "Section 1\n\nSection 2\n\nSection 3"
	if prompt.SystemConstraintPrompt != expected {
		t.Errorf("Expected '%s', got '%s'", expected, prompt.SystemConstraintPrompt)
	}

	t.Logf("Result: %s", prompt.SystemConstraintPrompt)
}

// TestSectionOrdering_CustomOrder 测试自定义顺序
func TestSectionOrdering_CustomOrder(t *testing.T) {
	orch := New()

	// 故意将sections放置为非常规顺序：Last, First, Middle
	// 最终Prompt应该严格按照数组顺序输出
	builder := &orchestrator.PromptBuilder{
		SystemConstraintSections: []*orchestrator.Section{
			{
				Name:     "last",
				Template: "Last Section",
			},
			{
				Name:     "first",
				Template: "First Section",
			},
			{
				Name:     "middle",
				Template: "Middle Section",
			},
		},
	}

	ctx := context.Background()
	prompt, err := orch.BuildPrompt(ctx, builder)
	if err != nil {
		t.Fatalf("BuildPrompt failed: %v", err)
	}

	// 验证顺序严格按照数组顺序
	lines := strings.Split(prompt.SystemConstraintPrompt, "\n\n")
	if len(lines) != 3 {
		t.Fatalf("Expected 3 sections, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "Last") {
		t.Errorf("Expected first line to be 'Last Section', got '%s'", lines[0])
	}
	if !strings.Contains(lines[1], "First") {
		t.Errorf("Expected second line to be 'First Section', got '%s'", lines[1])
	}
	if !strings.Contains(lines[2], "Middle") {
		t.Errorf("Expected third line to be 'Middle Section', got '%s'", lines[2])
	}

	t.Logf("Result:\n%s", prompt.SystemConstraintPrompt)
}
