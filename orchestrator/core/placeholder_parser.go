package core

import (
	"fmt"
	"regexp"
	"strings"

	orchestrator "nofx/orchestrator"
)

// ============================================================
// 占位符数据结构
// ============================================================

// Placeholder 占位符
type Placeholder struct {
	// Raw 原始占位符（包括大括号）
	// 示例: "{account.available_balance}"
	Raw string

	// Provider Provider名称
	// 示例: "account", "market", "config"
	Provider string

	// Field 字段路径（去除provider后的完整路径）
	// 示例:
	//   {account.available_balance} → "available_balance"
	//   {market.BTCUSDT.price} → "BTCUSDT.price"
	//   {market.1h.kline.close} → "1h.kline.close"
	Field string

	// Params 解析出的参数
	// 示例:
	//   {market.BTCUSDT.price} → {"symbol": "BTCUSDT"}
	//   {market.1h.kline.close} → {"interval": "1h"}
	//   {market.1h.BTCUSDT.kline.close} → {"interval": "1h", "symbol": "BTCUSDT"}
	Params map[string]string
}

// ============================================================
// 占位符解析器
// ============================================================

// PlaceholderParser 占位符解析器
type PlaceholderParser struct {
	// 占位符正则表达式: {xxx}
	placeholderRegex *regexp.Regexp

	// 模式匹配器
	patternMatcher *PatternMatcher
}

// NewPlaceholderParser 创建占位符解析器
func NewPlaceholderParser() *PlaceholderParser {
	return &PlaceholderParser{
		// 更严格的占位符正则表达式：
		// - 必须以字母或下划线开头（provider名）
		// - 后跟 .字段名（至少一个）
		// - 字段名可以包含字母、数字、下划线
		// - 不匹配包含引号、冒号、空格等JSON特殊字符的内容
		// 示例匹配：{config.leverage}, {market.BTCUSDT.price}, {account.total_equity}
		// 示例不匹配：{"key": "value"}, {config
		placeholderRegex: regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z0-9_]+)+)\}`),
		patternMatcher:   NewPatternMatcher(),
	}
}

// ExtractPlaceholders 从模板中提取所有占位符
func (p *PlaceholderParser) ExtractPlaceholders(templates []string) ([]*Placeholder, error) {
	placeholders := make([]*Placeholder, 0)
	seen := make(map[string]bool)

	for _, template := range templates {
		matches := p.placeholderRegex.FindAllStringSubmatch(template, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}

			raw := match[0]     // {account.available_balance}
			content := match[1] // account.available_balance

			// 去重
			if seen[raw] {
				continue
			}
			seen[raw] = true

			// 解析占位符
			placeholder, err := p.parsePlaceholder(raw, content)
			if err != nil {
				return nil, fmt.Errorf("failed to parse placeholder %s: %w", raw, err)
			}

			placeholders = append(placeholders, placeholder)
		}
	}

	return placeholders, nil
}

// parsePlaceholder 解析单个占位符
func (p *PlaceholderParser) parsePlaceholder(raw, content string) (*Placeholder, error) {
	// content格式: provider.field 或 provider.param1.param2.field
	// 示例:
	//   "account.available_balance"
	//   "market.BTCUSDT.price"
	//   "market.1h.kline.close"
	//   "market.1h.BTCUSDT.kline.close"

	parts := strings.Split(content, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid placeholder format: %s (expected at least provider.field)", content)
	}

	placeholder := &Placeholder{
		Raw:      raw,
		Provider: parts[0],
		Field:    strings.Join(parts[1:], "."),
		Params:   make(map[string]string),
	}

	// 解析参数（如果有）
	p.parseParams(placeholder, parts[1:])

	return placeholder, nil
}

// parseParams 解析占位符中的参数
func (p *PlaceholderParser) parseParams(placeholder *Placeholder, fieldParts []string) {
	// fieldParts示例:
	//   ["available_balance"]                → 无参数
	//   ["BTCUSDT", "price"]                 → symbol参数
	//   ["1h", "kline", "close"]             → interval参数
	//   ["1h", "BTCUSDT", "kline", "close"]  → interval + symbol参数

	// 启发式解析参数
	for i, part := range fieldParts {
		// 时间间隔参数: 1h, 4h, 1d等
		if p.patternMatcher.IsTimeInterval(part) {
			placeholder.Params[ParamInterval] = part
			continue
		}

		// 交易对参数: BTCUSDT, ETHUSDT等
		if p.patternMatcher.IsSymbol(part) {
			placeholder.Params[ParamSymbol] = part
			continue
		}

		// 其他参数作为通用参数
		// 可以根据位置或其他规则扩展
		_ = i
	}
}

// GroupByProvider 按Provider分组占位符
func (p *PlaceholderParser) GroupByProvider(placeholders []*Placeholder) map[string][]*Placeholder {
	grouped := make(map[string][]*Placeholder)

	for _, placeholder := range placeholders {
		providerName := placeholder.Provider
		grouped[providerName] = append(grouped[providerName], placeholder)
	}

	return grouped
}

// ToFieldRequests 将占位符转换为FieldRequest列表
func (p *PlaceholderParser) ToFieldRequests(placeholders []*Placeholder) []*orchestrator.FieldRequest {
	requests := make([]*orchestrator.FieldRequest, len(placeholders))

	for i, placeholder := range placeholders {
		requests[i] = &orchestrator.FieldRequest{
			Field:  placeholder.Field,
			Params: placeholder.Params,
		}
	}

	return requests
}

// GetPatternMatcher 获取模式匹配器（用于自定义配置）
func (p *PlaceholderParser) GetPatternMatcher() *PatternMatcher {
	return p.patternMatcher
}
