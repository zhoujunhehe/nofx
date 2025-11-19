package logger

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// TelegramHook 实现logrus.Hook接口，将日志推送到Telegram
type TelegramHook struct {
	sender  *TelegramSender
	levels  []logrus.Level
	enabled bool
}

// NewTelegramHook 创建Telegram Hook
func NewTelegramHook(config *TelegramConfig) (*TelegramHook, error) {
	if !config.Enabled {
		return &TelegramHook{enabled: false}, nil
	}

	if config.BotToken == "" || config.ChatID == 0 {
		return nil, fmt.Errorf("telegram配置不完整: bot_token和chat_id不能为空")
	}

	// 创建发送器（使用默认参数）
	sender, err := NewTelegramSender(config.BotToken, config.ChatID)
	if err != nil {
		return nil, fmt.Errorf("创建telegram发送器失败: %w", err)
	}

	hook := &TelegramHook{
		sender:  sender,
		levels:  config.GetLogrusLevels(),
		enabled: true,
	}

	return hook, nil
}

// Levels 返回需要触发的日志级别
func (h *TelegramHook) Levels() []logrus.Level {
	if !h.enabled {
		return []logrus.Level{}
	}
	return h.levels
}

// Fire 当日志触发时调用
func (h *TelegramHook) Fire(entry *logrus.Entry) error {
	if !h.enabled {
		return nil
	}

	// 格式化消息
	message := h.formatMessage(entry)

	// 异步发送（非阻塞）
	h.sender.SendAsync(message)

	return nil
}

// formatMessage 格式化日志消息为Telegram格式 (HTML)
func (h *TelegramHook) formatMessage(entry *logrus.Entry) string {
	// 级别emoji
	levelEmoji := h.getLevelEmoji(entry.Level)

	// 基本信息
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s Level: <code>%s</code>\n", levelEmoji, strings.ToUpper(entry.Level.String())))
	builder.WriteString(fmt.Sprintf("📝 Msg: <code>%s</code>\n", htmlEscape(entry.Message)))

	// 字段信息
	if len(entry.Data) > 0 {
		builder.WriteString("📊 字段:\n")
		for key, value := range entry.Data {
			builder.WriteString(fmt.Sprintf("  • %s: <code>%v</code>\n", htmlEscape(key), htmlEscape(fmt.Sprintf("%v", value))))
		}
	}

	// 时间戳
	builder.WriteString(fmt.Sprintf("🕐 Time: <code>%s</code>", entry.Time.Format("2006-01-02 15:04:05")))

	return builder.String()
}

// getLevelEmoji 获取日志级别对应的emoji
func (h *TelegramHook) getLevelEmoji(level logrus.Level) string {
	switch level {
	case logrus.PanicLevel:
		return "🔴"
	case logrus.FatalLevel:
		return "🔴"
	case logrus.ErrorLevel:
		return "🟠"
	case logrus.WarnLevel:
		return "⚠️"
	case logrus.InfoLevel:
		return "🟢"
	case logrus.DebugLevel:
		return "🔵"
	default:
		return "⚪"
	}
}

// htmlEscape 转义HTML特殊字符（只需转义少量字符）
func htmlEscape(text string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(text)
}

// Stop 停止Hook（优雅关闭）
func (h *TelegramHook) Stop() {
	if h.enabled && h.sender != nil {
		h.sender.Stop()
	}
}
