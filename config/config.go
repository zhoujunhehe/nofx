package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

// LeverageConfig 杠杆配置
type LeverageConfig struct {
	BTCETHLeverage  int `json:"btc_eth_leverage"` // BTC和ETH的杠杆倍数（主账户建议5-50，子账户≤5）
	AltcoinLeverage int `json:"altcoin_leverage"` // 山寨币的杠杆倍数（主账户建议5-20，子账户≤5）
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string          `json:"level"`    // 日志级别: debug, info, warn, error (默认: info)
	Telegram *TelegramConfig `json:"telegram"` // Telegram推送配置（可选）
}

// TelegramConfig Telegram推送配置（简化版，只保留必需字段）
type TelegramConfig struct {
	Enabled  bool   `json:"enabled"`   // 是否启用（默认: false）
	BotToken string `json:"bot_token"` // Bot Token
	ChatID   int64  `json:"chat_id"`   // Chat ID
	MinLevel string `json:"min_level"` // 最低日志级别，该级别及以上的日志会推送到Telegram（可选，默认: error）
}

// Config 总配置
type Config struct {
	BetaMode           bool           `json:"beta_mode"`
	APIServerPort      int            `json:"api_server_port"`
	UseDefaultCoins    bool           `json:"use_default_coins"`
	DefaultCoins       []string       `json:"default_coins"`
	CoinPoolAPIURL     string         `json:"coin_pool_api_url"`
	OITopAPIURL        string         `json:"oi_top_api_url"`
	MaxDailyLoss       float64        `json:"max_daily_loss"`
	MaxDrawdown        float64        `json:"max_drawdown"`
	StopTradingMinutes int            `json:"stop_trading_minutes"`
	Leverage           LeverageConfig `json:"leverage"`     // 杠杆配置
	JWTSecret          string         `json:"jwt_secret"`
	DataKLineTime      string         `json:"data_k_line_time"`
	Proxy              *ProxyConfig   `json:"proxy"` // HTTP 代理配置（可选）
	Log                *LogConfig     `json:"log"`   // 日志配置
}

// ProxyConfig HTTP 代理配置
type ProxyConfig struct {
	Enabled            bool     `json:"enabled"`              // 是否启用代理
	Mode               string   `json:"mode"`                 // 模式: "single", "pool", "brightdata"
	Timeout            int      `json:"timeout"`              // 超时时间（秒）
	ProxyURL           string   `json:"proxy_url"`            // 单个代理地址
	ProxyList          []string `json:"proxy_list"`           // 代理列表
	BrightDataEndpoint string   `json:"brightdata_endpoint"`  // Bright Data接口地址
	BrightDataToken    string   `json:"brightdata_token"`     // Bright Data访问令牌
	BrightDataZone     string   `json:"brightdata_zone"`      // Bright Data区域
	ProxyHost          string   `json:"proxy_host"`           // 代理主机
	ProxyUser          string   `json:"proxy_user"`           // 代理用户名模板
	ProxyPassword      string   `json:"proxy_password"`       // 代理密码
	RefreshInterval    int      `json:"refresh_interval"`     // 刷新间隔（秒）
	BlacklistTTL       int      `json:"blacklist_ttl"`        // 黑名单TTL
}
// LoadConfig 从文件加载配置
func LoadConfig(filename string) (*Config, error) {
	// 检查filename是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("📄 %s不存在，使用默认配置", filename)
		return &Config{}, nil
	}

	// 读取 filename
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取%s失败: %w", filename, err)
	}

	// 解析JSON
	var configFile Config
	if err := json.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("解析%s失败: %w", filename, err)
	}

	return &configFile, nil
}
