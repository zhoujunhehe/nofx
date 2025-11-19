package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	NOFX_BUILDER_ADDRESS = "0x891dc6f05ad47a3c1a05da55e7a7517971faaf0d"
	CSV_BASE_URL         = "https://stats-data.hyperliquid.xyz/Mainnet/builder_fills"
)

// UserStats 用户统计
type UserStats struct {
	Address       string
	TradeCount    int
	TotalFees     float64
	FirstTradeAt  time.Time
	LastTradeAt   time.Time
	CoinsTraded   map[string]int
}

func main() {
	// 参数：扫描天数（默认 30 天）
	days := 30
	if len(os.Args) > 1 {
		if d, err := strconv.Atoi(os.Args[1]); err == nil {
			days = d
		}
	}

	fmt.Printf("🔍 NOFX Builder Fee 统计分析\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Builder 地址: %s\n", NOFX_BUILDER_ADDRESS)
	fmt.Printf("扫描范围: 最近 %d 天\n\n", days)

	// 创建临时目录
	tmpDir := "/tmp/nofx_builder_stats"
	os.MkdirAll(tmpDir, 0755)

	// 收集所有用户数据
	userStats := make(map[string]*UserStats)
	totalDaysWithData := 0
	totalTrades := 0
	totalFees := 0.0

	// 扫描最近 N 天
	now := time.Now()
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("20060102")

		// 尝试下载 CSV
		csvURL := fmt.Sprintf("%s/%s/%s.csv.lz4", CSV_BASE_URL, NOFX_BUILDER_ADDRESS, dateStr)
		lz4File := filepath.Join(tmpDir, fmt.Sprintf("%s.csv.lz4", dateStr))
		csvFile := filepath.Join(tmpDir, fmt.Sprintf("%s.csv", dateStr))

		// 下载
		resp, err := http.Get(csvURL)
		if err != nil || resp.StatusCode != 200 {
			continue // 没有数据，跳过
		}

		// 保存 lz4 文件
		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		os.WriteFile(lz4File, data, 0644)

		// 解压
		cmd := exec.Command("lz4", "-d", lz4File, csvFile, "-f")
		if err := cmd.Run(); err != nil {
			log.Printf("⚠️  解压失败 %s: %v", dateStr, err)
			continue
		}

		// 解析 CSV
		trades := parseCSV(csvFile)
		if len(trades) > 0 {
			totalDaysWithData++
			fmt.Printf("📅 %s: %d 笔交易\n", date.Format("2006-01-02"), len(trades))

			for _, trade := range trades {
				totalTrades++
				totalFees += trade.BuilderFee

				// 更新用户统计
				stats, exists := userStats[trade.User]
				if !exists {
					stats = &UserStats{
						Address:      trade.User,
						TradeCount:   0,
						TotalFees:    0,
						FirstTradeAt: trade.Time,
						LastTradeAt:  trade.Time,
						CoinsTraded:  make(map[string]int),
					}
					userStats[trade.User] = stats
				}

				stats.TradeCount++
				stats.TotalFees += trade.BuilderFee
				stats.CoinsTraded[trade.Coin]++

				if trade.Time.Before(stats.FirstTradeAt) {
					stats.FirstTradeAt = trade.Time
				}
				if trade.Time.After(stats.LastTradeAt) {
					stats.LastTradeAt = trade.Time
				}
			}
		}

		// 清理临时文件
		os.Remove(lz4File)
		os.Remove(csvFile)
	}

	// 打印总览
	fmt.Printf("\n📊 总览统计\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("有数据天数:      %d 天\n", totalDaysWithData)
	fmt.Printf("总交易笔数:      %d 笔\n", totalTrades)
	fmt.Printf("总 Builder Fee:  %.6f USDC\n", totalFees)
	fmt.Printf("授权用户数量:    %d 人\n\n", len(userStats))

	// 按交易次数排序
	users := make([]*UserStats, 0, len(userStats))
	for _, stats := range userStats {
		users = append(users, stats)
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].TradeCount > users[j].TradeCount
	})

	// 打印用户详情
	fmt.Printf("📋 用户详情 (按交易次数排序)\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%-44s | 交易数 | 总费用 (USDC) | 首次交易    | 最后交易    | 交易币种\n", "用户地址")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	for _, user := range users {
		// 截断地址
		addrShort := user.Address[:10] + "..." + user.Address[len(user.Address)-6:]

		// 币种列表
		coins := make([]string, 0, len(user.CoinsTraded))
		for coin := range user.CoinsTraded {
			coins = append(coins, coin)
		}
		coinsStr := strings.Join(coins, ",")
		if len(coinsStr) > 20 {
			coinsStr = coinsStr[:20] + "..."
		}

		fmt.Printf("%-44s | %-6d | %13.6f | %s | %s | %s\n",
			addrShort,
			user.TradeCount,
			user.TotalFees,
			user.FirstTradeAt.Format("2006-01-02"),
			user.LastTradeAt.Format("2006-01-02"),
			coinsStr,
		)
	}

	fmt.Printf("\n✅ 分析完成！\n")
}

// Trade 交易记录
type Trade struct {
	Time       time.Time
	User       string
	Coin       string
	Side       string
	Price      float64
	Size       float64
	BuilderFee float64
}

// parseCSV 解析 CSV 文件
func parseCSV(filename string) []Trade {
	file, err := os.Open(filename)
	if err != nil {
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(bufio.NewReader(file))
	reader.FieldsPerRecord = -1 // 允许字段数量不一致

	// 读取标题行
	header, err := reader.Read()
	if err != nil {
		return nil
	}

	// 找到字段索引
	fieldIndex := make(map[string]int)
	for i, field := range header {
		fieldIndex[field] = i
	}

	var trades []Trade

	// 读取数据行
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		// 解析字段
		trade := Trade{}

		if idx, ok := fieldIndex["time"]; ok && idx < len(record) {
			trade.Time, _ = time.Parse(time.RFC3339, record[idx])
		}
		if idx, ok := fieldIndex["user"]; ok && idx < len(record) {
			trade.User = record[idx]
		}
		if idx, ok := fieldIndex["coin"]; ok && idx < len(record) {
			trade.Coin = record[idx]
		}
		if idx, ok := fieldIndex["side"]; ok && idx < len(record) {
			trade.Side = record[idx]
		}
		if idx, ok := fieldIndex["px"]; ok && idx < len(record) {
			trade.Price, _ = strconv.ParseFloat(record[idx], 64)
		}
		if idx, ok := fieldIndex["sz"]; ok && idx < len(record) {
			trade.Size, _ = strconv.ParseFloat(record[idx], 64)
		}
		if idx, ok := fieldIndex["builder_fee"]; ok && idx < len(record) {
			trade.BuilderFee, _ = strconv.ParseFloat(record[idx], 64)
		}

		trades = append(trades, trade)
	}

	return trades
}
