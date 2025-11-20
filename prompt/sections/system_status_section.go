package sections

import "nofx/prompt"

// SystemStatusSection 系统状态部分
type SystemStatusSection struct {
	content string
}

// SectionName常量
const SystemStatusSectionName = "system_status"

// NewSystemStatusSection 创建系统状态Section（使用默认模板）
func NewSystemStatusSection() *SystemStatusSection {
	return &SystemStatusSection{
		content: `时间: {system.current_time} | 周期: #{system.call_count} | 运行: {system.runtime_minutes}分钟

`,
	}
}

func (s *SystemStatusSection) GetName() string {
	return SystemStatusSectionName
}

func (s *SystemStatusSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *SystemStatusSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *SystemStatusSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
