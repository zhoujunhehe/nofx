package prompt

import (
	"fmt"
	"strings"
)

// ReplacePlaceholder 替换占位符
func ReplacePlaceholder(text, placeholder, value string) string {
	return strings.ReplaceAll(text, placeholder, value)
}

// ProcessCondition 处理条件语句
// 简化的条件处理：支持 {if condition}...{else}...{endif}
func ProcessCondition(text, condition string, result bool) string {
	ifPattern := fmt.Sprintf("{if %s}", condition)
	elsePattern := "{else}"
	endifPattern := "{endif}"

	// 找到if...else...endif块
	ifIdx := strings.Index(text, ifPattern)
	if ifIdx == -1 {
		return text
	}

	elseIdx := strings.Index(text[ifIdx:], elsePattern)
	endifIdx := strings.Index(text[ifIdx:], endifPattern)

	if endifIdx == -1 {
		return text // 没有找到endif，返回原文
	}

	if result {
		// 取if块的内容
		if elseIdx != -1 && elseIdx < endifIdx {
			// 有else块，取if到else之间的内容
			content := text[ifIdx+len(ifPattern) : ifIdx+elseIdx]
			return text[:ifIdx] + content + text[ifIdx+endifIdx+len(endifPattern):]
		}
		// 没有else块，取if到endif之间的内容
		content := text[ifIdx+len(ifPattern) : ifIdx+endifIdx]
		return text[:ifIdx] + content + text[ifIdx+endifIdx+len(endifPattern):]
	}

	// 取else块的内容
	if elseIdx != -1 && elseIdx < endifIdx {
		content := text[ifIdx+elseIdx+len(elsePattern) : ifIdx+endifIdx]
		return text[:ifIdx] + content + text[ifIdx+endifIdx+len(endifPattern):]
	}

	// 没有else块，删除整个if块
	return text[:ifIdx] + text[ifIdx+endifIdx+len(endifPattern):]
}

// UniqueStrings 字符串数组去重
func UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// GetFloat64 从map中安全获取float64值
func GetFloat64(data map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultValue
}

// GetString 从map中安全获取string值
func GetString(data map[string]interface{}, key string, defaultValue string) string {
	if val, ok := data[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return defaultValue
}

// GetInt 从map中安全获取int值
func GetInt(data map[string]interface{}, key string, defaultValue int) int {
	if val, ok := data[key]; ok {
		if i, ok := val.(int); ok {
			return i
		}
		// 尝试从float64转换
		if f, ok := val.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}
