# NOFX Examples - 數據查看工具

這個目錄包含獨立的 Go 腳本，用於查看 NOFX 數據庫中的數據。

## 📦 工具列表

### 1. `view_agent_wallets.go` - 命令行查看器

查看所有 Agent Wallet 和 Builder Fee 授權狀態的命令行工具。

**功能**：
- ✅ 顯示統計信息（總數、已激活、Builder Fee 授權數量）
- ✅ 列表展示所有 Agent Wallets
- ✅ 詳細信息（當記錄少於 5 條時）
- ✅ 美化的表格輸出

**使用方法**：

```bash
# 方式 1：直接運行（使用默認數據庫配置）
cd examples
go run view_agent_wallets.go

# 方式 2：自定義數據庫連接
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  go run view_agent_wallets.go

# 方式 3：編譯後運行
go build -o view-wallets view_agent_wallets.go
./view-wallets
```

**輸出示例**：

```
✅ 數據庫連接成功

📊 Agent Wallet 統計
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
總數量:              3
已激活 (ACTIVE):     2
Builder Fee 已授權:  1

📋 Agent Wallet 詳細列表
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ID | Main Wallet      | Agent Address    | Status        | Chain   | Builder Fee | Fee Rate | Created
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1  | 0x11960d12...8dc | 0x880debdb...448 | ✅ ACTIVE     | Mainnet | ✅ Yes      | 0.10%    | 2025-11-18 12:01
2  | 0x22334455...abc | 0x99aabbcc...def | ✅ ACTIVE     | Mainnet | ❌ No       | -        | 2025-11-17 10:30
3  | 0x33445566...xyz | 0xaabbccdd...123 | ⚠️ INIT      | Testnet | ❌ No       | -        | 2025-11-16 09:15
```

---

### 2. `web_viewer.go` - Web 界面查看器

帶 Web 界面的實時數據查看器，包含自動刷新功能。

**功能**：
- ✅ 精美的 Web 界面（NOFX 品牌風格）
- ✅ 實時統計卡片（總數、已激活、Builder Fee、待處理）
- ✅ 響應式表格展示
- ✅ 自動刷新（每 30 秒）
- ✅ RESTful API 端點

**使用方法**：

```bash
# 方式 1：直接運行（默認端口 8888）
cd examples
go run web_viewer.go

# 方式 2：自定義端口
PORT=9000 go run web_viewer.go

# 方式 3：自定義數據庫和端口
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  PORT=9000 \
  go run web_viewer.go

# 方式 4：編譯後運行
go build -o web-viewer web_viewer.go
./web-viewer
```

**訪問界面**：

```
🌐 打開瀏覽器訪問: http://localhost:8888
```

**API 端點**：

```bash
# 獲取所有 Agent Wallets
curl http://localhost:8888/api/wallets | jq

# 獲取統計信息
curl http://localhost:8888/api/stats | jq
```

**API 響應示例**：

```json
// GET /api/wallets
[
  {
    "id": 1,
    "main_wallet": "0x11960d12811a406dd18a4fc3896b9821113e58dc",
    "agent_address": "0x880debdb0fd8bfced602cc7408e412875d67e448",
    "status": "ACTIVE",
    "hyperliquid_chain": "Mainnet",
    "builder_fee_authorized": true,
    "builder_fee_max_rate": 100,
    "builder_fee_authorized_at": "2025-11-18T12:15:30Z",
    "created_at": "2025-11-18T12:01:45Z",
    "updated_at": "2025-11-18T12:15:30Z"
  }
]

// GET /api/stats
{
  "total_count": 3,
  "active_count": 2,
  "builder_fee_count": 1,
  "pending_count": 1
}
```

---

## 🔧 環境變量配置

### 數據庫連接

```bash
# PostgreSQL 連接字符串
DATABASE_URL="postgres://username:password@hostname:port/database?sslmode=disable"

# 示例：
# 本地開發
DATABASE_URL="postgres://nofx:nofx@localhost:5432/nofx?sslmode=disable"

# Docker 容器
DATABASE_URL="postgres://nofx:nofx@nofx-postgres:5432/nofx?sslmode=disable"

# 生產環境
DATABASE_URL="postgres://user:pass@db.example.com:5432/nofx?sslmode=require"
```

### Web Viewer 端口

```bash
# 自定義端口（默認 8888）
PORT=9000
```

---

## 📊 數據字段說明

### Agent Wallet 字段

| 字段 | 類型 | 說明 |
|------|------|------|
| `id` | int | 自增 ID |
| `main_wallet` | string | 主錢包地址（用戶的錢包） |
| `agent_address` | string | Agent 錢包地址（後端生成） |
| `status` | string | 狀態：`INIT`（初始）或 `ACTIVE`（已激活） |
| `hyperliquid_chain` | string | Hyperliquid 鏈：`Mainnet` 或 `Testnet` |
| `builder_fee_authorized` | bool | Builder Fee 是否已授權 |
| `builder_fee_max_rate` | int | Builder Fee 費率（tenths of bp，100 = 0.1%） |
| `builder_fee_authorized_at` | timestamp | Builder Fee 授權時間 |
| `created_at` | timestamp | 創建時間 |
| `updated_at` | timestamp | 更新時間 |

### Builder Fee Rate 計算

```
builder_fee_max_rate = 100 (tenths of basis points)
→ 10 basis points
→ 0.1%

計算公式：
percentage = builder_fee_max_rate / 1000
```

**示例**：
- `builder_fee_max_rate = 10` → 0.01%
- `builder_fee_max_rate = 100` → 0.1% ← 推薦值
- `builder_fee_max_rate = 1000` → 1.0%（最大值）

---

## 🐳 Docker 環境中使用

如果數據庫運行在 Docker 容器中：

```bash
# 方式 1：從宿主機連接（使用 localhost）
DATABASE_URL="postgres://nofx:nofx@localhost:5432/nofx?sslmode=disable" \
  go run view_agent_wallets.go

# 方式 2：在 Docker 網絡中運行腳本
# 首先進入 nofx-trading 容器
docker exec -it nofx-trading sh

# 然後在容器內運行
DATABASE_URL="postgres://nofx:nofx@nofx-postgres:5432/nofx?sslmode=disable" \
  go run /path/to/view_agent_wallets.go
```

---

## 🔍 故障排查

### 問題 1：無法連接數據庫

```
❌ 無法連接數據庫: dial tcp [::1]:5432: connect: connection refused
```

**解決方案**：
1. 檢查數據庫是否運行：
   ```bash
   docker ps | grep postgres
   ```

2. 檢查端口映射：
   ```bash
   docker-compose ps
   ```

3. 使用正確的主機名：
   - 宿主機 → Docker：`localhost:5432`
   - Docker → Docker：`nofx-postgres:5432`

### 問題 2：權限錯誤

```
❌ 數據庫連接失敗: pq: password authentication failed
```

**解決方案**：
檢查 `config/config.json` 中的數據庫配置：

```json
{
  "database": {
    "host": "localhost",
    "port": 5432,
    "user": "nofx",
    "password": "nofx",
    "dbname": "nofx"
  }
}
```

### 問題 3：找不到表

```
❌ 查詢失敗: pq: relation "agent_wallets" does not exist
```

**解決方案**：
運行數據庫遷移：

```bash
# 初始化數據庫
./scripts/init-db.sh

# 或者手動創建表
docker exec -it nofx-postgres psql -U nofx -d nofx -f /path/to/schema.sql
```

---

## 📝 開發建議

### 添加新的查看器

如果你想創建新的數據查看工具，可以參考現有的腳本結構：

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    // 1. 連接數據庫
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = "postgres://nofx:nofx@localhost:5432/nofx?sslmode=disable"
    }

    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("❌ 無法連接數據庫: %v", err)
    }
    defer db.Close()

    // 2. 查詢數據
    query := "SELECT ... FROM your_table"
    rows, err := db.Query(query)
    if err != nil {
        log.Fatalf("❌ 查詢失敗: %v", err)
    }
    defer rows.Close()

    // 3. 處理數據
    for rows.Next() {
        // ...
    }

    // 4. 顯示結果
    fmt.Println("結果...")
}
```

---

## 🎯 使用場景

### 場景 1：監控 Builder Fee 授權進度

```bash
# 定期運行查看器
watch -n 10 'go run view_agent_wallets.go'

# 或使用 Web 界面（自動刷新）
go run web_viewer.go
```

### 場景 2：調試授權問題

```bash
# 查看特定用戶的 Agent Wallet 狀態
go run view_agent_wallets.go | grep "0x11960d12"

# 查看所有未授權 Builder Fee 的錢包
curl http://localhost:8888/api/wallets | \
  jq '.[] | select(.builder_fee_authorized == false)'
```

### 場景 3：生成報告

```bash
# 導出為 JSON
curl http://localhost:8888/api/wallets > agent_wallets_$(date +%Y%m%d).json

# 統計信息
curl http://localhost:8888/api/stats | jq
```

---

## 📚 參考資源

- **PostgreSQL 文檔**: https://www.postgresql.org/docs/
- **Go database/sql**: https://pkg.go.dev/database/sql
- **lib/pq (PostgreSQL driver)**: https://github.com/lib/pq

---

**創建日期**: 2025-11-19
**維護者**: NOFX Team
**許可證**: AGPL-3.0
