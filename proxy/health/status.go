package health

import (
	"time"
)

// ============ IP健康状态定义 ============

// HealthStatus IP健康状态枚举
type HealthStatus int

const (
	// StatusHealthy IP健康
	StatusHealthy HealthStatus = iota
	// StatusTemporaryFailed 临时失败（首次失败）
	StatusTemporaryFailed
	// StatusBlacklisted 已加入黑名单（连续失败）
	StatusBlacklisted
)

func (s HealthStatus) String() string {
	switch s {
	case StatusHealthy:
		return "健康"
	case StatusTemporaryFailed:
		return "临时失败"
	case StatusBlacklisted:
		return "黑名单"
	default:
		return "未知"
	}
}

// IPHealth IP健康信息
type IPHealth struct {
	IP                  string        // IP地址
	Status              HealthStatus  // 当前状态
	ConsecutiveFailures int           // 连续失败次数
	LastCheckTime       time.Time     // 最后检查时间
	LastError           string        // 最后一次错误信息
}

// IsHealthy 是否健康（可用于分配）
func (h *IPHealth) IsHealthy() bool {
	return h.Status == StatusHealthy
}

// IsBlacklisted 是否在黑名单中
func (h *IPHealth) IsBlacklisted() bool {
	return h.Status == StatusBlacklisted
}

// CanAssignToUser 是否可以分配给用户
func (h *IPHealth) CanAssignToUser() bool {
	return h.IsHealthy()
}

// Copy 返回IPHealth的副本（用于线程安全的状态查询）
func (h *IPHealth) Copy() *IPHealth {
	return &IPHealth{
		IP:                  h.IP,
		Status:              h.Status,
		ConsecutiveFailures: h.ConsecutiveFailures,
		LastCheckTime:       h.LastCheckTime,
		LastError:           h.LastError,
	}
}

// ============ 健康检查事件 ============

// IPFailureEvent IP失败事件
type IPFailureEvent struct {
	IP            string    // 失败的IP
	AffectedUsers []string  // 受影响的用户列表
	Reason        string    // 失败原因
	OccurredAt    time.Time // 发生时间
}

// IPRecoveryEvent IP恢复事件
type IPRecoveryEvent struct {
	IP          string    // 恢复的IP
	OccurredAt  time.Time // 发生时间
	DowntimeSec int       // 宕机时长（秒）
}
