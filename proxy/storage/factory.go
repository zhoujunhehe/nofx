package storage

import (
	"fmt"
	"log"
	"time"
)

// StorageType 存储类型
const (
	TypeFile  = "file"
	TypeRedis = "redis"
	TypeDB    = "db"
)

// Config 存储配置
type Config struct {
	Type          string        // 存储类型: "file", "redis", "db"
	FilePath      string        // 文件路径（file类型）
	FlushInterval time.Duration // 刷新间隔
	RedisAddr     string        // Redis地址（redis类型）
	RedisPassword string        // Redis密码
	RedisDB       int           // Redis数据库编号
	DBDriver      string        // 数据库驱动（db类型）
	DBDSN         string        // 数据库连接串
}

// CreateStorage 根据配置创建存储
func CreateStorage(cfg Config) (Storage, error) {
	// 默认存储类型
	storageType := cfg.Type
	if storageType == "" {
		storageType = TypeFile // 默认使用文件存储
	}

	switch storageType {
	case TypeFile:
		// 文件存储配置
		filePath := cfg.FilePath
		if filePath == "" {
			filePath = "data/proxy_user_mapping.json"
		}

		flushInterval := cfg.FlushInterval
		if flushInterval == 0 {
			flushInterval = 5 * time.Second
		}

		log.Printf("✓ 使用文件存储: %s (刷新间隔: %v)", filePath, flushInterval)

		return NewFileStorage(FileStorageConfig{
			FilePath:      filePath,
			FlushInterval: flushInterval,
			BufferSize:    100,
		}), nil

	case TypeRedis:
		// Redis存储配置
		if cfg.RedisAddr == "" {
			return nil, fmt.Errorf("Redis存储需要配置RedisAddr")
		}

		log.Printf("✓ 使用Redis存储: %s (DB: %d)", cfg.RedisAddr, cfg.RedisDB)

		return NewRedisStorage(RedisStorageConfig{
			Addr:      cfg.RedisAddr,
			Password:  cfg.RedisPassword,
			DB:        cfg.RedisDB,
			KeyPrefix: "proxy_user_mapping:",
		}), nil

	case TypeDB:
		// 数据库存储配置
		if cfg.DBDriver == "" || cfg.DBDSN == "" {
			return nil, fmt.Errorf("数据库存储需要配置DBDriver和DBDSN")
		}

		log.Printf("✓ 使用数据库存储: %s", cfg.DBDriver)

		return NewDBStorage(DBStorageConfig{
			Driver: cfg.DBDriver,
			DSN:    cfg.DBDSN,
			Table:  "proxy_user_mapping",
		}), nil

	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
	}
}
