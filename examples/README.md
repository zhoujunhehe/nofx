# NOFX Examples - Builder Fee 統計工具

這個目錄包含獨立的 Go 腳本，用於統計 NOFX Builder Fee 使用情況。

## 🎯 設計理念

**✅ 完全基於 Hyperliquid 鏈上 CSV 數據，不依賴本地數據庫**

- 直接解析 Hyperliquid Builder Fills CSV
- 統計真實交易數據和費用
- 避免數據庫同步問題

---

## 📦 工具

### `analyze_builder_stats.go` - Builder Fee 統計分析工具

掃描 Hyperliquid Builder Fills CSV 數據，統計所有使用 NOFX Builder Fee 的用戶。

**功能**：
- ✅ 完全基於鏈上 CSV 數據
- ✅ 不依賴本地數據庫
- ✅ 統計總交易數、總費用、用戶數量
- ✅ 按用戶分組詳細統計交易次數和費用
- ✅ 顯示每個用戶的交易幣種和時間範圍

**使用方法**：

```bash
cd examples

# 默認掃描最近 30 天
go run analyze_builder_stats.go

# 自定義掃描天數（例如 60 天）
go run analyze_builder_stats.go 60

# 掃描最近 7 天
go run analyze_builder_stats.go 7
```

**輸出示例**：

```
🔍 NOFX Builder Fee 統計分析
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Builder 地址: 0x891dc6f05ad47a3c1a05da55e7a7517971faaf0d
掃描範圍: 最近 30 天

📅 2025-11-14: 1 笔交易

📊 總覽統計
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
有數據天數:      1 天
總交易筆數:      1 笔
總 Builder Fee:  0.000148 USDC
授權用戶數量:    1 人

📋 用戶詳情 (按交易次數排序)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
用戶地址                                         | 交易數 | 總費用 (USDC) | 首次交易    | 最後交易    | 交易幣種
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
0x11960d12...3e58dc                          | 1      |      0.000148 | 2025-11-14 | 2025-11-14 | BTC

✅ 分析完成！
```

---

## 📊 CSV 數據說明

**數據來源**：
```
https://stats-data.hyperliquid.xyz/Mainnet/builder_fills/{builder_address}/{YYYYMMDD}.csv.lz4
```

**CSV 字段**：

| 字段 | 說明 |
|------|------|
| `time` | 交易時間 (ISO 8601) |
| `user` | 用戶錢包地址 |
| `coin` | 交易幣種 (BTC, ETH, etc.) |
| `side` | 方向 (Bid/Ask) |
| `px` | 成交價格 |
| `sz` | 成交數量 |
| `builder_fee` | Builder 費用 (USDC) |

**CSV 示例**：
```csv
time,user,coin,side,px,sz,crossed,special_trade_type,tif,is_trigger,counterparty,closed_pnl,twap_id,builder_fee
2025-11-14T04:19:17Z,0x11960d12811a406dd18a4fc3896b9821113e58dc,BTC,Bid,98948,0.00015,true,Na,Ioc,false,0x3ac9b030594c1ef23a3e8fed9f62356b7bf98bf6,0,0,0.000148
```

---

## ⚠️ 重要限制

### 數據範圍
- ✅ **可以統計**：已經產生交易並支付 Builder Fee 的用戶
- ❌ **無法統計**：已授權但未交易的用戶
- ❌ **無法統計**：授權狀態（需要查詢 Hyperliquid API）

### 數據延遲
- ⏰ **約 1-2 天**（建議測試大型 Builder 驗證實際延遲）
- 例如：11/14 交易 → 11/15 或 11/16 CSV 可能才能訪問

### 統計限制
- 只能看到**已交易並收費**的用戶
- 如果用戶授權了但一直沒交易，CSV 中不會有記錄
- 如果 Builder Fee 為 0，也不會出現在 CSV 中

---

## 🔧 安裝依賴

工具需要 `lz4` 解壓 CSV 文件：

```bash
# macOS
brew install lz4

# Ubuntu/Debian
sudo apt-get install lz4

# CentOS/RHEL
sudo yum install lz4
```

---

## 🎯 使用場景

### 1. 查看 Builder 收益統計
```bash
go run analyze_builder_stats.go 90
```
掃描最近 90 天，統計總收益。

### 2. 分析用戶交易行為
查看輸出的用戶詳情表格，了解：
- 哪些用戶交易最頻繁
- 用戶偏好哪些幣種
- 用戶的活躍時間段

### 3. 定期監控
```bash
# 每天運行一次，保存結果
go run analyze_builder_stats.go 30 > daily_stats_$(date +%Y%m%d).txt
```

---

## 🐳 Docker 環境中使用

如果使用 Docker 環境，可以進入容器運行工具：

```bash
# 進入交易容器
docker exec -it nofx-trading sh

# 運行工具
cd /app/examples
go run analyze_builder_stats.go [days]
```

---

## 🔍 故障排查

### 問題 1: 沒有數據

**可能原因**：
1. 掃描時間範圍內沒有交易
2. CSV 數據尚未發布（有延遲，具體天數待驗證）
3. 該 Builder 交易量較少

**解決方案**：
- 擴大掃描天數：`go run analyze_builder_stats.go 90`
- 等待數據發布（建議 1-2 天後重試）
- 測試大型 Builder 驗證數據可用性

### 問題 2: CSV 解壓失敗

**錯誤信息**：
```
⚠️  解壓失敗 20251114: ...
```

**解決方案**：
```bash
# 檢查 lz4 是否安裝
which lz4

# 安裝 lz4
brew install lz4  # macOS
sudo apt-get install lz4  # Ubuntu/Debian
```

### 問題 3: 網絡連接失敗

**錯誤信息**：
```
⚠️  無法下載 CSV
```

**解決方案**：
- 檢查網絡連接
- 確認 Hyperliquid stats-data 服務是否正常
- 測試手動訪問：
  ```bash
  curl -I "https://stats-data.hyperliquid.xyz/Mainnet/builder_fills/0x891dc6f05ad47a3c1a05da55e7a7517971faaf0d/20251114.csv.lz4"
  ```

---

## 📝 NOFX Builder 信息

| 項目 | 值 |
|------|-----|
| **Builder 地址** | `0x891dc6f05ad47a3c1a05da55e7a7517971faaf0d` |
| **Builder Fee 費率** | 0.10% (100 tenths of bp) |
| **推薦碼** | AITRADING |
| **Referral API** | 18 位用戶（截至 2025-11-19）|

---

## 📚 參考資源

- **Hyperliquid Builder Codes**: https://hyperliquid.gitbook.io/hyperliquid-docs/trading/builder-codes
- **Hyperliquid API**: https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api
- **Builder Fills CSV**: https://hyperliquid.gitbook.io/hyperliquid-docs/for-developers/api/builder-fills

---

**創建日期**: 2025-11-19
**維護者**: NOFX Team
**許可證**: AGPL-3.0
