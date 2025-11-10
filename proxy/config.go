package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Config 代理配置
type Config struct {
	Enabled            bool          // 是否启用代理
	Mode               string        // 模式: "single", "pool", "brightdata"
	Timeout            time.Duration // 超时时间
	ProxyURL           string        // 单个代理地址 (single模式)
	ProxyList          []string      // 代理列表 (pool模式)
	BrightDataEndpoint string        // Bright Data接口地址 (brightdata模式)
	BrightDataToken    string        // Bright Data访问令牌 (brightdata模式)
	BrightDataZone     string        // Bright Data区域 (brightdata模式)
	ProxyHost          string        // 代理主机
	ProxyUser          string        // 代理用户名模板（支持%s占位符）
	ProxyPassword      string        // 代理密码
	RefreshInterval    time.Duration // IP列表刷新间隔

	// 用户IP映射存储配置
	MappingFile   string        // 文件存储路径（StorageType=file时使用）
	StorageType   string        // 存储类型: "file"(默认), "redis", "db"
	FlushInterval time.Duration // 刷新间隔（默认5秒）
	RedisAddr     string        // Redis地址（StorageType=redis时使用）
	RedisPassword string        // Redis密码
	RedisDB       int           // Redis数据库编号
	DBDriver      string        // 数据库驱动（StorageType=db时使用）
	DBDSN         string        // 数据库连接串

	// 健康检查配置
	HealthCheckURLs             []string      // 探活URL列表（对每个IP测试这些URL）
	HealthCheckInterval         time.Duration // 探活间隔（默认5秒）
	HealthCheckTimeout          time.Duration // 单次探活超时（默认3秒）
	HealthCheckFailureThreshold int           // 连续失败阈值（默认5次）
}

// ProxyConfigFile JSON配置文件结构
type ProxyConfigFile struct {
	Enabled            bool     `json:"enabled"`             // 是否启用代理
	Mode               string   `json:"mode"`                // 模式: "single", "pool", "brightdata"
	Timeout            int      `json:"timeout"`             // 超时时间（秒）
	ProxyURL           string   `json:"proxy_url"`           // 单个代理地址
	ProxyList          []string `json:"proxy_list"`          // 代理列表
	BrightDataEndpoint string   `json:"brightdata_endpoint"` // Bright Data接口地址
	BrightDataToken    string   `json:"brightdata_token"`    // Bright Data访问令牌
	BrightDataZone     string   `json:"brightdata_zone"`     // Bright Data区域
	ProxyHost          string   `json:"proxy_host"`          // 代理主机
	ProxyUser          string   `json:"proxy_user"`          // 代理用户名模板
	ProxyPassword      string   `json:"proxy_password"`      // 代理密码
	RefreshInterval    int      `json:"refresh_interval"`    // 刷新间隔（秒）

	// 用户IP映射存储配置
	MappingFile   string `json:"mapping_file"`   // 文件存储路径（StorageType=file时使用）
	StorageType   string `json:"storage_type"`   // 存储类型: "file"(默认), "redis", "db"
	FlushInterval int    `json:"flush_interval"` // 刷新间隔（秒，默认5秒）
	RedisAddr     string `json:"redis_addr"`     // Redis地址（StorageType=redis时使用）
	RedisPassword string `json:"redis_password"` // Redis密码
	RedisDB       int    `json:"redis_db"`       // Redis数据库编号
	DBDriver      string `json:"db_driver"`      // 数据库驱动（StorageType=db时使用）
	DBDSN         string `json:"db_dsn"`         // 数据库连接串

	// 健康检查配置
	HealthCheckURLs             []string `json:"health_check_urls"`              // 探活URL列表
	HealthCheckInterval         int      `json:"health_check_interval"`          // 探活间隔（秒，默认5秒）
	HealthCheckTimeout          int      `json:"health_check_timeout"`           // 单次探活超时（秒，默认3秒）
	HealthCheckFailureThreshold int      `json:"health_check_failure_threshold"` // 连续失败阈值（默认5次）
}

// ConfigFile 配置文件结构
type ConfigFile struct {
	Proxy *ProxyConfigFile `json:"proxy"`
}

// ConvertToConfig 将文件配置转换为内部配置
func (c *ConfigFile) ConvertToConfig() *Config {
	if c.Proxy == nil {
		return &Config{}
	}
	return &Config{
		Enabled:            c.Proxy.Enabled,
		Mode:               c.Proxy.Mode,
		Timeout:            time.Duration(c.Proxy.Timeout) * time.Second,
		ProxyURL:           c.Proxy.ProxyURL,
		ProxyList:          c.Proxy.ProxyList,
		BrightDataEndpoint: c.Proxy.BrightDataEndpoint,
		BrightDataToken:    c.Proxy.BrightDataToken,
		BrightDataZone:     c.Proxy.BrightDataZone,
		ProxyHost:          c.Proxy.ProxyHost,
		ProxyUser:          c.Proxy.ProxyUser,
		ProxyPassword:      c.Proxy.ProxyPassword,
		RefreshInterval:    time.Duration(c.Proxy.RefreshInterval) * time.Second,

		MappingFile:   c.Proxy.MappingFile,
		StorageType:   c.Proxy.StorageType,
		FlushInterval: time.Duration(c.Proxy.FlushInterval) * time.Second,
		RedisAddr:     c.Proxy.RedisAddr,
		RedisPassword: c.Proxy.RedisPassword,
		RedisDB:       c.Proxy.RedisDB,
		DBDriver:      c.Proxy.DBDriver,
		DBDSN:         c.Proxy.DBDSN,

		HealthCheckURLs:             c.Proxy.HealthCheckURLs,
		HealthCheckInterval:         time.Duration(c.Proxy.HealthCheckInterval) * time.Second,
		HealthCheckTimeout:          time.Duration(c.Proxy.HealthCheckTimeout) * time.Second,
		HealthCheckFailureThreshold: c.Proxy.HealthCheckFailureThreshold,
	}
}

// LoadConfig 从文件加载配置
// TODO：此处为临时政策，未来会统一到主配置加载中
func LoadConfig(filename string) (*ConfigFile, error) {
	// 检查filename是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		log.Printf("📄 %s不存在，使用默认配置", filename)
		return &ConfigFile{}, nil
	}

	// 读取 filename
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取%s失败: %w", filename, err)
	}

	// 解析JSON
	var configFile ConfigFile
	if err := json.Unmarshal(data, &configFile); err != nil {
		return nil, fmt.Errorf("解析%s失败: %w", filename, err)
	}

	return &configFile, nil
}
