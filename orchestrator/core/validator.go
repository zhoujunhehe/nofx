package core

import (
	"fmt"
	"regexp"
	"strings"

	orchestrator "nofx/orchestrator"
)

// ValidationError 校验错误
type ValidationError struct {
	SectionName string   // 出错的Section名称
	Errors      []string // 错误列表
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed in section '%s': %s", e.SectionName, strings.Join(e.Errors, "; "))
}

// PromptValidator Prompt校验器
type PromptValidator struct {
	placeholderPattern *regexp.Regexp
}

// NewPromptValidator 创建校验器
func NewPromptValidator() *PromptValidator {
	return &PromptValidator{
		// 更严格的占位符正则表达式（与PlaceholderParser保持一致）：
		// - 必须以字母或下划线开头（provider名）
		// - 后跟 .字段名（至少一个）
		// - 字段名可以包含字母、数字、下划线
		// - 不匹配包含引号、冒号、空格等JSON特殊字符的内容
		// 示例匹配：{config.leverage}, {market.BTCUSDT.price}
		// 示例不匹配：{"key": "value"}
		placeholderPattern: regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*(?:\.[a-zA-Z0-9_]+)+)\}`),
	}
}

// ValidateSections 校验所有Sections
// 返回所有校验错误
func (v *PromptValidator) ValidateSections(sections []*orchestrator.Section, registeredProviders map[string]bool) []*ValidationError {
	errors := make([]*ValidationError, 0)

	for _, section := range sections {
		sectionErrors := v.validateSection(section, registeredProviders)
		if len(sectionErrors) > 0 {
			errors = append(errors, &ValidationError{
				SectionName: section.Name,
				Errors:      sectionErrors,
			})
		}
	}

	return errors
}

// validateSection 校验单个Section
func (v *PromptValidator) validateSection(section *orchestrator.Section, registeredProviders map[string]bool) []string {
	errors := make([]string, 0)

	// 提取所有占位符
	matches := v.placeholderPattern.FindAllStringSubmatch(section.Template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[1] // {account.balance} -> account.balance
		placeholderErrors := v.validatePlaceholder(placeholder, registeredProviders)
		errors = append(errors, placeholderErrors...)
	}

	return errors
}

// validatePlaceholder 校验单个占位符
func (v *PromptValidator) validatePlaceholder(placeholder string, registeredProviders map[string]bool) []string {
	errors := make([]string, 0)

	// 1. 检查占位符是否为空
	if strings.TrimSpace(placeholder) == "" {
		errors = append(errors, fmt.Sprintf("empty placeholder '{%s}'", placeholder))
		return errors
	}

	// 2. 解析占位符：provider.field...
	parts := strings.Split(placeholder, ".")
	if len(parts) < 2 {
		errors = append(errors, fmt.Sprintf("invalid placeholder format '{%s}': must be in format '{provider.field}'", placeholder))
		return errors
	}

	// 3. 提取provider名称（第一部分）
	providerName := parts[0]
	if providerName == "" {
		errors = append(errors, fmt.Sprintf("empty provider name in placeholder '{%s}'", placeholder))
		return errors
	}

	// 4. 检查provider是否已注册
	if !registeredProviders[providerName] {
		errors = append(errors, fmt.Sprintf("provider '%s' not registered (placeholder: '{%s}')", providerName, placeholder))
	}

	// 5. 检查field部分是否为空
	fieldParts := parts[1:]
	if len(fieldParts) == 0 || (len(fieldParts) == 1 && fieldParts[0] == "") {
		errors = append(errors, fmt.Sprintf("empty field in placeholder '{%s}'", placeholder))
	}

	return errors
}

// FormatValidationErrors 格式化所有校验错误为可读字符串
func FormatValidationErrors(errors []*ValidationError) string {
	if len(errors) == 0 {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("prompt validation failed:\n")

	for i, err := range errors {
		builder.WriteString(fmt.Sprintf("  [%d] Section '%s':\n", i+1, err.SectionName))
		for _, msg := range err.Errors {
			builder.WriteString(fmt.Sprintf("      - %s\n", msg))
		}
	}

	return builder.String()
}
