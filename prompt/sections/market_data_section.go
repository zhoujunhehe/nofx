package sections

import "nofx/prompt"

// MarketDataSection 市场数据部分
type MarketDataSection struct {
	showDetailedTA bool
	content        string
	itemTemplate   string
}

// SectionName常量
const MarketDataSectionName = "market_data"

// NewMarketDataSection 创建市场数据Section
func NewMarketDataSection(showDetailedTA bool) *MarketDataSection {
	itemTemplate := "### {symbol}\n价格: ${price} | 1h: {change_1h}% | 4h: {change_4h}%"
	if showDetailedTA {
		itemTemplate += " | RSI: {rsi} | MACD: {macd}"
	}

	return &MarketDataSection{
		showDetailedTA: showDetailedTA,
		content:        "## 市场数据\n{market_list}",
		itemTemplate:   itemTemplate,
	}
}

func (s *MarketDataSection) GetName() string {
	return MarketDataSectionName
}

func (s *MarketDataSection) GetContent() string {
	return s.content
}

// SetContent 设置自定义内容
func (s *MarketDataSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从数据库/文件加载模板
func (s *MarketDataSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}

