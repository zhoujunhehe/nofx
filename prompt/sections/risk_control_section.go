package sections

import "nofx/prompt"

// RiskControlSection 风险控制硬约束部分
type RiskControlSection struct {
	content string
}

// SectionName常量
const RiskControlSectionName = "risk_control"

// NewRiskControlSection 创建风险控制Section（使用默认内容）
func NewRiskControlSection() *RiskControlSection {
	return &RiskControlSection{
		content: `# 硬约束（风险控制）

1. 风险回报比: 必须 ≥ 1:3（冒1%风险，赚3%+收益）
2. 最多持仓: 3个币种（质量>数量）
3. 单币仓位: 山寨{config.altcoin_min_position}-{config.altcoin_max_position} U | BTC/ETH {config.btc_eth_min_position}-{config.btc_eth_max_position} U
4. 杠杆限制: **山寨币最大{config.altcoin_leverage}x杠杆** | **BTC/ETH最大{config.btc_eth_leverage}x杠杆** (⚠️ 严格执行，不可超过)
5. 保证金: 总使用率 ≤ 90%
6. 开仓金额: 建议 **≥{config.min_position_size_general} USDT** (交易所最小名义价值 10 USDT + 安全边际)

`,
	}
}

func (s *RiskControlSection) GetName() string {
	return RiskControlSectionName
}

func (s *RiskControlSection) GetContent() string {
	return s.content
}

// SetContent 设置自定义内容
func (s *RiskControlSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从数据库/文件加载模板（未来扩展）
func (s *RiskControlSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
