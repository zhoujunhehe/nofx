package core

import (
	"fmt"
	"strings"
)

// ============================================================
// 占位符替换器
// ============================================================

// PlaceholderReplacer 占位符替换器
type PlaceholderReplacer struct {
	formatter *ValueFormatter
}

// NewPlaceholderReplacer 创建占位符替换器
func NewPlaceholderReplacer() *PlaceholderReplacer {
	return &PlaceholderReplacer{
		formatter: NewValueFormatter(),
	}
}

// Replace 替换模板中的所有占位符
func (r *PlaceholderReplacer) Replace(template string, placeholders []*Placeholder, values map[string]interface{}) string {
	result := template

	for _, placeholder := range placeholders {
		// 获取值
		value, ok := values[placeholder.Raw]
		if !ok {
			// 如果没有找到值，保持原样或替换为空
			continue
		}

		// 格式化值
		valueStr := r.formatter.Format(value)

		// 替换
		result = strings.ReplaceAll(result, placeholder.Raw, valueStr)
	}

	return result
}

// ============================================================
// 值格式化器
// ============================================================

// ValueFormatter 值格式化器
type ValueFormatter struct {
	floatPrecision int
}

// NewValueFormatter 创建值格式化器
func NewValueFormatter() *ValueFormatter {
	return &ValueFormatter{
		floatPrecision: 6, // 默认保留2位小数
	}
}

// SetFloatPrecision 设置浮点数精度
func (f *ValueFormatter) SetFloatPrecision(precision int) {
	f.floatPrecision = precision
}

// Format 格式化值为字符串
func (f *ValueFormatter) Format(value interface{}) string {
	switch v := value.(type) {
	case float64:
		return f.formatFloat64(v)
	case float32:
		return f.formatFloat32(v)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	case int32:
		return fmt.Sprintf("%d", v)
	case uint:
		return fmt.Sprintf("%d", v)
	case uint64:
		return fmt.Sprintf("%d", v)
	case uint32:
		return fmt.Sprintf("%d", v)
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatFloat64 格式化 float64
func (f *ValueFormatter) formatFloat64(v float64) string {
	format := fmt.Sprintf("%%.%df", f.floatPrecision)
	return fmt.Sprintf(format, v)
}

// formatFloat32 格式化 float32
func (f *ValueFormatter) formatFloat32(v float32) string {
	format := fmt.Sprintf("%%.%df", f.floatPrecision)
	return fmt.Sprintf(format, v)
}
