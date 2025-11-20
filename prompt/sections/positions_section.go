package sections

import "nofx/prompt"

// PositionsSection 当前持仓详情部分
type PositionsSection struct {
	content string
}

// SectionName常量
const PositionsSectionName = "current_positions"

// NewPositionsSection 创建持仓Section（使用默认模板）
func NewPositionsSection() *PositionsSection {
	return &PositionsSection{
		content: `## 当前持仓

{position.list}
`,
	}
}

func (s *PositionsSection) GetName() string {
	return PositionsSectionName
}

func (s *PositionsSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *PositionsSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *PositionsSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
