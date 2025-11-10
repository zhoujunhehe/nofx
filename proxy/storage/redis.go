package storage

import "fmt"

// RedisStorageConfig Redis存储配置
type RedisStorageConfig struct {
	Addr      string
	Password  string
	DB        int
	KeyPrefix string
}

// RedisStorage Redis存储实现（TODO: 未实现）
type RedisStorage struct {
	config RedisStorageConfig
}

// NewRedisStorage 创建Redis存储
func NewRedisStorage(config RedisStorageConfig) *RedisStorage {
	return &RedisStorage{config: config}
}

// Load 加载映射数据（占位实现）
func (r *RedisStorage) Load() (*MappingData, error) {
	return nil, fmt.Errorf("Redis存储未实现")
}

// Save 保存映射数据（占位实现）
func (r *RedisStorage) Save(data *MappingData) error {
	return fmt.Errorf("Redis存储未实现")
}

// AddUserMapping 添加用户映射（占位实现）
func (r *RedisStorage) AddUserMapping(userID, ip string) error {
	return fmt.Errorf("Redis存储未实现")
}

// RemoveUserMapping 删除用户映射（占位实现）
func (r *RedisStorage) RemoveUserMapping(userID, ip string) error {
	return fmt.Errorf("Redis存储未实现")
}

// Flush 强制刷新（占位实现）
func (r *RedisStorage) Flush() error {
	return nil // no-op
}

// Close 关闭存储（占位实现）
func (r *RedisStorage) Close() error {
	return nil // no-op
}
