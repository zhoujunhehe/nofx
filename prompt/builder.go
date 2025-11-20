package prompt

// Note: PromptBuilder, FilledSections, AssemblyRules 类型定义在 orchestrator包中
// 这里不重复定义，直接使用 orchestrator包的类型

// Section 单个部分的定义
type Section interface {
	// GetName 获取Section名称
	GetName() string

	// GetContent 获取内容（包含占位符）
	GetContent() string

	// SetContent 设置内容
	SetContent(content string) Section

	// LoadTemplate 从数据库/文件加载模板
	LoadTemplate(id int64) error
}
