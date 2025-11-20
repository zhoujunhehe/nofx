package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"nofx/crypto"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgreSQLDatabase PostgreSQL数据库配置
type PostgreSQLDatabase struct {
	db            *sqlx.DB
	cryptoService *crypto.CryptoService
}

// NewPostgreSQLDatabase 创建PostgreSQL数据库连接
func NewPostgreSQLDatabase() (*PostgreSQLDatabase, error) {
	// 从环境变量获取数据库连接信息
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	dbname := getEnv("POSTGRES_DB", "nofx")
	user := getEnv("POSTGRES_USER", "nofx")
	password := getEnv("POSTGRES_PASSWORD", "nofx123456")

	// 构建连接字符串
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	log.Printf("📋 连接PostgreSQL数据库: %s:%s/%s", host, port, dbname)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开PostgreSQL数据库失败: %w", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("连接PostgreSQL数据库失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	database := &PostgreSQLDatabase{db: db}
	log.Printf("✅ PostgreSQL数据库连接成功")

	// 初始化默认数据
	if err := database.initDefaultData(); err != nil {
		return nil, fmt.Errorf("初始化默认数据失败: %w", err)
	}

	return database, nil
}

func (d *PostgreSQLDatabase) SetCryptoService(cs *crypto.CryptoService) {
	d.cryptoService = cs
}

func (d *PostgreSQLDatabase) encryptValue(value string, aadParts ...string) (string, error) {
	if value == "" {
		return "", nil
	}
	if d.cryptoService == nil {
		return "", fmt.Errorf("crypto service not initialized")
	}
	if !d.cryptoService.HasDataKey() {
		return "", fmt.Errorf("data encryption key not configured")
	}
	if d.cryptoService.IsEncryptedStorageValue(value) {
		return value, nil
	}
	return d.cryptoService.EncryptForStorage(value, aadParts...)
}

func (d *PostgreSQLDatabase) decryptValue(value string, aadParts ...string) (string, error) {
	if value == "" {
		return "", nil
	}
	if d.cryptoService == nil {
		return "", fmt.Errorf("crypto service not initialized")
	}
	if !d.cryptoService.HasDataKey() {
		return "", fmt.Errorf("data encryption key not configured")
	}
	if !d.cryptoService.IsEncryptedStorageValue(value) {
		return "", fmt.Errorf("value is not encrypted")
	}
	return d.cryptoService.DecryptFromStorage(value, aadParts...)
}

// getEnv 获取环境变量，如果不存在返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// CreateUser 创建用户
func (d *PostgreSQLDatabase) CreateUser(user *User) error {
	_, err := d.db.Exec(`
		INSERT INTO users (id, email, password_hash, otp_secret, otp_verified)
		VALUES ($1, $2, $3, $4, $5)
	`, user.ID, user.Email, user.PasswordHash, user.OTPSecret, user.OTPVerified)
	return err
}

// GetUserByEmail 通过邮箱获取用户
func (d *PostgreSQLDatabase) GetUserByEmail(email string) (*User, error) {
	var user User
	err := d.db.QueryRow(`
		SELECT id, email, password_hash, otp_secret, otp_verified, created_at, updated_at
		FROM users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.OTPSecret,
		&user.OTPVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 通过ID获取用户
func (d *PostgreSQLDatabase) GetUserByID(userID string) (*User, error) {
	var user User
	err := d.db.QueryRow(`
		SELECT id, email, password_hash, otp_secret, otp_verified, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.OTPSecret,
		&user.OTPVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers 获取所有用户ID列表
func (d *PostgreSQLDatabase) GetAllUsers() ([]string, error) {
	rows, err := d.db.Query(`SELECT id FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}

// UpdateUserOTPVerified 更新用户OTP验证状态
func (d *PostgreSQLDatabase) UpdateUserOTPVerified(userID string, verified bool) error {
	_, err := d.db.Exec(`UPDATE users SET otp_verified = $1 WHERE id = $2`, verified, userID)
	return err
}

// UpdateUserPassword 更新用户密码
func (d *PostgreSQLDatabase) UpdateUserPassword(userID, newPasswordHash string) error {
	_, err := d.db.Exec(`UPDATE users SET password_hash = $1 WHERE id = $2`, newPasswordHash, userID)
	return err
}

// GetAIModels 获取用户的AI模型配置
func (d *PostgreSQLDatabase) GetAIModels(userID string) ([]*AIModelConfig, error) {
	rows, err := d.db.Query(`
		SELECT id, user_id, name, provider, enabled, api_key,
		       COALESCE(custom_api_url, '') as custom_api_url,
		       COALESCE(custom_model_name, '') as custom_model_name,
		       COALESCE(deleted, FALSE) as deleted,
		       created_at, updated_at
		FROM ai_models WHERE user_id = $1 AND COALESCE(deleted, FALSE) = FALSE ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为空切片而不是nil，确保JSON序列化为[]而不是null
	models := make([]*AIModelConfig, 0)
	for rows.Next() {
		var model AIModelConfig
		var deleted bool // 临时变量，用于读取 deleted 字段但不保存到结构体
		err := rows.Scan(
			&model.ID, &model.UserID, &model.Name, &model.Provider,
			&model.Enabled, &model.APIKey, &model.CustomAPIURL, &model.CustomModelName,
			&deleted, &model.CreatedAt, &model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if model.APIKey != "" {
			decrypted, err := d.decryptValue(model.APIKey, model.UserID, model.ID, "api_key")
			if err != nil {
				return nil, err
			}
			model.APIKey = decrypted
		}

		models = append(models, &model)
	}

	return models, nil
}

// UpdateAIModel 更新AI模型配置，如果不存在则创建用户特定配置
func (d *PostgreSQLDatabase) UpdateAIModel(userID, id string, enabled bool, apiKey, customAPIURL, customModelName string) error {
	log.Printf("🔧 UpdateAIModel: userID=%s, id=%s, enabled=%v", userID, id, enabled)

	// 检查是否为删除操作（API Key 为空且 enabled 为 false 表示删除）
	isDelete := !enabled && apiKey == "" && customAPIURL == "" && customModelName == ""

	if isDelete {
		// 执行软删除：标记为已删除并清空敏感数据
		// 先尝试精确匹配 ID
		var existingID string
		err := d.db.QueryRow(`
			SELECT id FROM ai_models WHERE user_id = $1 AND id = $2 LIMIT 1
		`, userID, id).Scan(&existingID)

		if err == nil {
			// 找到了现有配置（精确匹配 ID），标记为删除并清空敏感数据
			_, err = d.db.Exec(`
				UPDATE ai_models SET enabled = FALSE, deleted = TRUE, api_key = '', custom_api_url = '', custom_model_name = '', updated_at = CURRENT_TIMESTAMP
				WHERE id = $1 AND user_id = $2
			`, existingID, userID)
			if err != nil {
				log.Printf("❌ UpdateAIModel: 标记删除失败: %v", err)
				return err
			}
			log.Printf("🗑️ UpdateAIModel: 已标记删除用户 %s 的模型配置 %s", userID, existingID)
			return nil
		}

		// ID 不存在，尝试兼容旧逻辑：将 id 作为 provider 查找
		provider := id
		err = d.db.QueryRow(`
			SELECT id FROM ai_models WHERE user_id = $1 AND provider = $2 LIMIT 1
		`, userID, provider).Scan(&existingID)

		if err == nil {
			// 找到了现有配置（通过 provider 匹配），标记为删除并清空敏感数据
			_, err = d.db.Exec(`
				UPDATE ai_models SET enabled = FALSE, deleted = TRUE, api_key = '', custom_api_url = '', custom_model_name = '', updated_at = CURRENT_TIMESTAMP
				WHERE id = $1 AND user_id = $2
			`, existingID, userID)
			if err != nil {
				log.Printf("❌ UpdateAIModel: 标记删除失败: %v", err)
				return err
			}
			log.Printf("🗑️ UpdateAIModel: 已标记删除用户 %s 的模型配置 %s (通过provider匹配)", userID, existingID)
			return nil
		}

		// 没有找到配置，返回成功（幂等性）
		log.Printf("ℹ️ UpdateAIModel: 模型配置不存在，跳过删除: %s", id)
		return nil
	}

	// 启用模型的情况：先尝试精确匹配 ID（新版逻辑，支持多个相同 provider 的模型）
	var existingID string
	err := d.db.QueryRow(`
		SELECT id FROM ai_models WHERE user_id = $1 AND id = $2 LIMIT 1
	`, userID, id).Scan(&existingID)

	if err == nil {
		apiKeyEnc, err := d.encryptValue(apiKey, userID, existingID, "api_key")
		if err != nil {
			return err
		}
		// 找到了现有配置（精确匹配 ID），更新它
		_, err = d.db.Exec(`
			UPDATE ai_models SET enabled = $1, api_key = $2, custom_api_url = $3, custom_model_name = $4, deleted = FALSE, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5 AND user_id = $6
		`, enabled, apiKeyEnc, customAPIURL, customModelName, existingID, userID)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}

	// ID 不存在，尝试兼容旧逻辑：将 id 作为 provider 查找
	provider := id
	err = d.db.QueryRow(`
		SELECT id FROM ai_models WHERE user_id = $1 AND provider = $2 LIMIT 1
	`, userID, provider).Scan(&existingID)

	if err == nil {
		apiKeyEnc, err := d.encryptValue(apiKey, userID, existingID, "api_key")
		if err != nil {
			return err
		}
		// 找到了现有配置（通过 provider 匹配，兼容旧版），更新它
		log.Printf("⚠️  使用旧版 provider 匹配更新模型: %s -> %s", provider, existingID)
		_, err = d.db.Exec(`
			UPDATE ai_models SET enabled = $1, api_key = $2, custom_api_url = $3, custom_model_name = $4, deleted = FALSE, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5 AND user_id = $6
		`, enabled, apiKeyEnc, customAPIURL, customModelName, existingID, userID)
		return err
	}
	if err != sql.ErrNoRows {
		return err
	}

	// 没有找到任何现有配置，创建新的
	// 推断 provider（从 id 中提取，或者直接使用 id）
	if provider == id && (provider == "deepseek" || provider == "qwen") {
		// id 本身就是 provider
		provider = id
	} else {
		// 从 id 中提取 provider（假设格式是 userID_provider 或 timestamp_userID_provider）
		parts := strings.Split(id, "_")
		if len(parts) >= 2 {
			provider = parts[len(parts)-1] // 取最后一部分作为 provider
		} else {
			provider = id
		}
	}

	// 获取模型的基本信息
	var name string
	err = d.db.QueryRow(`
		SELECT name FROM ai_models WHERE provider = $1 LIMIT 1
	`, provider).Scan(&name)
	if err != nil {
		// 如果找不到基本信息，使用默认值
		if provider == "deepseek" {
			name = "DeepSeek AI"
		} else if provider == "qwen" {
			name = "Qwen AI"
		} else {
			name = provider + " AI"
		}
	}

	// 如果传入的 ID 已经是完整格式（如 "admin_deepseek_custom1"），直接使用
	// 否则生成新的 ID
	newModelID := id
	if id == provider {
		// id 就是 provider，生成新的用户特定 ID
		newModelID = fmt.Sprintf("%s_%s", userID, provider)
	}

	apiKeyEnc, err := d.encryptValue(apiKey, userID, newModelID, "api_key")
	if err != nil {
		return err
	}

	log.Printf("✓ 创建新的 AI 模型配置: ID=%s, Provider=%s, Name=%s", newModelID, provider, name)
	_, err = d.db.Exec(`
		INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_api_url, custom_model_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, newModelID, userID, name, provider, enabled, apiKeyEnc, customAPIURL, customModelName)

	return err
}

// GetExchanges 获取用户的交易所配置
func (d *PostgreSQLDatabase) GetExchanges(userID string) ([]*ExchangeConfig, error) {
	rows, err := d.db.Query(`
		SELECT id, user_id, name, type, enabled, api_key, secret_key, testnet,
		       COALESCE(hyperliquid_wallet_addr, '') AS hyperliquid_wallet_addr,
		       COALESCE(aster_user, '') AS aster_user,
		       COALESCE(aster_signer, '') AS aster_signer,
		       COALESCE(aster_private_key, '') AS aster_private_key,
		       COALESCE(deleted, FALSE) AS deleted,
		       created_at, updated_at
		FROM exchanges
		WHERE user_id = $1 AND COALESCE(deleted, FALSE) = FALSE
		ORDER BY id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 初始化为空切片而不是nil，确保JSON序列化为[]而不是null
	exchanges := make([]*ExchangeConfig, 0)
	for rows.Next() {
		var exchange ExchangeConfig
		err := rows.Scan(
			&exchange.ID, &exchange.UserID, &exchange.Name, &exchange.Type,
			&exchange.Enabled, &exchange.APIKey, &exchange.SecretKey, &exchange.Testnet,
			&exchange.HyperliquidWalletAddr, &exchange.AsterUser,
			&exchange.AsterSigner, &exchange.AsterPrivateKey,
			&exchange.Deleted,
			&exchange.CreatedAt, &exchange.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if decrypted, err := d.decryptValue(exchange.APIKey, exchange.UserID, exchange.ID, "api_key"); err == nil {
			exchange.APIKey = decrypted
		} else {
			return nil, err
		}
		if decrypted, err := d.decryptValue(exchange.SecretKey, exchange.UserID, exchange.ID, "secret_key"); err == nil {
			exchange.SecretKey = decrypted
		} else {
			return nil, err
		}
		if decrypted, err := d.decryptValue(exchange.HyperliquidWalletAddr, exchange.UserID, exchange.ID, "hyperliquid_wallet_addr"); err == nil {
			exchange.HyperliquidWalletAddr = decrypted
		} else {
			return nil, err
		}
		if decrypted, err := d.decryptValue(exchange.AsterUser, exchange.UserID, exchange.ID, "aster_user"); err == nil {
			exchange.AsterUser = decrypted
		} else {
			return nil, err
		}
		if decrypted, err := d.decryptValue(exchange.AsterSigner, exchange.UserID, exchange.ID, "aster_signer"); err == nil {
			exchange.AsterSigner = decrypted
		} else {
			return nil, err
		}
		if decrypted, err := d.decryptValue(exchange.AsterPrivateKey, exchange.UserID, exchange.ID, "aster_private_key"); err == nil {
			exchange.AsterPrivateKey = decrypted
		} else {
			return nil, err
		}

		exchanges = append(exchanges, &exchange)
	}

	return exchanges, nil
}

// UpdateExchange 更新交易所配置，如果不存在则创建用户特定配置
func (d *PostgreSQLDatabase) UpdateExchange(userID, id string, enabled bool, apiKey, secretKey string, testnet bool, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey string) error {
	log.Printf("🔧 UpdateExchange: userID=%s, id=%s, enabled=%v", userID, id, enabled)

	// 如果请求禁用该交易所，执行软删除
	if !enabled {
		_, err := d.db.Exec(`
			UPDATE exchanges
			SET enabled = FALSE,
			    deleted = TRUE,
			    api_key = '',
			    secret_key = '',
			    testnet = FALSE,
			    hyperliquid_wallet_addr = '',
			    aster_user = '',
			    aster_signer = '',
			    aster_private_key = '',
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND user_id = $2
		`, id, userID)
		if err != nil {
			log.Printf("❌ UpdateExchange: 标记删除失败: %v", err)
			return err
		}
		log.Printf("🗑️ UpdateExchange: 已标记删除用户 %s 的交易所配置 %s", userID, id)
		return nil
	}

	apiKeyEnc, err := d.encryptValue(apiKey, userID, id, "api_key")
	if err != nil {
		return fmt.Errorf("encrypt api_key failed: %w", err)
	}
	secretKeyEnc, err := d.encryptValue(secretKey, userID, id, "secret_key")
	if err != nil {
		return fmt.Errorf("encrypt secret_key failed: %w", err)
	}
	hyperAddrEnc, err := d.encryptValue(hyperliquidWalletAddr, userID, id, "hyperliquid_wallet_addr")
	if err != nil {
		return fmt.Errorf("encrypt hyperliquid_wallet_addr failed: %w", err)
	}
	asterUserEnc, err := d.encryptValue(asterUser, userID, id, "aster_user")
	if err != nil {
		return fmt.Errorf("encrypt aster_user failed: %w", err)
	}
	asterSignerEnc, err := d.encryptValue(asterSigner, userID, id, "aster_signer")
	if err != nil {
		return fmt.Errorf("encrypt aster_signer failed: %w", err)
	}
	asterPrivateKeyEnc, err := d.encryptValue(asterPrivateKey, userID, id, "aster_private_key")
	if err != nil {
		return fmt.Errorf("encrypt aster_private_key failed: %w", err)
	}

	// 首先尝试更新现有的用户配置
	result, err := d.db.Exec(`
		UPDATE exchanges SET enabled = $1, api_key = $2, secret_key = $3, testnet = $4,
		       hyperliquid_wallet_addr = $5, aster_user = $6, aster_signer = $7, aster_private_key = $8,
		       deleted = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9 AND user_id = $10
	`, enabled, apiKeyEnc, secretKeyEnc, testnet, hyperAddrEnc, asterUserEnc, asterSignerEnc, asterPrivateKeyEnc, id, userID)
	if err != nil {
		log.Printf("❌ UpdateExchange: 更新失败: %v", err)
		return err
	}

	// 检查是否有行被更新
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("❌ UpdateExchange: 获取影响行数失败: %v", err)
		return err
	}

	log.Printf("📊 UpdateExchange: 影响行数 = %d", rowsAffected)

	// 如果没有行被更新，说明用户没有这个交易所的配置，需要创建
	if rowsAffected == 0 {
		log.Printf("💡 UpdateExchange: 没有现有记录，创建新记录")

		// 根据交易所ID确定基本信息
		var name, typ string
		if id == "binance" {
			name = "Binance Futures"
			typ = "cex"
		} else if id == "hyperliquid" {
			name = "Hyperliquid"
			typ = "dex"
		} else if id == "aster" {
			name = "Aster DEX"
			typ = "dex"
		} else {
			name = id + " Exchange"
			typ = "cex"
		}

		log.Printf("🆕 UpdateExchange: 创建新记录 ID=%s, name=%s, type=%s", id, name, typ)

		// 创建用户特定的配置，使用原始的交易所ID
		_, err = d.db.Exec(`
			INSERT INTO exchanges (id, user_id, name, type, enabled, api_key, secret_key, testnet,
			                       hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key,
			                       deleted, created_at, updated_at)
			VALUES ($1, $2, $3, $4, TRUE, $5, $6, $7, $8, $9, $10, $11, FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		`, id, userID, name, typ, apiKeyEnc, secretKeyEnc, testnet, hyperAddrEnc, asterUserEnc, asterSignerEnc, asterPrivateKeyEnc)

		if err != nil {
			log.Printf("❌ UpdateExchange: 创建记录失败: %v", err)
		} else {
			log.Printf("✅ UpdateExchange: 创建记录成功")
		}
		return err
	}

	log.Printf("✅ UpdateExchange: 更新现有记录成功")
	return nil
}

// CreateAIModel 创建AI模型配置
func (d *PostgreSQLDatabase) CreateAIModel(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error {
	apiKeyEnc, err := d.encryptValue(apiKey, userID, id, "api_key")
	if err != nil {
		return err
	}

	_, err = d.db.Exec(`
		INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_api_url) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`, id, userID, name, provider, enabled, apiKeyEnc, customAPIURL)
	return err
}

// CreateExchange 创建交易所配置
func (d *PostgreSQLDatabase) CreateExchange(userID, id, name, typ string, enabled bool, apiKey, secretKey string, testnet bool, hyperliquidWalletAddr, asterUser, asterSigner, asterPrivateKey string) error {
	apiKeyEnc, err := d.encryptValue(apiKey, userID, id, "api_key")
	if err != nil {
		return fmt.Errorf("encrypt api_key failed: %w", err)
	}
	secretKeyEnc, err := d.encryptValue(secretKey, userID, id, "secret_key")
	if err != nil {
		return fmt.Errorf("encrypt secret_key failed: %w", err)
	}
	hyperAddrEnc, err := d.encryptValue(hyperliquidWalletAddr, userID, id, "hyperliquid_wallet_addr")
	if err != nil {
		return fmt.Errorf("encrypt hyperliquid_wallet_addr failed: %w", err)
	}
	asterUserEnc, err := d.encryptValue(asterUser, userID, id, "aster_user")
	if err != nil {
		return fmt.Errorf("encrypt aster_user failed: %w", err)
	}
	asterSignerEnc, err := d.encryptValue(asterSigner, userID, id, "aster_signer")
	if err != nil {
		return fmt.Errorf("encrypt aster_signer failed: %w", err)
	}
	asterPrivateKeyEnc, err := d.encryptValue(asterPrivateKey, userID, id, "aster_private_key")
	if err != nil {
		return fmt.Errorf("encrypt aster_private_key failed: %w", err)
	}

	_, err = d.db.Exec(`
		INSERT INTO exchanges (id, user_id, name, type, enabled, api_key, secret_key, testnet, hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id, user_id) DO NOTHING
	`, id, userID, name, typ, enabled, apiKeyEnc, secretKeyEnc, testnet, hyperAddrEnc, asterUserEnc, asterSignerEnc, asterPrivateKeyEnc)
	return err
}

// CreateTrader 创建交易员
func (d *PostgreSQLDatabase) CreateTrader(trader *TraderRecord) error {
	_, err := d.db.Exec(`
		INSERT INTO traders (id, user_id, name, ai_model_id, exchange_id, initial_balance, scan_interval_minutes, is_running, btc_eth_leverage, altcoin_leverage, trading_symbols, use_coin_pool, use_oi_top, custom_prompt, override_base_prompt, system_prompt_template, is_cross_margin, kline_intervals)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, trader.ID, trader.UserID, trader.Name, trader.AIModelID, trader.ExchangeID, trader.InitialBalance, trader.ScanIntervalMinutes, trader.IsRunning, trader.BTCETHLeverage, trader.AltcoinLeverage, trader.TradingSymbols, trader.UseCoinPool, trader.UseOITop, trader.CustomPrompt, trader.OverrideBasePrompt, trader.SystemPromptTemplate, trader.IsCrossMargin, trader.KlineIntervals)
	return err
}

// GetTraders 获取用户的交易员
func (d *PostgreSQLDatabase) GetTraders(userID string) ([]*TraderRecord, error) {
	rows, err := d.db.Query(`
		SELECT id, user_id, name, ai_model_id, exchange_id, initial_balance, scan_interval_minutes, is_running,
		       COALESCE(btc_eth_leverage, 5) as btc_eth_leverage, COALESCE(altcoin_leverage, 5) as altcoin_leverage,
		       COALESCE(trading_symbols, '') as trading_symbols,
		       COALESCE(use_coin_pool, false) as use_coin_pool, COALESCE(use_oi_top, false) as use_oi_top,
		       COALESCE(custom_prompt, '') as custom_prompt, COALESCE(override_base_prompt, false) as override_base_prompt,
		       COALESCE(system_prompt_template, 'default') as system_prompt_template,
		       COALESCE(is_cross_margin, true) as is_cross_margin, 
		       COALESCE(kline_intervals, '') as kline_intervals, created_at, updated_at
		FROM traders WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traders []*TraderRecord
	for rows.Next() {
		var trader TraderRecord
		err := rows.Scan(
			&trader.ID, &trader.UserID, &trader.Name, &trader.AIModelID, &trader.ExchangeID,
			&trader.InitialBalance, &trader.ScanIntervalMinutes, &trader.IsRunning,
			&trader.BTCETHLeverage, &trader.AltcoinLeverage, &trader.TradingSymbols,
			&trader.UseCoinPool, &trader.UseOITop,
			&trader.CustomPrompt, &trader.OverrideBasePrompt, &trader.SystemPromptTemplate,
			&trader.IsCrossMargin, &trader.KlineIntervals,
			&trader.CreatedAt, &trader.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		traders = append(traders, &trader)
	}

	return traders, nil
}

// UpdateTraderStatus 更新交易员状态
func (d *PostgreSQLDatabase) UpdateTraderStatus(userID, id string, isRunning bool) error {
	_, err := d.db.Exec(`UPDATE traders SET is_running = $1 WHERE id = $2 AND user_id = $3`, isRunning, id, userID)
	return err
}

// UpdateTrader 更新交易员配置
func (d *PostgreSQLDatabase) UpdateTrader(trader *TraderRecord) error {
	_, err := d.db.Exec(`
		UPDATE traders SET
			name = $1, ai_model_id = $2, exchange_id = $3, initial_balance = $4,
			scan_interval_minutes = $5, btc_eth_leverage = $6, altcoin_leverage = $7,
			trading_symbols = $8, custom_prompt = $9, override_base_prompt = $10,
			system_prompt_template = $11, is_cross_margin = $12, kline_intervals = $13, updated_at = CURRENT_TIMESTAMP
		WHERE id = $14 AND user_id = $15
	`, trader.Name, trader.AIModelID, trader.ExchangeID, trader.InitialBalance,
		trader.ScanIntervalMinutes, trader.BTCETHLeverage, trader.AltcoinLeverage,
		trader.TradingSymbols, trader.CustomPrompt, trader.OverrideBasePrompt,
		trader.SystemPromptTemplate, trader.IsCrossMargin, trader.KlineIntervals, trader.ID, trader.UserID)
	return err
}

// UpdateTraderCustomPrompt 更新交易员自定义Prompt
func (d *PostgreSQLDatabase) UpdateTraderCustomPrompt(userID, id string, customPrompt string, overrideBase bool) error {
	_, err := d.db.Exec(`UPDATE traders SET custom_prompt = $1, override_base_prompt = $2 WHERE id = $3 AND user_id = $4`, customPrompt, overrideBase, id, userID)
	return err
}

// UpdateTraderInitialBalance 更新交易员初始余额（用于自动同步交易所实际余额）
func (d *PostgreSQLDatabase) UpdateTraderInitialBalance(userID, id string, newBalance float64) error {
	_, err := d.db.Exec(`UPDATE traders SET initial_balance = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3`, newBalance, id, userID)
	return err
}

// DeleteTrader 删除交易员
func (d *PostgreSQLDatabase) DeleteTrader(userID, id string) error {
	_, err := d.db.Exec(`DELETE FROM traders WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// GetTraderConfig 获取交易员完整配置（包含AI模型和交易所信息）
func (d *PostgreSQLDatabase) GetTraderConfig(userID, traderID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error) {
	var trader TraderRecord
	var aiModel AIModelConfig
	var exchange ExchangeConfig

	err := d.db.QueryRow(`
		SELECT 
			t.id, t.user_id, t.name, t.ai_model_id, t.exchange_id, t.initial_balance, t.scan_interval_minutes, t.is_running,
			COALESCE(t.btc_eth_leverage, 5) as btc_eth_leverage, COALESCE(t.altcoin_leverage, 5) as altcoin_leverage,
			COALESCE(t.trading_symbols, '') as trading_symbols,
			COALESCE(t.use_coin_pool, false) as use_coin_pool, COALESCE(t.use_oi_top, false) as use_oi_top,
			COALESCE(t.custom_prompt, '') as custom_prompt, COALESCE(t.override_base_prompt, false) as override_base_prompt,
			COALESCE(t.system_prompt_template, 'default') as system_prompt_template,
			COALESCE(t.is_cross_margin, true) as is_cross_margin,
			COALESCE(t.kline_intervals, '') as kline_intervals,
			t.created_at, t.updated_at,
			a.id, a.user_id, a.name, a.provider, a.enabled, a.api_key, a.created_at, a.updated_at,
			e.id, e.user_id, e.name, e.type, e.enabled, e.api_key, e.secret_key, e.testnet,
			COALESCE(e.hyperliquid_wallet_addr, '') as hyperliquid_wallet_addr,
			COALESCE(e.aster_user, '') as aster_user,
			COALESCE(e.aster_signer, '') as aster_signer,
			COALESCE(e.aster_private_key, '') as aster_private_key,
			e.created_at, e.updated_at
		FROM traders t
		JOIN ai_models a ON t.ai_model_id = a.id AND t.user_id = a.user_id
		JOIN exchanges e ON t.exchange_id = e.id AND t.user_id = e.user_id
		WHERE t.id = $1 AND t.user_id = $2
	`, traderID, userID).Scan(
		&trader.ID, &trader.UserID, &trader.Name, &trader.AIModelID, &trader.ExchangeID,
		&trader.InitialBalance, &trader.ScanIntervalMinutes, &trader.IsRunning,
		&trader.BTCETHLeverage, &trader.AltcoinLeverage, &trader.TradingSymbols,
		&trader.UseCoinPool, &trader.UseOITop,
		&trader.CustomPrompt, &trader.OverrideBasePrompt, &trader.SystemPromptTemplate,
		&trader.IsCrossMargin, &trader.KlineIntervals,
		&trader.CreatedAt, &trader.UpdatedAt,
		&aiModel.ID, &aiModel.UserID, &aiModel.Name, &aiModel.Provider, &aiModel.Enabled, &aiModel.APIKey,
		&aiModel.CreatedAt, &aiModel.UpdatedAt,
		&exchange.ID, &exchange.UserID, &exchange.Name, &exchange.Type, &exchange.Enabled,
		&exchange.APIKey, &exchange.SecretKey, &exchange.Testnet,
		&exchange.HyperliquidWalletAddr, &exchange.AsterUser, &exchange.AsterSigner, &exchange.AsterPrivateKey,
		&exchange.CreatedAt, &exchange.UpdatedAt,
	)

	if err != nil {
		return nil, nil, nil, err
	}

	if aiModel.APIKey != "" {
		decrypted, err := d.decryptValue(aiModel.APIKey, aiModel.UserID, aiModel.ID, "api_key")
		if err != nil {
			return nil, nil, nil, err
		}
		aiModel.APIKey = decrypted
	}

	if exchange.APIKey != "" {
		decrypted, err := d.decryptValue(exchange.APIKey, exchange.UserID, exchange.ID, "api_key")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.APIKey = decrypted
	}
	if exchange.SecretKey != "" {
		decrypted, err := d.decryptValue(exchange.SecretKey, exchange.UserID, exchange.ID, "secret_key")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.SecretKey = decrypted
	}
	if exchange.HyperliquidWalletAddr != "" {
		decrypted, err := d.decryptValue(exchange.HyperliquidWalletAddr, exchange.UserID, exchange.ID, "hyperliquid_wallet_addr")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.HyperliquidWalletAddr = decrypted
	}
	if exchange.AsterUser != "" {
		decrypted, err := d.decryptValue(exchange.AsterUser, exchange.UserID, exchange.ID, "aster_user")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.AsterUser = decrypted
	}
	if exchange.AsterSigner != "" {
		decrypted, err := d.decryptValue(exchange.AsterSigner, exchange.UserID, exchange.ID, "aster_signer")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.AsterSigner = decrypted
	}
	if exchange.AsterPrivateKey != "" {
		decrypted, err := d.decryptValue(exchange.AsterPrivateKey, exchange.UserID, exchange.ID, "aster_private_key")
		if err != nil {
			return nil, nil, nil, err
		}
		exchange.AsterPrivateKey = decrypted
	}

	return &trader, &aiModel, &exchange, nil
}

// GetSystemConfig 获取系统配置
func (d *PostgreSQLDatabase) GetSystemConfig(key string) (string, error) {
	var value string
	err := d.db.QueryRow(`SELECT value FROM system_config WHERE key = $1`, key).Scan(&value)
	return value, err
}

// SetSystemConfig 设置系统配置
func (d *PostgreSQLDatabase) SetSystemConfig(key, value string) error {
	_, err := d.db.Exec(`
		INSERT INTO system_config (key, value) VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP
	`, key, value)
	return err
}

// CreateUserSignalSource 创建用户信号源配置
func (d *PostgreSQLDatabase) CreateUserSignalSource(userID, coinPoolURL, oiTopURL string) error {
	_, err := d.db.Exec(`
		INSERT INTO user_signal_sources (user_id, coin_pool_url, oi_top_url, updated_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id) DO UPDATE SET 
			coin_pool_url = $2, oi_top_url = $3, updated_at = CURRENT_TIMESTAMP
	`, userID, coinPoolURL, oiTopURL)
	return err
}

// GetUserSignalSource 获取用户信号源配置
func (d *PostgreSQLDatabase) GetUserSignalSource(userID string) (*UserSignalSource, error) {
	var source UserSignalSource
	err := d.db.QueryRow(`
		SELECT id, user_id, coin_pool_url, oi_top_url, created_at, updated_at
		FROM user_signal_sources WHERE user_id = $1
	`, userID).Scan(
		&source.ID, &source.UserID, &source.CoinPoolURL, &source.OITopURL,
		&source.CreatedAt, &source.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

// UpdateUserSignalSource 更新用户信号源配置
func (d *PostgreSQLDatabase) UpdateUserSignalSource(userID, coinPoolURL, oiTopURL string) error {
	_, err := d.db.Exec(`
		UPDATE user_signal_sources SET coin_pool_url = $1, oi_top_url = $2, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $3
	`, coinPoolURL, oiTopURL, userID)
	return err
}

// GetCustomCoins 获取所有交易员自定义币种
func (d *PostgreSQLDatabase) GetCustomCoins() []string {
	var symbol string
	var symbols []string

	err := d.db.QueryRow(`
		SELECT STRING_AGG(custom_coins, ',') as symbol
		FROM traders WHERE custom_coins != ''
	`).Scan(&symbol)

	// 检测用户是否未配置币种 - 兼容性
	if err != nil || symbol == "" {
		symbolJSON, _ := d.GetSystemConfig("default_coins")
		if err := json.Unmarshal([]byte(symbolJSON), &symbols); err != nil {
			log.Printf("⚠️  解析default_coins配置失败: %v，使用硬编码默认值", err)
			symbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT", "BNBUSDT"}
		}
	}

	// filter Symbol
	for _, s := range strings.Split(symbol, ",") {
		if s == "" {
			continue
		}
		coin := normalizeSymbol(s)
		if !slices.Contains(symbols, coin) {
			symbols = append(symbols, coin)
		}
	}
	return symbols
}

// normalizeSymbol 标准化交易对符号（确保以USDT结尾）
// 复制自 market.Normalize 以避免循环导入
func normalizeSymbol(symbol string) string {
	symbol = strings.ToUpper(symbol)
	if strings.HasSuffix(symbol, "USDT") {
		return symbol
	}
	return symbol + "USDT"
}

// LoadBetaCodesFromFile 从文件加载内测码到数据库
func (d *PostgreSQLDatabase) LoadBetaCodesFromFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取内测码文件失败: %w", err)
	}

	// 按行分割内测码
	lines := strings.Split(string(content), "\n")
	var codes []string
	for _, line := range lines {
		code := strings.TrimSpace(line)
		if code != "" && !strings.HasPrefix(code, "#") {
			codes = append(codes, code)
		}
	}

	// 批量插入内测码
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO beta_codes (code) VALUES ($1) ON CONFLICT (code) DO NOTHING`)
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	insertedCount := 0
	for _, code := range codes {
		result, err := stmt.Exec(code)
		if err != nil {
			log.Printf("插入内测码 %s 失败: %v", code, err)
			continue
		}

		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
			insertedCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	log.Printf("✅ 成功加载 %d 个内测码到数据库 (总计 %d 个)", insertedCount, len(codes))
	return nil
}

// ValidateBetaCode 验证内测码是否有效且未使用
func (d *PostgreSQLDatabase) ValidateBetaCode(code string) (bool, error) {
	var used bool
	err := d.db.QueryRow(`SELECT used FROM beta_codes WHERE code = $1`, code).Scan(&used)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // 内测码不存在
		}
		return false, err
	}
	return !used, nil // 内测码存在且未使用
}

// UseBetaCode 使用内测码（标记为已使用）
func (d *PostgreSQLDatabase) UseBetaCode(code, userEmail string) error {
	result, err := d.db.Exec(`
		UPDATE beta_codes SET used = true, used_by = $1, used_at = CURRENT_TIMESTAMP 
		WHERE code = $2 AND used = false
	`, userEmail, code)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("内测码无效或已被使用")
	}

	return nil
}

// GetBetaCodeStats 获取内测码统计信息
func (d *PostgreSQLDatabase) GetBetaCodeStats() (total, used int, err error) {
	err = d.db.QueryRow(`SELECT COUNT(*) FROM beta_codes`).Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	err = d.db.QueryRow(`SELECT COUNT(*) FROM beta_codes WHERE used = true`).Scan(&used)
	if err != nil {
		return 0, 0, err
	}

	return total, used, nil
}

// initDefaultData 初始化默认数据（AI模型和交易所）
func (d *PostgreSQLDatabase) initDefaultData() error {
	// 确保traders表存在custom_coins列，防止旧环境缺少字段
	if _, err := d.db.Exec(`ALTER TABLE traders ADD COLUMN IF NOT EXISTS custom_coins TEXT DEFAULT ''`); err != nil {
		return fmt.Errorf("添加custom_coins列失败: %w", err)
	}

	// 确保traders表存在kline_intervals列，防止旧环境缺少字段
	if _, err := d.db.Exec(`ALTER TABLE traders ADD COLUMN IF NOT EXISTS kline_intervals TEXT DEFAULT '1m,15m,4h'`); err != nil {
		return fmt.Errorf("添加kline_intervals列失败: %w", err)
	}

	// 确保exchanges表存在deleted列
	if _, err := d.db.Exec(`ALTER TABLE exchanges ADD COLUMN IF NOT EXISTS deleted BOOLEAN DEFAULT FALSE`); err != nil {
		return fmt.Errorf("添加deleted列失败: %w", err)
	}

	// 创建agent_wallets表（如果不存在）
	if _, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_wallets (
			id SERIAL PRIMARY KEY,
			main_wallet TEXT NOT NULL UNIQUE,
			agent_address TEXT NOT NULL UNIQUE,
			encrypted_private_key TEXT NOT NULL,
			authorization_signature TEXT DEFAULT '',
			status TEXT DEFAULT 'INIT',
			hyperliquid_chain TEXT DEFAULT 'Mainnet',
			builder_fee_authorized BOOLEAN DEFAULT FALSE,
			builder_fee_max_rate INTEGER DEFAULT 0,
			builder_fee_authorized_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("创建agent_wallets表失败: %w", err)
	}

	// 首先创建default用户（如果不存在）
	_, err := d.db.Exec(`
		INSERT INTO users (id, email, password_hash, otp_secret, otp_verified)
		VALUES ('default', 'default@localhost', '', '', true)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		return fmt.Errorf("创建default用户失败: %w", err)
	}

	// 初始化AI模型（使用default用户）
	aiModels := []struct {
		id, name, provider string
	}{
		{"deepseek", "DeepSeek", "deepseek"},
		{"qwen", "Qwen", "qwen"},
		{"gemini3", "Gemini 3", "gemini"},
	}

	for _, model := range aiModels {
		_, err := d.db.Exec(`
			INSERT INTO ai_models (id, user_id, name, provider, enabled, api_key, custom_api_url, custom_model_name, created_at, updated_at) 
			VALUES ($1, 'default', $2, $3, false, '', '', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (id) DO NOTHING
		`, model.id, model.name, model.provider)
		if err != nil {
			return fmt.Errorf("初始化AI模型失败: %w", err)
		}
	}

	// 初始化交易所（使用default用户）
	exchanges := []struct {
		id, name, typ string
	}{
		{"binance", "Binance Futures", "binance"},
		{"hyperliquid", "Hyperliquid", "hyperliquid"},
		{"aster", "Aster DEX", "aster"},
	}

	for _, exchange := range exchanges {
		_, err := d.db.Exec(`
			INSERT INTO exchanges (id, user_id, name, type, enabled, api_key, secret_key, testnet, 
				hyperliquid_wallet_addr, aster_user, aster_signer, aster_private_key, created_at, updated_at) 
			VALUES ($1, 'default', $2, $3, false, '', '', false, '', '', '', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			ON CONFLICT (id, user_id) DO NOTHING
		`, exchange.id, exchange.name, exchange.typ)
		if err != nil {
			return fmt.Errorf("初始化交易所失败: %w", err)
		}
	}

	return nil
}

// GetDB 获取底层数据库连接（用于直接 SQL 操作）
func (d *PostgreSQLDatabase) GetDB() interface{} {
	return d.db
}

// Close 关闭数据库连接
func (d *PostgreSQLDatabase) Close() error {
	return d.db.Close()
}
