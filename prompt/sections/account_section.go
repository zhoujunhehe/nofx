package sections

import "nofx/prompt"

// AccountSection 账户摘要部分
type AccountSection struct {
	content string
}

// SectionName常量
const AccountSectionName = "account_summary"

// NewAccountSection 创建账户摘要Section（使用默认模板）
func NewAccountSection() *AccountSection {
	return &AccountSection{
		content: `账户: 净值{account.total_equity} | 余额{account.available_balance} ({account.available_balance_pct}%) | 盈亏{account.total_pnl_pct}% | 保证金{account.margin_used_pct}% | 持仓{account.position_count}个

`,
	}
}

func (s *AccountSection) GetName() string {
	return AccountSectionName
}

func (s *AccountSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *AccountSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *AccountSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
