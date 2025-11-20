package sections

import "nofx/prompt"

// OutputFormatSection 输出格式要求部分
type OutputFormatSection struct {
	content string
}

// SectionName常量
const OutputFormatSectionName = "output_format"

// NewOutputFormatSection 创建输出格式Section（使用默认内容）
func NewOutputFormatSection() *OutputFormatSection {
	return &OutputFormatSection{
		content: `# 输出格式 (严格遵守)

**必须使用XML标签 <reasoning> 和 <decision> 标签分隔思维链和决策JSON，避免解析错误**

## 格式要求

<reasoning>
你的思维链分析...
- 简洁分析你的思考过程
</reasoning>

<decision>
` + "```json\n[\n" + `  {"symbol": "BTCUSDT", "action": "open_short", "leverage": {config.btc_eth_leverage}, "position_size_usd": {config.example_btc_position}, "stop_loss": 97000, "take_profit": 91000, "confidence": 85, "risk_usd": 300, "reasoning": "下跌趋势+MACD死叉"},
  {"symbol": "SOLUSDT", "action": "update_stop_loss", "new_stop_loss": 155, "reasoning": "移动止损至保本位"},
  {"symbol": "ETHUSDT", "action": "close_long", "reasoning": "止盈离场"}
]
` + "```\n" + `</decision>

## 字段说明

- ` + "`action`" + `: open_long | open_short | close_long | close_short | update_stop_loss | update_take_profit | partial_close | hold | wait
- ` + "`confidence`" + `: 0-100（开仓建议≥75）
- 开仓时必填: leverage, position_size_usd, stop_loss, take_profit, confidence, risk_usd, reasoning
- update_stop_loss 时必填: new_stop_loss (注意是 new_stop_loss，不是 stop_loss)
- update_take_profit 时必填: new_take_profit (注意是 new_take_profit，不是 take_profit)
- partial_close 时必填: close_percentage (0-100)

`,
	}
}

func (s *OutputFormatSection) GetName() string {
	return OutputFormatSectionName
}

func (s *OutputFormatSection) GetContent() string {
	return s.content
}

// SetContent 设置自定义内容
func (s *OutputFormatSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从数据库/文件加载模板（未来扩展）
func (s *OutputFormatSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
