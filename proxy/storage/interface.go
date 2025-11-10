package storage

// MappingData 映射数据结构
type MappingData struct {
	UserToIP  map[string]string   `json:"user_to_ip"`
	IPToUsers map[string][]string `json:"ip_to_users"`
}

// Storage 存储接口（支持未来扩展Redis/DB）
type Storage interface {
	// Load 加载映射数据
	Load() (*MappingData, error)

	// Save 保存映射数据（可能是异步的，用于全量保存）
	Save(data *MappingData) error

	// 增量操作接口（性能优化）
	AddUserMapping(userID, ip string) error    // 添加用户映射
	RemoveUserMapping(userID, ip string) error // 删除用户映射

	// Flush 强制刷新到持久化存储
	Flush() error

	// Close 关闭存储（清理资源）
	Close() error
}
