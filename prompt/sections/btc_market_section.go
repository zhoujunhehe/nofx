package sections

import "nofx/prompt"

// BTCMarketSection BTC市场概况部分
type BTCMarketSection struct {
	content string
}

// SectionName常量
const BTCMarketSectionName = "btc_market"

// NewBTCMarketSection 创建BTC市场Section（使用默认模板）
func NewBTCMarketSection() *BTCMarketSection {
	return &BTCMarketSection{
		content: `BTC: {market.BTCUSDT.price} (1h: {market.BTCUSDT.change_1h}%, 4h: {market.BTCUSDT.change_4h}%) | MACD: {market.BTCUSDT.macd} | RSI: {market.BTCUSDT.rsi}

`,
	}
}

func (s *BTCMarketSection) GetName() string {
	return BTCMarketSectionName
}

func (s *BTCMarketSection) GetContent() string {
	return s.content
}

// SetTemplate 设置自定义模板（未来扩展）
func (s *BTCMarketSection) SetContent(content string) prompt.Section {
	s.content = content
	return s
}

// LoadTemplate 从文件加载模板（未来扩展）
func (s *BTCMarketSection) LoadTemplate(id int64) error {
	// TODO: 未来实现从数据库/文件加载
	return nil
}
