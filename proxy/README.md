# HTTP 代理模块

## 概述

这是一个企业级HTTP代理管理模块，专为高频API请求场景设计，特别适用于币安等交易所API的访问控制。支持单代理、代理池两种模式，提供用户-IP绑定、健康检查、持久化存储等完整的生产级特性。

## 功能特性

- ✅ **多种工作模式**：单代理模式、固定代理池模式
- ✅ **用户-IP绑定**：支持用户与代理IP的稳定绑定，避免频繁切换触发风控
- ✅ **负载均衡**：最少用户优先（Least Loaded）算法，均匀分配用户到各个IP
- ✅ **健康检查**：自动检测IP健康状态，失败IP自动隔离并尝试恢复
- ✅ **持久化存储**：支持文件、Redis、数据库多种存储后端（当前实现文件存储）
- ✅ **线程安全**：所有操作使用读写锁保护，支持高并发访问
- ✅ **自动刷新**：支持定时刷新代理IP列表
- ✅ **降级策略**：IP不足时自动降级，确保服务可用
- ✅ **可选启用**：禁用时自动使用直连，不影响业务代码

## 架构设计

```
proxy/
├── README.md                    # 本文档
├── types.go                     # 核心数据结构定义
├── config.go                    # 配置结构和加载
├── init.go                      # 全局管理器初始化
├── manager.go                   # 代理管理器（核心协调层）
├── client.go                    # HTTP客户端封装
├── proxy.go                     # 公共API导出
├── manager_test.go              # 管理器测试
├── concurrent_test.go           # 并发测试
├── health/                      # 健康检查模块
│   ├── checker.go               # 健康检查器实现
│   ├── checker_test.go          # 健康检查测试
│   └── types.go                 # 健康检查类型定义
├── mapper/                      # 用户IP映射模块
│   ├── mapper.go                # 用户-IP映射逻辑
│   ├── mapper_test.go           # 映射器测试
│   └── types.go                 # 映射类型定义
├── provider/                    # IP提供者模块
│   ├── provider.go              # IP提供者接口和实现
│   ├── provider_test.go         # Provider测试
│   └── types.go                 # Provider类型定义
└── storage/                     # 存储模块
    ├── storage.go               # 存储接口
    ├── file.go                  # 文件存储实现
    ├── memory.go                # 内存存储实现（测试用）
    ├── storage_test.go          # 存储测试
    └── types.go                 # 存储类型定义
```

### 设计原则

1. **分层架构**：Manager → Health/Mapper/Provider/Storage，职责清晰
2. **接口抽象**：通过接口实现不同Provider和Storage的统一管理
3. **单例模式**：全局ProxyManager确保资源统一管理
4. **防御性编程**：多层边界检查，降级策略，优雅处理异常
5. **测试覆盖**：32+测试用例，覆盖核心功能和并发场景

## 配置说明

在 `config.json` 中添加 `proxy` 配置段：

```json
{
  "proxy": {
    "enabled": true,
    "mode": "pool",
    "timeout": 30,
    "proxy_url": "",
    "proxy_list": [
      "http://proxy1.example.com:8080",
      "http://proxy2.example.com:8080"
    ],
    "proxy_host": "proxy.example.com:8080",
    "proxy_user": "user-%s",
    "proxy_password": "your_password",
    "refresh_interval": 1800,

    "mapping_file": "data/user_ip_mapping.json",
    "storage_type": "file",
    "flush_interval": 5,

    "health_check_urls": [
      "https://fapi.binance.com/fapi/v1/ping"
    ],
    "health_check_interval": 30,
    "health_check_timeout": 5,
    "health_check_failure_threshold": 3
  }
}
```

### 配置字段详解

#### 基础配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `enabled` | bool | 是 | 是否启用代理（false时使用直连） |
| `mode` | string | 是 | 代理模式：`single`（单代理）、`pool`（代理池） |
| `timeout` | int | 否 | HTTP请求超时时间（秒），默认30 |

#### 代理源配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `proxy_url` | string | single模式必填 | 单个代理地址，如 `http://127.0.0.1:7890` |
| `proxy_list` | []string | pool模式必填 | 代理列表，支持 `http://`、`https://`、`socks5://` |
| `proxy_host` | string | 否 | 代理主机（用于认证代理） |
| `proxy_user` | string | 否 | 代理用户名模板，支持 `%s` 占位符替换IP |
| `proxy_password` | string | 否 | 代理密码 |
| `refresh_interval` | int | 否 | IP列表刷新间隔（秒），默认1800（30分钟） |

#### 用户-IP映射存储配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mapping_file` | string | 否 | 文件存储路径，默认为空（不持久化） |
| `storage_type` | string | 否 | 存储类型：`file`（默认）、`redis`、`db` |
| `flush_interval` | int | 否 | 刷新间隔（秒），默认5秒 |
| `redis_addr` | string | redis模式必填 | Redis地址，如 `localhost:6379` |
| `redis_password` | string | 否 | Redis密码 |
| `redis_db` | int | 否 | Redis数据库编号，默认0 |
| `db_driver` | string | db模式必填 | 数据库驱动：`mysql`、`postgres` |
| `db_dsn` | string | db模式必填 | 数据库连接串 |

#### 健康检查配置

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `health_check_urls` | []string | 否 | 探活URL列表，为空时禁用健康检查 |
| `health_check_interval` | int | 否 | 探活间隔（秒），默认30秒 |
| `health_check_timeout` | int | 否 | 单次探活超时（秒），默认5秒 |
| `health_check_failure_threshold` | int | 否 | 连续失败阈值，默认3次 |

## 使用方法

### 1. 初始化代理管理器

在 `main.go` 或初始化代码中：

```go
import (
    "nofx/proxy"
    "time"
)

// 使用配置结构体初始化
proxyConfig := &proxy.Config{
    Enabled: true,
    Mode: "pool",
    Timeout: 30 * time.Second,
    ProxyList: []string{
        "http://proxy1.example.com:8080",
        "http://proxy2.example.com:8080",
    },
    ProxyHost: "proxy.example.com:8080",
    ProxyUser: "user-%s",  // %s会被替换为IP
    ProxyPassword: "password",

    // 持久化配置
    MappingFile: "data/user_ip_mapping.json",
    StorageType: "file",
    FlushInterval: 5 * time.Second,

    // 健康检查配置
    HealthCheckURLs: []string{
        "https://fapi.binance.com/fapi/v1/ping",
    },
    HealthCheckInterval: 30 * time.Second,
    HealthCheckTimeout: 5 * time.Second,
    HealthCheckFailureThreshold: 3,
}

err := proxy.InitGlobalProxyManager(proxyConfig)
if err != nil {
    log.Fatalf("初始化代理管理器失败: %v", err)
}
```

### 2. 获取代理HTTP客户端（不绑定用户）

适用于一次性请求或无需用户绑定的场景：

```go
// 获取代理客户端（随机分配）
proxyClient, err := proxy.GetProxyHTTPClient()
if err != nil {
    log.Printf("获取代理客户端失败: %v", err)
    return
}

// 使用代理客户端发送请求
resp, err := proxyClient.Client.Get("https://fapi.binance.com/fapi/v1/ticker/24hr")
if err != nil {
    // 请求失败，将此代理加入黑名单
    proxy.AddBlacklist(proxyClient.IP)
    log.Printf("请求失败，代理IP %s 已加入黑名单", proxyClient.IP)
    return
}
defer resp.Body.Close()

// 处理响应...
log.Printf("请求成功，使用代理: %s", proxyClient.IP)
```

### 3. 获取用户绑定的代理客户端（推荐）

适用于需要用户-IP稳定绑定的场景（如交易账户）：

```go
// 为特定用户获取代理客户端（保证同一用户总是使用相同IP）
userID := "user_123"
proxyClient, err := proxy.GetGlobalProxyManager().GetProxyClientForUser(userID)
if err != nil {
    log.Printf("获取用户代理失败: %v", err)
    return
}

// 使用代理发送请求
resp, err := proxyClient.Client.Get("https://fapi.binance.com/fapi/v1/account")
if err != nil {
    // 请求失败，将IP加入黑名单
    // 健康检查器会自动为该用户重新分配健康的IP
    proxy.AddBlacklist(proxyClient.IP)
    log.Printf("用户 %s 的代理IP %s 请求失败，已加入黑名单", userID, proxyClient.IP)
    return
}
defer resp.Body.Close()

log.Printf("用户 %s 使用代理 %s 请求成功", userID, proxyClient.IP)
```

### 4. 黑名单管理

```go
// 添加失败的代理IP到黑名单
proxy.AddBlacklist("192.168.1.1")

// 获取黑名单状态
total, blacklisted, available := proxy.GetGlobalProxyManager().GetBlacklistStatus()
log.Printf("代理状态: 总计%d个，黑名单%d个，可用%d个", total, blacklisted, available)
```

### 5. 用户-IP映射管理

```go
manager := proxy.GetGlobalProxyManager()

// 获取某个IP绑定的所有用户
users := manager.GetUsersForIP("192.168.1.1")
log.Printf("IP 192.168.1.1 绑定的用户: %v", users)

// 获取所有用户-IP映射
allMappings := manager.GetAllUserIPMappings()
for ip, users := range allMappings {
    log.Printf("IP %s: %d 个用户", ip, len(users))
}

// 获取映射统计
totalUsers, totalIPs, avgUsersPerIP := manager.GetMappingStats()
log.Printf("总用户: %d, 总IP: %d, 平均每IP用户数: %.2f",
    totalUsers, totalIPs, avgUsersPerIP)

// 手动解除用户绑定
err := manager.RemoveUserMapping("user_123")
if err != nil {
    log.Printf("解除绑定失败: %v", err)
}

// 手动保存映射到存储
err = manager.SaveMappings()
if err != nil {
    log.Printf("保存映射失败: %v", err)
}
```

### 6. 手动刷新IP列表

```go
err := proxy.RefreshIPList()
if err != nil {
    log.Printf("刷新IP列表失败: %v", err)
}
```

### 7. 检查代理是否启用

```go
if proxy.IsEnabled() {
    log.Println("代理已启用")
} else {
    log.Println("代理未启用，使用直连")
}
```

## 工作模式详解

### Mode 1: Single（单代理模式）

适用场景：本地代理工具（如Clash、V2Ray）或单个固定代理服务器

```json
{
  "proxy": {
    "enabled": true,
    "mode": "single",
    "proxy_url": "http://127.0.0.1:7890",
    "timeout": 30
  }
}
```

特点：
- ✅ 简单直接，适合本地开发和测试
- ✅ 所有请求通过同一个代理
- ✅ 不需要健康检查和IP轮换
- ⚠️  单点故障风险，代理失败则所有请求失败

### Mode 2: Pool（代理池模式）

适用场景：拥有多个固定代理服务器，需要负载均衡和用户-IP绑定

```json
{
  "proxy": {
    "enabled": true,
    "mode": "pool",
    "proxy_list": [
      "http://proxy1.example.com:8080",
      "http://proxy2.example.com:8080",
      "http://proxy3.example.com:8080"
    ],
    "proxy_host": "proxy.example.com:8080",
    "proxy_user": "user-%s",
    "proxy_password": "password",
    "timeout": 30,

    "mapping_file": "data/user_ip_mapping.json",
    "storage_type": "file",

    "health_check_urls": [
      "https://fapi.binance.com/fapi/v1/ping"
    ],
    "health_check_interval": 30,
    "health_check_timeout": 5,
    "health_check_failure_threshold": 3
  }
}
```

特点：
- ✅ 支持多协议：HTTP、HTTPS、SOCKS5
- ✅ 用户-IP稳定绑定，同一用户总是使用相同IP
- ✅ 负载均衡：最少用户优先算法，均匀分配
- ✅ 自动健康检查：失败IP自动隔离，并尝试恢复
- ✅ 持久化存储：重启后用户-IP绑定关系保持不变
- ✅ 降级策略：IP不足时自动降级，确保服务可用

**代理认证模板说明**：

`proxy_user` 支持 `%s` 占位符，会被替换为IP地址：

```
proxy_user: "user-%s"
proxy_host: "proxy.example.com:8080"
IP: "192.168.1.1"

实际生成的代理URL:
http://user-192.168.1.1:password@proxy.example.com:8080
```

这种模式常用于Bright Data、Luminati等支持IP粒度认证的代理服务。

## 核心API

### 全局函数

```go
// 初始化全局代理管理器（只执行一次，使用sync.Once保护）
func InitGlobalProxyManager(config *Config) error

// 获取全局代理管理器实例
func GetGlobalProxyManager() *ProxyManager

// 获取代理HTTP客户端（随机分配，不绑定用户）
func GetProxyHTTPClient() (*ProxyClient, error)

// 获取HTTP Transport（用于自定义HTTP客户端）
func GetTransport() *http.Transport

// 将代理IP添加到黑名单
func AddBlacklist(ip string)

// 刷新IP列表
func RefreshIPList() error

// 检查代理是否启用
func IsEnabled() bool
```

### ProxyManager 方法

#### 代理客户端获取

```go
// 获取代理客户端（随机分配）
func (m *ProxyManager) GetProxyClient() (*ProxyClient, error)

// 获取用户绑定的代理客户端（推荐）
func (m *ProxyManager) GetProxyClientForUser(userID string) (*ProxyClient, error)
```

#### 黑名单管理

```go
// 添加IP到黑名单
func (m *ProxyManager) AddBlacklist(ip string)

// 获取黑名单状态
func (m *ProxyManager) GetBlacklistStatus() (total, blacklisted, available int)
```

#### 用户-IP映射管理

```go
// 获取某个IP绑定的所有用户
func (m *ProxyManager) GetUsersForIP(ip string) []string

// 获取所有用户-IP映射
func (m *ProxyManager) GetAllUserIPMappings() map[string][]string

// 获取映射统计信息
func (m *ProxyManager) GetMappingStats() (totalUsers, totalIPs int, avgUsersPerIP float64)

// 移除用户映射
func (m *ProxyManager) RemoveUserMapping(userID string) error

// 手动保存映射到存储
func (m *ProxyManager) SaveMappings() error

// 获取UserIPMapper实例（高级用法）
func (m *ProxyManager) GetUserIPMapper() *mapper.UserIPMapper
```

#### IP列表管理

```go
// 刷新IP列表
func (m *ProxyManager) RefreshIPList() error

// 启动自动刷新
func (m *ProxyManager) StartAutoRefresh()

// 停止自动刷新（支持重复调用）
func (m *ProxyManager) StopAutoRefresh()
```

### 数据结构

```go
// ProxyClient 代理客户端
type ProxyClient struct {
    Client  *http.Client  // HTTP客户端
    ProxyID int           // 代理ID（-1表示直连）
    IP      string        // 代理IP
}

// Config 代理配置
type Config struct {
    Enabled            bool
    Mode               string
    Timeout            time.Duration
    ProxyURL           string
    ProxyList          []string
    ProxyHost          string
    ProxyUser          string
    ProxyPassword      string
    RefreshInterval    time.Duration

    // 存储配置
    MappingFile       string
    StorageType       string
    FlushInterval     time.Duration
    RedisAddr         string
    RedisPassword     string
    RedisDB           int
    DBDriver          string
    DBDSN             string

    // 健康检查配置
    HealthCheckURLs             []string
    HealthCheckInterval         time.Duration
    HealthCheckTimeout          time.Duration
    HealthCheckFailureThreshold int
}
```

## 健康检查机制

### 工作原理

健康检查器会定期对每个代理IP进行健康探测，自动隔离失败的IP并尝试恢复。

#### 1. 健康状态定义

```go
const (
    StatusHealthy          = "healthy"           // 健康
    StatusTemporaryFailed  = "temporary_failed"  // 临时失败
    StatusBlacklisted      = "blacklisted"       // 已加入黑名单
)
```

#### 2. 状态转换流程

```
初始状态: Healthy
    ↓ (连续失败达到阈值)
临时失败: TemporaryFailed
    ↓ (手动添加黑名单)
黑名单: Blacklisted
    ↓ (健康检查成功)
恢复: Healthy
```

#### 3. 自动降级与恢复

- **降级**：当IP连续失败次数达到阈值（默认3次），健康检查器会通知Manager，Manager会为绑定该IP的用户重新分配健康的IP
- **恢复**：黑名单中的IP如果健康检查成功，会被自动移出黑名单，重新可用
- **用户重分配**：IP失败后，绑定该IP的用户会自动迁移到其他健康IP上

#### 4. 配置示例

```go
config := &proxy.Config{
    HealthCheckURLs: []string{
        "https://fapi.binance.com/fapi/v1/ping",  // 主要检测URL
        "https://www.google.com",                  // 备用检测URL
    },
    HealthCheckInterval: 30 * time.Second,         // 每30秒检查一次
    HealthCheckTimeout: 5 * time.Second,           // 单次请求超时5秒
    HealthCheckFailureThreshold: 3,                // 连续失败3次触发降级
}
```

### 黑名单机制

#### 工作原理

1. **手动添加**：调用 `AddBlacklist(ip)` 手动将IP加入黑名单
2. **自动移除**：健康检查器定期探测黑名单中的IP，一旦检测成功，自动移出黑名单
3. **用户保护**：IP加入黑名单后，绑定该IP的用户会被自动重新分配到健康IP

#### 线程安全保证

```go
// 添加黑名单使用健康检查器协调
func (m *ProxyManager) AddBlacklist(ip string) {
    if m.healthChecker != nil {
        m.healthChecker.MarkAsBlacklisted(ip)
    }
    log.Printf("🚫 IP %s 已加入黑名单", ip)
}

// 所有状态访问都使用读写锁保护
func (h *HealthChecker) GetHealthStatus() map[string]HealthStatus {
    h.mu.RLock()
    defer h.mu.RUnlock()
    // ... 读取操作
}
```

#### 状态查询

```go
manager := proxy.GetGlobalProxyManager()

// 获取黑名单状态
total, blacklisted, available := manager.GetBlacklistStatus()
log.Printf("总IP: %d, 黑名单: %d, 可用: %d", total, blacklisted, available)

// 输出示例:
// 总IP: 5, 黑名单: 2, 可用: 3
```

## 完整使用示例

### 示例1：币安API请求（单代理模式）

```go
package main

import (
    "log"
    "nofx/proxy"
    "time"
)

func main() {
    // 初始化代理
    err := proxy.InitGlobalProxyManager(&proxy.Config{
        Enabled: true,
        Mode: "single",
        ProxyURL: "http://127.0.0.1:7890",
        Timeout: 30 * time.Second,
    })
    if err != nil {
        log.Fatalf("初始化代理失败: %v", err)
    }

    // 获取币安数据
    proxyClient, err := proxy.GetProxyHTTPClient()
    if err != nil {
        log.Fatalf("获取代理客户端失败: %v", err)
    }

    resp, err := proxyClient.Client.Get("https://fapi.binance.com/fapi/v1/ticker/24hr")
    if err != nil {
        log.Printf("请求失败: %v", err)
        return
    }
    defer resp.Body.Close()

    log.Printf("请求成功，使用代理: %s", proxyClient.IP)
}
```

### 示例2：多用户交易系统（代理池 + 用户绑定）

```go
package main

import (
    "fmt"
    "log"
    "nofx/proxy"
    "time"
)

func main() {
    // 初始化代理池，启用健康检查和持久化
    err := proxy.InitGlobalProxyManager(&proxy.Config{
        Enabled: true,
        Mode: "pool",
        ProxyList: []string{
            "http://proxy1.example.com:8080",
            "http://proxy2.example.com:8080",
            "http://proxy3.example.com:8080",
        },
        ProxyHost: "proxy.example.com:8080",
        ProxyUser: "user-%s",
        ProxyPassword: "password",
        Timeout: 30 * time.Second,

        // 持久化配置
        MappingFile: "data/user_ip_mapping.json",
        StorageType: "file",
        FlushInterval: 5 * time.Second,

        // 健康检查配置
        HealthCheckURLs: []string{
            "https://fapi.binance.com/fapi/v1/ping",
        },
        HealthCheckInterval: 30 * time.Second,
        HealthCheckTimeout: 5 * time.Second,
        HealthCheckFailureThreshold: 3,

        // 自动刷新
        RefreshInterval: 30 * time.Minute,
    })
    if err != nil {
        log.Fatalf("初始化代理失败: %v", err)
    }

    // 启动自动刷新
    proxy.GetGlobalProxyManager().StartAutoRefresh()

    // 模拟多个用户同时交易
    users := []string{"alice", "bob", "charlie", "david", "eve"}

    for _, userID := range users {
        go func(uid string) {
            for i := 0; i < 10; i++ {
                if err := placeOrder(uid, "BTCUSDT"); err != nil {
                    log.Printf("❌ 用户 %s 下单失败: %v", uid, err)
                } else {
                    log.Printf("✅ 用户 %s 下单成功", uid)
                }
                time.Sleep(2 * time.Second)
            }
        }(userID)
    }

    // 定期打印状态
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        printStatus()
    }
}

func placeOrder(userID, symbol string) error {
    // 为用户获取绑定的代理（保证同一用户总是使用相同IP）
    proxyClient, err := proxy.GetGlobalProxyManager().GetProxyClientForUser(userID)
    if err != nil {
        return fmt.Errorf("获取代理失败: %w", err)
    }

    // 模拟下单请求
    url := fmt.Sprintf("https://fapi.binance.com/fapi/v1/order?symbol=%s", symbol)
    resp, err := proxyClient.Client.Get(url)
    if err != nil {
        // 请求失败，将IP加入黑名单
        // 健康检查器会自动为该用户重新分配健康IP
        proxy.AddBlacklist(proxyClient.IP)
        return fmt.Errorf("请求失败 (IP: %s): %w", proxyClient.IP, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        proxy.AddBlacklist(proxyClient.IP)
        return fmt.Errorf("状态码异常: %d (IP: %s)", resp.StatusCode, proxyClient.IP)
    }

    return nil
}

func printStatus() {
    manager := proxy.GetGlobalProxyManager()

    // 黑名单状态
    total, blacklisted, available := manager.GetBlacklistStatus()
    log.Printf("📊 代理状态: 总计=%d, 黑名单=%d, 可用=%d", total, blacklisted, available)

    // 映射统计
    totalUsers, totalIPs, avg := manager.GetMappingStats()
    log.Printf("👥 用户统计: %d 用户, %d IP, 平均 %.2f 用户/IP", totalUsers, totalIPs, avg)

    // IP分布
    allMappings := manager.GetAllUserIPMappings()
    for ip, users := range allMappings {
        log.Printf("  📍 IP %s: %d 用户", ip, len(users))
    }
}
```

### 示例3：OI数据采集（随机代理）

```go
package main

import (
    "fmt"
    "io"
    "log"
    "nofx/proxy"
    "time"
)

func fetchOIData(symbol string) error {
    // 获取随机代理（不绑定用户）
    proxyClient, err := proxy.GetProxyHTTPClient()
    if err != nil {
        return fmt.Errorf("获取代理失败: %w", err)
    }

    url := fmt.Sprintf("https://fapi.binance.com/futures/data/openInterestHist?symbol=%s&period=5m&limit=1", symbol)
    resp, err := proxyClient.Client.Get(url)
    if err != nil {
        proxy.AddBlacklist(proxyClient.IP)
        return fmt.Errorf("请求失败 (代理: %s): %w", proxyClient.IP, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        proxy.AddBlacklist(proxyClient.IP)
        return fmt.Errorf("状态码异常: %d (代理: %s)", resp.StatusCode, proxyClient.IP)
    }

    body, _ := io.ReadAll(resp.Body)
    log.Printf("✓ 获取 %s OI数据成功 (代理: %s): %s", symbol, proxyClient.IP, string(body))
    return nil
}

func main() {
    // 初始化代理池
    err := proxy.InitGlobalProxyManager(&proxy.Config{
        Enabled: true,
        Mode: "pool",
        ProxyList: []string{
            "http://proxy1.example.com:8080",
            "http://proxy2.example.com:8080",
            "http://proxy3.example.com:8080",
        },
        Timeout: 30 * time.Second,

        HealthCheckURLs: []string{
            "https://fapi.binance.com/fapi/v1/ping",
        },
        HealthCheckInterval: 30 * time.Second,
    })
    if err != nil {
        log.Fatalf("初始化代理失败: %v", err)
    }

    // 循环获取数据
    symbols := []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
    for {
        for _, symbol := range symbols {
            if err := fetchOIData(symbol); err != nil {
                log.Printf("⚠️  %v", err)
            }
            time.Sleep(1 * time.Second)
        }
        time.Sleep(10 * time.Second)
    }
}
```

## 注意事项

### 1. 用户-IP绑定稳定性

- ✅ 同一用户在整个生命周期内使用相同IP，避免触发交易所风控
- ✅ IP失败时自动重新分配，确保用户服务不中断
- ✅ 持久化存储保证重启后绑定关系不变
- ⚠️  扩容IP时，老用户绑定不变，新用户分配到新IP

### 2. 线程安全

- ✅ 所有公开方法都是线程安全的
- ✅ 支持高并发场景下的代理获取和黑名单操作
- ✅ 读写锁优化性能：读操作可并发，写操作独占
- ✅ 32+测试用例覆盖并发场景，通过-race检测

### 3. 错误处理

```go
proxyClient, err := proxy.GetProxyHTTPClient()
if err != nil {
    // 可能的错误：
    // - 代理未启用（返回直连客户端）
    // - 代理IP列表为空
    // - 所有IP都在黑名单中（降级策略）
    // - 代理URL解析失败
    log.Printf("获取代理失败: %v", err)

    // 建议：检查配置或降级处理
    return
}
```

### 4. 性能优化建议

- ✅ 复用 `http.Client`，每次调用返回同一个client实例
- ✅ 合理设置 `health_check_interval`，避免过于频繁的探活
- ✅ 使用文件存储时，`flush_interval` 建议5秒，平衡性能和数据安全
- ✅ 负载均衡算法已优化，O(n)时间复杂度选择最少用户IP

### 5. 安全建议

- 生产环境中代理密钥应使用环境变量或密钥管理服务
- 避免在日志中打印完整的代理URL（包含密码）
- `mapping_file` 文件权限建议设置为 600（仅所有者可读写）
- TLS验证默认开启，不要跳过证书验证

### 6. 调试技巧

```go
manager := proxy.GetGlobalProxyManager()

// 1. 检查代理是否启用
if !proxy.IsEnabled() {
    log.Println("⚠️  代理未启用，请检查配置")
}

// 2. 获取黑名单状态
total, blacklisted, available := manager.GetBlacklistStatus()
log.Printf("📊 代理池状态: 总计=%d, 黑名单=%d, 可用=%d", total, blacklisted, available)

// 3. 查看用户-IP分布
allMappings := manager.GetAllUserIPMappings()
for ip, users := range allMappings {
    log.Printf("📍 IP %s: %d 用户 - %v", ip, len(users), users)
}

// 4. 检查特定用户的IP
proxyClient, err := manager.GetProxyClientForUser("user_123")
if err == nil {
    log.Printf("👤 user_123 绑定的IP: %s", proxyClient.IP)
}

// 5. 查看映射统计
totalUsers, totalIPs, avg := manager.GetMappingStats()
log.Printf("📈 统计: %d 用户, %d IP, 平均 %.2f 用户/IP",
    totalUsers, totalIPs, avg)
```

## 故障排查

### 问题1：获取代理失败 - "代理IP列表为空"

**原因**：
- `single` 模式：未配置 `proxy_url`
- `pool` 模式：`proxy_list` 为空

**解决方案**：
```bash
# 1. 检查配置文件
cat config.json | jq '.proxy'

# 2. 检查日志，查看初始化信息
# 应该看到类似：🌐 HTTP 代理已启用 (pool模式, 3个IP)

# 3. 检查Provider是否正确创建
# 日志应显示: ✓ Provider创建成功: PoolProvider
```

### 问题2：所有IP都不可用

**原因**：所有代理IP健康检查失败或在黑名单中

**解决方案**：
```go
manager := proxy.GetGlobalProxyManager()

// 1. 检查黑名单状态
total, blacklisted, available := manager.GetBlacklistStatus()
log.Printf("总IP: %d, 黑名单: %d, 可用: %d", total, blacklisted, available)

// 2. 如果所有IP都在黑名单，检查代理本身是否可用
// 使用curl测试代理：
// curl -x http://proxy_url https://fapi.binance.com/fapi/v1/ping

// 3. 检查健康检查URL是否正确
// health_check_urls 应该配置为可访问的URL

// 4. 降级策略会自动生效，即使IP不健康也会分配
// 检查日志中是否有降级相关的警告
```

### 问题3：用户IP频繁变化

**原因**：
- 未启用持久化存储，重启后绑定关系丢失
- IP健康检查失败触发自动重分配

**解决方案**：
```json
{
  "proxy": {
    "mapping_file": "data/user_ip_mapping.json",
    "storage_type": "file",
    "flush_interval": 5,

    "health_check_urls": [
      "https://fapi.binance.com/fapi/v1/ping"
    ],
    "health_check_interval": 30,
    "health_check_failure_threshold": 5
  }
}
```

**说明**：
- `mapping_file` 启用持久化，重启后绑定关系保持不变
- 增加 `health_check_failure_threshold` 提高容错性，避免偶发失败触发重分配

### 问题4：代理连接超时

**原因**：代理服务器响应慢或网络不稳定

**解决方案**：
```json
{
  "proxy": {
    "timeout": 60,
    "health_check_timeout": 10
  }
}
```

### 问题5：用户IP分布不均

**原因**：老用户已绑定IP，新用户集中分配到新IP

**解决方案**：
```go
// 查看当前分布
manager := proxy.GetGlobalProxyManager()
allMappings := manager.GetAllUserIPMappings()

for ip, users := range allMappings {
    log.Printf("IP %s: %d 用户", ip, len(users))
}

// 如果需要重新平衡（谨慎操作，会改变用户IP）：
// 1. 停止服务
// 2. 删除 mapping_file
// 3. 重启服务，用户会被重新分配
```

### 问题6：健康检查导致高频请求

**原因**：`health_check_interval` 设置过小

**解决方案**：
```json
{
  "proxy": {
    "health_check_interval": 60,
    "health_check_timeout": 10
  }
}
```

**建议配置**：
- 生产环境：`health_check_interval` 30-60秒
- 开发环境：可设置更短间隔以快速发现问题

## 扩展开发

### 添加新的Provider

实现 `provider.IPProvider` 接口即可：

```go
// provider/custom_provider.go
package provider

type CustomProvider struct {
    endpoint string
    // 其他自定义字段
}

func NewCustomProvider(endpoint string) *CustomProvider {
    return &CustomProvider{
        endpoint: endpoint,
    }
}

func (p *CustomProvider) GetIPList() ([]ProxyIP, error) {
    // 实现获取IP列表的逻辑
    // 例如：从自定义API获取
    return []ProxyIP{
        {IP: "192.168.1.1", Protocol: "http"},
        {IP: "192.168.1.2", Protocol: "http"},
    }, nil
}

func (p *CustomProvider) RefreshIPList() ([]ProxyIP, error) {
    // 实现刷新IP列表的逻辑
    return p.GetIPList()
}
```

然后在 `provider/provider.go` 的 `CreateProvider` 函数中添加新模式：

```go
case "custom":
    provider = NewCustomProvider(customEndpoint)
    log.Printf("✓ Provider创建成功: CustomProvider")
```

### 添加新的Storage后端

实现 `storage.Storage` 接口：

```go
// storage/redis.go
package storage

import (
    "github.com/go-redis/redis/v8"
)

type RedisStorage struct {
    client *redis.Client
    // 其他字段
}

func NewRedisStorage(config RedisConfig) *RedisStorage {
    return &RedisStorage{
        client: redis.NewClient(&redis.Options{
            Addr:     config.Addr,
            Password: config.Password,
            DB:       config.DB,
        }),
    }
}

func (r *RedisStorage) Load() (*MappingData, error) {
    // 从Redis加载数据
    return &MappingData{}, nil
}

func (r *RedisStorage) Save(data *MappingData) error {
    // 保存到Redis
    return nil
}

func (r *RedisStorage) AddUserMapping(userID, ip string) error {
    // 增量添加
    return nil
}

func (r *RedisStorage) RemoveUserMapping(userID, ip string) error {
    // 删除映射
    return nil
}

func (r *RedisStorage) Flush() error {
    // Redis实时写入，无需刷新
    return nil
}

func (r *RedisStorage) Close() error {
    return r.client.Close()
}
```

然后在 `storage/storage.go` 的 `CreateStorage` 函数中添加：

```go
case "redis":
    stor = NewRedisStorage(RedisConfig{
        Addr:     config.RedisAddr,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
    })
```

## 测试

运行所有测试：

```bash
# 运行所有测试
go test ./proxy/... -v

# 运行测试并检查竞态条件
go test ./proxy/... -race

# 运行特定包的测试
go test ./proxy/mapper -v
go test ./proxy/health -v
go test ./proxy/provider -v
go test ./proxy/storage -v

# 查看测试覆盖率
go test ./proxy/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

测试覆盖：
- ✅ proxy: 14个测试（manager, concurrent）
- ✅ proxy/health: 健康检查测试
- ✅ proxy/mapper: 9个测试（映射逻辑、负载均衡、持久化）
- ✅ proxy/provider: 7个测试（Single, Pool提供者）
- ✅ proxy/storage: 11个测试（文件、内存存储、并发）

## 更新日志

### v2.0.0 (当前版本)
- ✅ **重大重构**：分层架构设计（Manager/Health/Mapper/Provider/Storage）
- ✅ **用户-IP绑定**：支持用户与代理IP的稳定绑定
- ✅ **负载均衡**：最少用户优先算法
- ✅ **健康检查**：自动探测IP健康状态，失败自动降级
- ✅ **持久化存储**：支持文件、Redis、数据库（当前实现文件存储）
- ✅ **降级策略**：IP不足时自动降级，确保服务可用
- ✅ **全面测试**：32+测试用例，覆盖核心功能和并发场景
- ✅ **生产就绪**：已修复InitGlobalProxyManager初始化失败、IP降级策略等关键问题

### v1.0.0 (已废弃)
- ✅ 支持三种代理模式：single、pool、brightdata
- ✅ 线程安全的IP轮换和黑名单管理
- ✅ 自动刷新机制
- ✅ TTL黑名单自动恢复

## 技术支持

如有问题或建议，请联系项目维护者@hzb1115。
