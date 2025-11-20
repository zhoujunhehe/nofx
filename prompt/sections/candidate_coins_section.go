package sections

import "nofx/prompt"

// CandidateCoinsSection 候选币种详情部分
type CandidateCoinsSection struct {
	content string
}

// SectionName常量
const CandidateCoinsSectionName = "candidate_coins"

// NewCandidateCoinsSection 创建候选币种Section（使用默认模板）
func NewCandidateCoinsSection() *CandidateCoinsSection {
	return &CandidateCoinsSection{
		content: `## 候选币种 ({candidate.count}个)

{candidate.list}
`,
	}
}

func (s *CandidateCoinsSection) GetName() string {
	return CandidateCoinsSectionName
}

func (s *CandidateCoinsSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *CandidateCoinsSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *CandidateCoinsSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
