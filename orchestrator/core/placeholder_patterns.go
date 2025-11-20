package core

import "strings"

// ============================================================
// 占位符模式定义
// ============================================================

// TimeInterval 时间间隔常量
type TimeInterval string

const (
	// 分钟级别
	Interval1m  TimeInterval = "1m"
	Interval5m  TimeInterval = "5m"
	Interval15m TimeInterval = "15m"
	Interval30m TimeInterval = "30m"

	// 小时级别
	Interval1h  TimeInterval = "1h"
	Interval2h  TimeInterval = "2h"
	Interval4h  TimeInterval = "4h"
	Interval6h  TimeInterval = "6h"
	Interval12h TimeInterval = "12h"

	// 天级别
	Interval1d TimeInterval = "1d"
	Interval3d TimeInterval = "3d"

	// 周/月级别
	Interval1w TimeInterval = "1w"
	Interval1M TimeInterval = "1M"
)

// SupportedTimeIntervals 支持的所有时间间隔
var SupportedTimeIntervals = []TimeInterval{
	Interval1m, Interval5m, Interval15m, Interval30m,
	Interval1h, Interval2h, Interval4h, Interval6h, Interval12h,
	Interval1d, Interval3d,
	Interval1w, Interval1M,
}

// QuoteCurrency 计价货币常量
type QuoteCurrency string

const (
	QuoteUSDT QuoteCurrency = "USDT"
	QuoteUSDC QuoteCurrency = "USDC"
	QuoteBUSD QuoteCurrency = "BUSD"
	QuoteBTC  QuoteCurrency = "BTC"
	QuoteETH  QuoteCurrency = "ETH"
	QuoteBNB  QuoteCurrency = "BNB"
)

// SupportedQuoteCurrencies 支持的所有计价货币
var SupportedQuoteCurrencies = []QuoteCurrency{
	QuoteUSDT,
	QuoteUSDC,
	QuoteBUSD,
	QuoteBTC,
	QuoteETH,
	QuoteBNB,
}

// ParamName 参数名称常量
const (
	ParamInterval = "interval" // 时间间隔参数
	ParamSymbol   = "symbol"   // 交易对参数
)

// SymbolConfig 交易对识别配置
type SymbolConfig struct {
	MinLength       int             // 最小长度
	MaxLength       int             // 最大长度
	QuoteCurrencies []QuoteCurrency // 计价货币列表
}

// DefaultSymbolConfig 默认交易对配置
var DefaultSymbolConfig = &SymbolConfig{
	MinLength:       5,
	MaxLength:       12,
	QuoteCurrencies: SupportedQuoteCurrencies,
}

// ============================================================
// 模式匹配器
// ============================================================

// PatternMatcher 模式匹配器
type PatternMatcher struct {
	timeIntervals   map[string]bool
	symbolConfig    *SymbolConfig
}

// NewPatternMatcher 创建模式匹配器
func NewPatternMatcher() *PatternMatcher {
	// 构建时间间隔查找表
	timeIntervalMap := make(map[string]bool)
	for _, interval := range SupportedTimeIntervals {
		timeIntervalMap[string(interval)] = true
	}

	return &PatternMatcher{
		timeIntervals: timeIntervalMap,
		symbolConfig:  DefaultSymbolConfig,
	}
}

// IsTimeInterval 判断是否是时间间隔参数
func (m *PatternMatcher) IsTimeInterval(s string) bool {
	return m.timeIntervals[s]
}

// IsSymbol 判断是否是交易对参数
func (m *PatternMatcher) IsSymbol(s string) bool {
	// 检查长度
	if len(s) < m.symbolConfig.MinLength || len(s) > m.symbolConfig.MaxLength {
		return false
	}

	// 检查是否全大写
	if s != strings.ToUpper(s) {
		return false
	}

	// 检查是否以支持的计价货币结尾
	for _, quote := range m.symbolConfig.QuoteCurrencies {
		if strings.HasSuffix(s, string(quote)) {
			return true
		}
	}

	return false
}

// AddTimeInterval 添加自定义时间间隔
func (m *PatternMatcher) AddTimeInterval(interval string) {
	m.timeIntervals[interval] = true
}

// AddQuoteCurrency 添加自定义计价货币
func (m *PatternMatcher) AddQuoteCurrency(currency QuoteCurrency) {
	for _, existing := range m.symbolConfig.QuoteCurrencies {
		if existing == currency {
			return
		}
	}
	m.symbolConfig.QuoteCurrencies = append(m.symbolConfig.QuoteCurrencies, currency)
}
