package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	_ "github.com/lib/pq"
)

// AgentWallet 數據結構
type AgentWallet struct {
	ID                      int
	MainWallet              string
	AgentAddress            string
	Status                  string
	HyperliquidChain        string
	BuilderFeeAuthorized    bool
	BuilderFeeMaxRate       int
	BuilderFeeAuthorizedAt  *time.Time
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

func main() {
	// 從環境變量或使用默認值連接數據庫
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nofx:nofx@localhost:5432/nofx?sslmode=disable"
	}

	// 連接數據庫
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ 無法連接數據庫: %v", err)
	}
	defer db.Close()

	// 測試連接
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 數據庫連接失敗: %v", err)
	}

	fmt.Println("✅ 數據庫連接成功")
	fmt.Println()

	// 查詢所有 Agent Wallets
	query := `
		SELECT
			id,
			main_wallet,
			agent_address,
			status,
			hyperliquid_chain,
			builder_fee_authorized,
			builder_fee_max_rate,
			builder_fee_authorized_at,
			created_at,
			updated_at
		FROM agent_wallets
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("❌ 查詢失敗: %v", err)
	}
	defer rows.Close()

	// 統計數據
	var (
		totalCount           int
		activeCount          int
		builderFeeCount      int
		wallets              []AgentWallet
	)

	// 讀取數據
	for rows.Next() {
		var wallet AgentWallet
		err := rows.Scan(
			&wallet.ID,
			&wallet.MainWallet,
			&wallet.AgentAddress,
			&wallet.Status,
			&wallet.HyperliquidChain,
			&wallet.BuilderFeeAuthorized,
			&wallet.BuilderFeeMaxRate,
			&wallet.BuilderFeeAuthorizedAt,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		)
		if err != nil {
			log.Printf("⚠️ 讀取數據失敗: %v", err)
			continue
		}

		wallets = append(wallets, wallet)
		totalCount++

		if wallet.Status == "ACTIVE" {
			activeCount++
		}

		if wallet.BuilderFeeAuthorized {
			builderFeeCount++
		}
	}

	// 顯示統計信息
	fmt.Println("📊 Agent Wallet 統計")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("總數量:              %d\n", totalCount)
	fmt.Printf("已激活 (ACTIVE):     %d\n", activeCount)
	fmt.Printf("Builder Fee 已授權:  %d\n", builderFeeCount)
	fmt.Println()

	if totalCount == 0 {
		fmt.Println("⚠️ 沒有找到任何 Agent Wallet")
		return
	}

	// 顯示詳細列表
	fmt.Println("📋 Agent Wallet 詳細列表")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Fprintln(w, "ID\tMain Wallet\tAgent Address\tStatus\tChain\tBuilder Fee\tFee Rate\tCreated")
	fmt.Fprintln(w, "━━\t━━━━━━━━━━━━\t━━━━━━━━━━━━━\t━━━━━━\t━━━━━\t━━━━━━━━━━━\t━━━━━━━━\t━━━━━━━")

	for _, wallet := range wallets {
		// 截斷地址顯示
		mainWalletShort := truncateAddress(wallet.MainWallet)
		agentAddressShort := truncateAddress(wallet.AgentAddress)

		// Builder Fee 狀態
		builderFeeStatus := "❌ No"
		if wallet.BuilderFeeAuthorized {
			builderFeeStatus = "✅ Yes"
		}

		// 狀態顯示
		statusDisplay := wallet.Status
		if wallet.Status == "ACTIVE" {
			statusDisplay = "✅ " + wallet.Status
		} else {
			statusDisplay = "⚠️ " + wallet.Status
		}

		// Fee Rate 顯示
		feeRateDisplay := "-"
		if wallet.BuilderFeeMaxRate > 0 {
			feeRateDisplay = fmt.Sprintf("%.2f%%", float64(wallet.BuilderFeeMaxRate)/1000)
		}

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			wallet.ID,
			mainWalletShort,
			agentAddressShort,
			statusDisplay,
			wallet.HyperliquidChain,
			builderFeeStatus,
			feeRateDisplay,
			wallet.CreatedAt.Format("2006-01-02 15:04"),
		)
	}

	w.Flush()

	// 顯示詳細信息（如果只有少量記錄）
	if totalCount <= 5 {
		fmt.Println()
		fmt.Println("🔍 詳細信息")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

		for i, wallet := range wallets {
			fmt.Printf("\n[%d] Agent Wallet #%d\n", i+1, wallet.ID)
			fmt.Printf("  Main Wallet:           %s\n", wallet.MainWallet)
			fmt.Printf("  Agent Address:         %s\n", wallet.AgentAddress)
			fmt.Printf("  Status:                %s\n", wallet.Status)
			fmt.Printf("  Hyperliquid Chain:     %s\n", wallet.HyperliquidChain)
			fmt.Printf("  Builder Fee Authorized: %v\n", wallet.BuilderFeeAuthorized)

			if wallet.BuilderFeeAuthorized {
				fmt.Printf("  Builder Fee Rate:      %.2f%% (%d tenths of bp)\n",
					float64(wallet.BuilderFeeMaxRate)/1000,
					wallet.BuilderFeeMaxRate)

				if wallet.BuilderFeeAuthorizedAt != nil {
					fmt.Printf("  Authorized At:         %s\n",
						wallet.BuilderFeeAuthorizedAt.Format("2006-01-02 15:04:05"))
				}
			}

			fmt.Printf("  Created At:            %s\n", wallet.CreatedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Updated At:            %s\n", wallet.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✅ 共顯示 %d 條記錄\n", totalCount)
}

// truncateAddress 截斷地址顯示
func truncateAddress(addr string) string {
	if len(addr) <= 20 {
		return addr
	}
	return addr[:10] + "..." + addr[len(addr)-6:]
}
