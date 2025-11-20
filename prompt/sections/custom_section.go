package sections

import "nofx/prompt"

// CustomStrategySection 个性化交易策略部分（可选）
type CustomStrategySection struct {
	content string
}

// SectionName常量
const CustomStrategySectionName = "custom_strategy"

// NewCustomStrategySection 创建自定义策略Section（使用默认模板）
func NewCustomStrategySection() *CustomStrategySection {
	return &CustomStrategySection{
		content: `# 📌 个性化交易策略

{custom.user_prompt}

注意: 以上个性化策略是对基础规则的补充，不能违背基础风险控制原则。
`,
	}
}

func (s *CustomStrategySection) GetName() string {
	return CustomStrategySectionName
}

func (s *CustomStrategySection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *CustomStrategySection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *CustomStrategySection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}