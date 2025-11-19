package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// AgentWallet 數據結構
type AgentWallet struct {
	ID                      int        `json:"id"`
	MainWallet              string     `json:"main_wallet"`
	AgentAddress            string     `json:"agent_address"`
	Status                  string     `json:"status"`
	HyperliquidChain        string     `json:"hyperliquid_chain"`
	BuilderFeeAuthorized    bool       `json:"builder_fee_authorized"`
	BuilderFeeMaxRate       int        `json:"builder_fee_max_rate"`
	BuilderFeeAuthorizedAt  *time.Time `json:"builder_fee_authorized_at"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

// Stats 統計信息
type Stats struct {
	TotalCount       int `json:"total_count"`
	ActiveCount      int `json:"active_count"`
	BuilderFeeCount  int `json:"builder_fee_count"`
	PendingCount     int `json:"pending_count"`
}

var db *sql.DB

func main() {
	// 從環境變量或使用默認值連接數據庫
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://nofx:nofx@localhost:5432/nofx?sslmode=disable"
	}

	// 連接數據庫
	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("❌ 無法連接數據庫: %v", err)
	}
	defer db.Close()

	// 測試連接
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ 數據庫連接失敗: %v", err)
	}

	fmt.Println("✅ 數據庫連接成功")

	// 設置路由
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/wallets", handleAPIWallets)
	http.HandleFunc("/api/stats", handleAPIStats)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	fmt.Printf("\n🌐 Web 界面已啟動\n")
	fmt.Printf("📍 訪問地址: http://localhost:%s\n", port)
	fmt.Printf("📊 API 端點:\n")
	fmt.Printf("   - GET  /api/wallets  - 獲取所有 Agent Wallets\n")
	fmt.Printf("   - GET  /api/stats    - 獲取統計信息\n")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// handleIndex 首頁
func handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("index").Parse(indexHTML))
	tmpl.Execute(w, nil)
}

// handleAPIWallets API: 獲取所有 wallets
func handleAPIWallets(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wallets []AgentWallet

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
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallets)
}

// handleAPIStats API: 獲取統計信息
func handleAPIStats(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active,
			COUNT(CASE WHEN builder_fee_authorized = true THEN 1 END) as builder_fee,
			COUNT(CASE WHEN status != 'ACTIVE' THEN 1 END) as pending
		FROM agent_wallets
	`

	var stats Stats
	err := db.QueryRow(query).Scan(
		&stats.TotalCount,
		&stats.ActiveCount,
		&stats.BuilderFeeCount,
		&stats.PendingCount,
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// indexHTML 網頁模板
const indexHTML = `
<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NOFX Agent Wallet Viewer</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #0F1419;
            color: #EAECEF;
            padding: 20px;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
        }

        h1 {
            color: #F0B90B;
            margin-bottom: 10px;
            font-size: 32px;
        }

        .subtitle {
            color: #848E9C;
            margin-bottom: 30px;
        }

        .stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }

        .stat-card {
            background: #1C2127;
            padding: 20px;
            border-radius: 12px;
            border: 1px solid #2F3640;
        }

        .stat-label {
            color: #848E9C;
            font-size: 14px;
            margin-bottom: 8px;
        }

        .stat-value {
            font-size: 32px;
            font-weight: bold;
            color: #F0B90B;
        }

        .table-container {
            background: #1C2127;
            border-radius: 12px;
            border: 1px solid #2F3640;
            overflow: hidden;
        }

        table {
            width: 100%;
            border-collapse: collapse;
        }

        thead {
            background: #0F1419;
        }

        th {
            padding: 15px;
            text-align: left;
            font-weight: 600;
            color: #F0B90B;
            border-bottom: 2px solid #2F3640;
        }

        td {
            padding: 15px;
            border-bottom: 1px solid #2F3640;
        }

        tr:last-child td {
            border-bottom: none;
        }

        tbody tr:hover {
            background: rgba(240, 185, 11, 0.05);
        }

        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }

        .status-active {
            background: rgba(34, 197, 94, 0.2);
            color: #22C55E;
        }

        .status-pending {
            background: rgba(251, 191, 36, 0.2);
            color: #FBBF24;
        }

        .builder-fee-yes {
            color: #22C55E;
        }

        .builder-fee-no {
            color: #848E9C;
        }

        .address {
            font-family: 'Courier New', monospace;
            font-size: 13px;
            color: #848E9C;
        }

        .loading {
            text-align: center;
            padding: 40px;
            color: #848E9C;
        }

        .refresh-btn {
            background: #F0B90B;
            color: #000;
            border: none;
            padding: 10px 20px;
            border-radius: 8px;
            font-weight: 600;
            cursor: pointer;
            margin-bottom: 20px;
        }

        .refresh-btn:hover {
            background: #C99A09;
        }

        @media (max-width: 768px) {
            .table-container {
                overflow-x: auto;
            }

            th, td {
                white-space: nowrap;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔐 NOFX Agent Wallet Viewer</h1>
        <p class="subtitle">實時查看 Agent Wallet 和 Builder Fee 授權狀態</p>

        <button class="refresh-btn" onclick="loadData()">🔄 刷新數據</button>

        <div class="stats">
            <div class="stat-card">
                <div class="stat-label">總數量</div>
                <div class="stat-value" id="total-count">-</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">已激活</div>
                <div class="stat-value" id="active-count">-</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">Builder Fee 已授權</div>
                <div class="stat-value" id="builder-fee-count">-</div>
            </div>
            <div class="stat-card">
                <div class="stat-label">待處理</div>
                <div class="stat-value" id="pending-count">-</div>
            </div>
        </div>

        <div class="table-container">
            <table>
                <thead>
                    <tr>
                        <th>ID</th>
                        <th>Main Wallet</th>
                        <th>Agent Address</th>
                        <th>Status</th>
                        <th>Chain</th>
                        <th>Builder Fee</th>
                        <th>Fee Rate</th>
                        <th>Created</th>
                    </tr>
                </thead>
                <tbody id="wallets-tbody">
                    <tr>
                        <td colspan="8" class="loading">載入中...</td>
                    </tr>
                </tbody>
            </table>
        </div>
    </div>

    <script>
        // 載入數據
        async function loadData() {
            try {
                // 載入統計信息
                const statsRes = await fetch('/api/stats');
                const stats = await statsRes.json();

                document.getElementById('total-count').textContent = stats.total_count;
                document.getElementById('active-count').textContent = stats.active_count;
                document.getElementById('builder-fee-count').textContent = stats.builder_fee_count;
                document.getElementById('pending-count').textContent = stats.pending_count;

                // 載入 wallets
                const walletsRes = await fetch('/api/wallets');
                const wallets = await walletsRes.json();

                const tbody = document.getElementById('wallets-tbody');
                tbody.innerHTML = '';

                if (wallets.length === 0) {
                    tbody.innerHTML = '<tr><td colspan="8" class="loading">沒有數據</td></tr>';
                    return;
                }

                wallets.forEach(wallet => {
                    const row = document.createElement('tr');

                    // 截斷地址
                    const mainWalletShort = truncateAddress(wallet.main_wallet);
                    const agentAddressShort = truncateAddress(wallet.agent_address);

                    // Status badge
                    const statusClass = wallet.status === 'ACTIVE' ? 'status-active' : 'status-pending';
                    const statusBadge = \`<span class="status-badge \${statusClass}">\${wallet.status}</span>\`;

                    // Builder Fee
                    const builderFee = wallet.builder_fee_authorized
                        ? '<span class="builder-fee-yes">✅ Yes</span>'
                        : '<span class="builder-fee-no">❌ No</span>';

                    // Fee Rate
                    const feeRate = wallet.builder_fee_max_rate > 0
                        ? (wallet.builder_fee_max_rate / 1000).toFixed(2) + '%'
                        : '-';

                    // Created
                    const created = new Date(wallet.created_at).toLocaleString('zh-TW', {
                        year: 'numeric',
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit'
                    });

                    row.innerHTML = \`
                        <td>\${wallet.id}</td>
                        <td class="address" title="\${wallet.main_wallet}">\${mainWalletShort}</td>
                        <td class="address" title="\${wallet.agent_address}">\${agentAddressShort}</td>
                        <td>\${statusBadge}</td>
                        <td>\${wallet.hyperliquid_chain}</td>
                        <td>\${builderFee}</td>
                        <td>\${feeRate}</td>
                        <td>\${created}</td>
                    \`;

                    tbody.appendChild(row);
                });

            } catch (error) {
                console.error('載入失敗:', error);
                document.getElementById('wallets-tbody').innerHTML =
                    '<tr><td colspan="8" class="loading">載入失敗: ' + error.message + '</td></tr>';
            }
        }

        // 截斷地址
        function truncateAddress(addr) {
            if (addr.length <= 20) return addr;
            return addr.substring(0, 10) + '...' + addr.substring(addr.length - 6);
        }

        // 頁面載入時自動載入數據
        loadData();

        // 每 30 秒自動刷新
        setInterval(loadData, 30000);
    </script>
</body>
</html>
`
