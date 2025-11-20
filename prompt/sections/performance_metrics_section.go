package sections

import "nofx/prompt"

// PerformanceMetricsSection 历史表现指标部分（可选）
type PerformanceMetricsSection struct {
	content string
}

// SectionName常量
const PerformanceMetricsSectionName = "performance_metrics"

// NewPerformanceMetricsSection 创建历史表现指标Section（使用默认模板）
func NewPerformanceMetricsSection() *PerformanceMetricsSection {
	return &PerformanceMetricsSection{
		content: `## 📊 夏普比率: {performance.sharpe_ratio}

`,
	}
}

func (s *PerformanceMetricsSection) GetName() string {
	return PerformanceMetricsSectionName
}

func (s *PerformanceMetricsSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *PerformanceMetricsSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *PerformanceMetricsSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
