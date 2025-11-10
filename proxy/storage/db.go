package storage

import "fmt"

// DBStorageConfig 数据库存储配置
type DBStorageConfig struct {
	Driver string
	DSN    string
	Table  string
}

// DBStorage 数据库存储实现（TODO: 未实现）
type DBStorage struct {
	config DBStorageConfig
}

// NewDBStorage 创建数据库存储
func NewDBStorage(config DBStorageConfig) *DBStorage {
	return &DBStorage{config: config}
}

// Load 加载映射数据（占位实现）
func (d *DBStorage) Load() (*MappingData, error) {
	return nil, fmt.Errorf("数据库存储未实现")
}

// Save 保存映射数据（占位实现）
func (d *DBStorage) Save(data *MappingData) error {
	return fmt.Errorf("数据库存储未实现")
}

// AddUserMapping 添加用户映射（占位实现）
func (d *DBStorage) AddUserMapping(userID, ip string) error {
	return fmt.Errorf("数据库存储未实现")
}

// RemoveUserMapping 删除用户映射（占位实现）
func (d *DBStorage) RemoveUserMapping(userID, ip string) error {
	return fmt.Errorf("数据库存储未实现")
}

// Flush 强制刷新（占位实现）
func (d *DBStorage) Flush() error {
	return nil // no-op
}

// Close 关闭存储（占位实现）
func (d *DBStorage) Close() error {
	return nil // no-op
}
