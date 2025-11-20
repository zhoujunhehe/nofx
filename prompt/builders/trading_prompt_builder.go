package builders

import (
	"fmt"

	"nofx/orchestrator"
	"nofx/prompt"
	"nofx/prompt/sections"
)

// TradingPromptBuilder 交易决策Prompt构建器
type TradingPromptBuilder struct {
	// SystemPrompt的各部分
	templateSection     prompt.Section
	riskControlSection  prompt.Section
	outputFormatSection prompt.Section

	// UserPrompt的各部分
	accountSection    prompt.Section
	positionsSection  prompt.Section
	marketDataSection prompt.Section

	// 配置
	config *TradingPromptConfig
}

// TradingPromptConfig 交易Prompt配置
type TradingPromptConfig struct {
	TemplateType    string // "conservative", "aggressive", "balanced"
	ShowPerformance bool   // 是否显示历史绩效
	ShowDetailedTA  bool   // 是否显示详细技术分析
}

// NewTradingPromptBuilder 创建交易Prompt构建器
func NewTradingPromptBuilder(config *TradingPromptConfig) orchestrator.PromptBuilder {
	if config == nil {
		config = &TradingPromptConfig{
			TemplateType:    "balanced",
			ShowPerformance: false,
			ShowDetailedTA:  false,
		}
	}

	return &TradingPromptBuilder{
		// 初始化SystemPrompt各部分
		templateSection:     sections.NewTemplateSection(config.TemplateType),
		riskControlSection:  sections.NewRiskControlSection(),
		outputFormatSection: sections.NewOutputFormatSection(),

		// 初始化UserPrompt各部分
		accountSection:    sections.NewAccountSection(),
		positionsSection:  sections.NewPositionsSection(),
		marketDataSection: sections.NewMarketDataSection(config.ShowDetailedTA),

		config: config,
	}
}

// GetDataRequirements 返回需要的数据类型
func (b *TradingPromptBuilder) GetDataRequirements() []string {
	requirements := make([]string, 0)

	// 收集所有Section的数据需求
	requirements = append(requirements, b.templateSection.GetDataRequirements()...)
	requirements = append(requirements, b.riskControlSection.GetDataRequirements()...)
	requirements = append(requirements, b.outputFormatSection.GetDataRequirements()...)
	requirements = append(requirements, b.accountSection.GetDataRequirements()...)
	requirements = append(requirements, b.positionsSection.GetDataRequirements()...)
	requirements = append(requirements, b.marketDataSection.GetDataRequirements()...)

	// TODO: 如果ShowPerformance=true，添加performance数据需求

	// 去重
	return prompt.UniqueStrings(requirements)
}

// FillSections 填充各个部分
func (b *TradingPromptBuilder) FillSections(data map[string]interface{}) (orchestrator.FilledSections, error) {
	sections := make(orchestrator.FilledSections)
	var err error

	// 填充SystemPrompt各部分（使用预定义常量）
	sections[orchestrator.SectionTemplate], err = b.templateSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill template section: %w", err)
	}

	sections[orchestrator.SectionRiskControl], err = b.riskControlSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill risk control section: %w", err)
	}

	sections[orchestrator.SectionOutputFormat], err = b.outputFormatSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill output format section: %w", err)
	}

	// 填充UserPrompt各部分（使用预定义常量）
	sections[orchestrator.SectionAccount], err = b.accountSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill account section: %w", err)
	}

	sections[orchestrator.SectionPositions], err = b.positionsSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill positions section: %w", err)
	}

	sections[orchestrator.SectionMarketData], err = b.marketDataSection.Fill(data)
	if err != nil {
		return nil, fmt.Errorf("failed to fill market data section: %w", err)
	}

	// TODO: 如果ShowPerformance=true，填充performance section

	return sections, nil
}

// GetAssemblyRules 获取拼接规则
func (b *TradingPromptBuilder) GetAssemblyRules() *orchestrator.AssemblyRules {
	// 使用预定义常量，清晰且避免拼写错误
	userPromptSections := []string{
		orchestrator.SectionAccount,
		orchestrator.SectionPositions,
		orchestrator.SectionMarketData,
	}

	if b.config.ShowPerformance {
		userPromptSections = append(userPromptSections, orchestrator.SectionPerformance)
	}

	return &orchestrator.AssemblyRules{
		SystemPromptSections: []string{
			orchestrator.SectionTemplate,
			orchestrator.SectionRiskControl,
			orchestrator.SectionOutputFormat,
		},
		UserPromptSections: userPromptSections,
		SectionSeparator:   "\n\n",
	}
}
