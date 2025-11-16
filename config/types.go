package config

import (
	"fmt"
	"nofx/crypto"
	"time"
)

// DatabaseInterface 数据库接口
type DatabaseInterface interface {
	// 用户相关
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	GetUserByID(userID string) (*User, error)
	UpdateUserOTPVerified(userID string, verified bool) error
	UpdateUserPassword(userID, newPasswordHash string) error
	
	// AI模型配置
	GetAIModelsByUserID(userID string) ([]AIModelConfig, error)
	CreateAIModel(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error
	UpdateAIModel(model *AIModelConfig) error
	DeleteAIModel(modelID, userID string) error
	
	// 交易所配置
	GetExchangesByUserID(userID string) ([]ExchangeConfig, error)
	CreateExchange(exchange *ExchangeConfig) error
	UpdateExchange(exchange *ExchangeConfig) error
	DeleteExchange(exchangeID, userID string) error
	
	// 交易员记录
	GetTradersByUserID(userID string) ([]TraderRecord, error)
	CreateTrader(trader *TraderRecord) error
	UpdateTrader(trader *TraderRecord) error
	DeleteTrader(traderID, userID string) error
	GetTraderByID(traderID, userID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error)
	GetTraderConfig(traderID, userID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error)
	UpdateTraderInitialBalance(traderID, userID string, balance float64) error
	UpdateTraderStatus(traderID, userID string, isRunning bool) error
	UpdateTraderCustomPrompt(traderID, userID string, customPrompt string, overrideBase bool, systemTemplate string) error
	
	// 信号源
	GetUserSignalSourceByUserID(userID string) (*UserSignalSource, error)
	GetUserSignalSource(userID string) (*UserSignalSource, error)
	CreateUserSignalSource(source *UserSignalSource) error
	
	// 系统配置和管理方法
	GetAllUsers() ([]User, error)
	GetTraders(userID string) ([]*TraderRecord, error)
	GetSystemConfig(key string) (interface{}, error)
	SetSystemConfig(key string, value interface{}) error
	GetAIModels(userID string) ([]*AIModelConfig, error)
	GetExchanges(userID string) ([]*ExchangeConfig, error)
	
	// 内测码相关
	VerifyBetaCode(code string) (bool, error)
	ValidateBetaCode(code string) (bool, error)
	UseBetaCode(code, userEmail string) error
	LoadBetaCodesFromFile(filename string) error
	GetBetaCodeStats() (total int, used int, err error)
	
	// 连接管理和配置
	SetCryptoService(cryptoService interface{}) error
	GetCustomCoins() ([]string, error)
	Close() error
}

// Database 数据库类型别名，兼容现有代码
type Database = DatabaseInterface

// NewDatabase 创建数据库实例
func NewDatabase() (DatabaseInterface, error) {
	pgDB, err := NewPostgreSQLDatabase()
	if err != nil {
		return nil, fmt.Errorf("创建PostgreSQL数据库失败: %w", err)
	}
	return &DatabaseWrapper{impl: pgDB}, nil
}

// DatabaseWrapper 包装PostgreSQLDatabase以实现接口
type DatabaseWrapper struct {
	impl *PostgreSQLDatabase
}

// 实现最基础的接口方法来通过编译，其他方法使用stub
func (w *DatabaseWrapper) CreateUser(user *User) error { 
	return w.impl.CreateUser(user)
}
func (w *DatabaseWrapper) GetUserByEmail(email string) (*User, error) { 
	return w.impl.GetUserByEmail(email)
}
func (w *DatabaseWrapper) GetUserByID(userID string) (*User, error) { 
	return w.impl.GetUserByID(userID)
}
func (w *DatabaseWrapper) UpdateUserOTPVerified(userID string, verified bool) error { 
	return w.impl.UpdateUserOTPVerified(userID, verified)
}
func (w *DatabaseWrapper) UpdateUserPassword(userID, hash string) error { 
	return w.impl.UpdateUserPassword(userID, hash)
}
func (w *DatabaseWrapper) GetAIModelsByUserID(userID string) ([]AIModelConfig, error) { 
	models, err := w.impl.GetAIModels(userID)
	if err != nil {
		return nil, err
	}
	var result []AIModelConfig
	for _, model := range models {
		result = append(result, *model)
	}
	return result, nil
}
func (w *DatabaseWrapper) CreateAIModel(userID, id, name, provider string, enabled bool, apiKey, customAPIURL string) error { 
	return w.impl.CreateAIModel(userID, id, name, provider, enabled, apiKey, customAPIURL)
}
func (w *DatabaseWrapper) UpdateAIModel(model *AIModelConfig) error { 
	return w.impl.UpdateAIModel(model.UserID, model.ID, model.Enabled, model.APIKey, model.CustomAPIURL, model.CustomModelName)
}
func (w *DatabaseWrapper) DeleteAIModel(modelID, userID string) error { 
	return w.impl.UpdateAIModel(userID, modelID, false, "", "", "")
}
func (w *DatabaseWrapper) GetExchangesByUserID(userID string) ([]ExchangeConfig, error) { 
	exchanges, err := w.impl.GetExchanges(userID)
	if err != nil {
		return nil, err
	}
	var result []ExchangeConfig
	for _, exchange := range exchanges {
		result = append(result, *exchange)
	}
	return result, nil
}
func (w *DatabaseWrapper) CreateExchange(exchange *ExchangeConfig) error { 
	return w.impl.CreateExchange(exchange.UserID, exchange.ID, exchange.Name, exchange.Type, exchange.Enabled, exchange.APIKey, exchange.SecretKey, exchange.Testnet, exchange.HyperliquidWalletAddr, exchange.AsterUser, exchange.AsterSigner, exchange.AsterPrivateKey)
}
func (w *DatabaseWrapper) UpdateExchange(exchange *ExchangeConfig) error { 
	return w.impl.UpdateExchange(exchange.UserID, exchange.ID, exchange.Enabled, exchange.APIKey, exchange.SecretKey, exchange.Testnet, exchange.HyperliquidWalletAddr, exchange.AsterUser, exchange.AsterSigner, exchange.AsterPrivateKey)
}
func (w *DatabaseWrapper) DeleteExchange(exchangeID, userID string) error { 
	return w.impl.UpdateExchange(userID, exchangeID, false, "", "", false, "", "", "", "")
}
func (w *DatabaseWrapper) GetTradersByUserID(userID string) ([]TraderRecord, error) { 
	traders, err := w.impl.GetTraders(userID)
	if err != nil {
		return nil, err
	}
	var result []TraderRecord
	for _, trader := range traders {
		result = append(result, *trader)
	}
	return result, nil
}
func (w *DatabaseWrapper) CreateTrader(trader *TraderRecord) error { 
	return w.impl.CreateTrader(trader)
}
func (w *DatabaseWrapper) UpdateTrader(trader *TraderRecord) error { 
	return w.impl.UpdateTrader(trader)
}
func (w *DatabaseWrapper) DeleteTrader(traderID, userID string) error { 
	return w.impl.DeleteTrader(traderID, userID)
}
func (w *DatabaseWrapper) GetTraderByID(traderID, userID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error) { 
	return w.impl.GetTraderConfig(traderID, userID)
}
func (w *DatabaseWrapper) GetTraderConfig(traderID, userID string) (*TraderRecord, *AIModelConfig, *ExchangeConfig, error) { 
	return w.impl.GetTraderConfig(traderID, userID)
}
func (w *DatabaseWrapper) UpdateTraderInitialBalance(traderID, userID string, balance float64) error { 
	return w.impl.UpdateTraderInitialBalance(traderID, userID, balance)
}
func (w *DatabaseWrapper) UpdateTraderStatus(traderID, userID string, isRunning bool) error { 
	return w.impl.UpdateTraderStatus(traderID, userID, isRunning)
}
func (w *DatabaseWrapper) UpdateTraderCustomPrompt(traderID, userID, prompt string, override bool, template string) error { 
	return w.impl.UpdateTraderCustomPrompt(userID, traderID, prompt, override)
}
func (w *DatabaseWrapper) GetUserSignalSourceByUserID(userID string) (*UserSignalSource, error) { 
	return w.impl.GetUserSignalSource(userID)
}
func (w *DatabaseWrapper) GetUserSignalSource(userID string) (*UserSignalSource, error) { 
	return w.impl.GetUserSignalSource(userID)
}
func (w *DatabaseWrapper) CreateUserSignalSource(source *UserSignalSource) error { 
	return w.impl.CreateUserSignalSource(source.UserID, source.CoinPoolURL, source.OITopURL)
}
func (w *DatabaseWrapper) GetAllUsers() ([]User, error) { 
	users, err := w.impl.GetAllUsers()
	if err != nil {
		return nil, err
	}
	// Convert []string to []User if needed
	var result []User
	for _, userID := range users {
		result = append(result, User{ID: userID})
	}
	return result, nil
}
func (w *DatabaseWrapper) GetTraders(userID string) ([]*TraderRecord, error) { 
	return w.impl.GetTraders(userID)
}
func (w *DatabaseWrapper) GetSystemConfig(key string) (interface{}, error) { 
	result, err := w.impl.GetSystemConfig(key)
	return result, err
}
func (w *DatabaseWrapper) SetSystemConfig(key string, value interface{}) error { 
	if str, ok := value.(string); ok {
		return w.impl.SetSystemConfig(key, str)
	}
	return w.impl.SetSystemConfig(key, fmt.Sprintf("%v", value))
}
func (w *DatabaseWrapper) GetAIModels(userID string) ([]*AIModelConfig, error) { 
	return w.impl.GetAIModels(userID)
}
func (w *DatabaseWrapper) GetExchanges(userID string) ([]*ExchangeConfig, error) { 
	return w.impl.GetExchanges(userID)
}
func (w *DatabaseWrapper) VerifyBetaCode(code string) (bool, error) { 
	return w.impl.ValidateBetaCode(code)
}
func (w *DatabaseWrapper) ValidateBetaCode(code string) (bool, error) { 
	return w.impl.ValidateBetaCode(code)
}
func (w *DatabaseWrapper) UseBetaCode(code, userEmail string) error { 
	return w.impl.UseBetaCode(code, userEmail)
}
func (w *DatabaseWrapper) LoadBetaCodesFromFile(filename string) error { 
	return w.impl.LoadBetaCodesFromFile(filename)
}
func (w *DatabaseWrapper) GetBetaCodeStats() (int, int, error) { 
	return w.impl.GetBetaCodeStats()
}
func (w *DatabaseWrapper) SetCryptoService(cryptoService interface{}) error { 
	if cs, ok := cryptoService.(*crypto.CryptoService); ok {
		w.impl.SetCryptoService(cs)
	}
	return nil
}
func (w *DatabaseWrapper) GetCustomCoins() ([]string, error) { 
	return w.impl.GetCustomCoins(), nil
}
func (w *DatabaseWrapper) Close() error { 
	return w.impl.Close()
}

// User 用户结构
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	OTPSecret    string    `db:"otp_secret" json:"-"`
	OTPVerified  bool      `db:"otp_verified" json:"otp_verified"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// AIModelConfig AI模型配置
type AIModelConfig struct {
	ID                string    `db:"id" json:"id"`
	UserID            string    `db:"user_id" json:"user_id"`
	Name              string    `db:"name" json:"name"`
	Provider          string    `db:"provider" json:"provider"`
	Enabled           bool      `db:"enabled" json:"enabled"`
	APIKey            string    `db:"api_key" json:"-"`
	CustomAPIURL      string    `db:"custom_api_url" json:"custom_api_url"`
	CustomModelName   string    `db:"custom_model_name" json:"custom_model_name"`
	CreatedAt         time.Time `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time `db:"updated_at" json:"updated_at"`
}

// ExchangeConfig 交易所配置
type ExchangeConfig struct {
	ID                     string    `db:"id" json:"id"`
	UserID                 string    `db:"user_id" json:"user_id"`
	Name                   string    `db:"name" json:"name"`
	Type                   string    `db:"type" json:"type"`
	Enabled                bool      `db:"enabled" json:"enabled"`
	APIKey                 string    `db:"api_key" json:"-"`
	SecretKey              string    `db:"secret_key" json:"-"`
	Testnet                bool      `db:"testnet" json:"testnet"`
	HyperliquidWalletAddr  string    `db:"hyperliquid_wallet_addr" json:"-"`
	AsterUser              string    `db:"aster_user" json:"-"`
	AsterSigner            string    `db:"aster_signer" json:"-"`
	AsterPrivateKey        string    `db:"aster_private_key" json:"-"`
	DEXWalletPrivateKey    string    `db:"dex_wallet_private_key" json:"-"`
	Deleted                bool      `db:"deleted" json:"deleted"`
	CreatedAt              time.Time `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time `db:"updated_at" json:"updated_at"`
}

// TraderRecord 交易员记录
type TraderRecord struct {
	ID                   string    `db:"id" json:"id"`
	UserID               string    `db:"user_id" json:"user_id"`
	Name                 string    `db:"name" json:"name"`
	AIModelID            string    `db:"ai_model_id" json:"ai_model_id"`
	ExchangeID           string    `db:"exchange_id" json:"exchange_id"`
	InitialBalance       float64   `db:"initial_balance" json:"initial_balance"`
	ScanIntervalMinutes  int       `db:"scan_interval_minutes" json:"scan_interval_minutes"`
	IsRunning            bool      `db:"is_running" json:"is_running"`
	BTCETHLeverage       int       `db:"btc_eth_leverage" json:"btc_eth_leverage"`
	AltcoinLeverage      int       `db:"altcoin_leverage" json:"altcoin_leverage"`
	TradingSymbols       string    `db:"trading_symbols" json:"trading_symbols"`
	UseCoinPool          bool      `db:"use_coin_pool" json:"use_coin_pool"`
	UseOITop             bool      `db:"use_oi_top" json:"use_oi_top"`
	CustomPrompt         string    `db:"custom_prompt" json:"custom_prompt"`
	OverrideBasePrompt   bool      `db:"override_base_prompt" json:"override_base_prompt"`
	SystemPromptTemplate string    `db:"system_prompt_template" json:"system_prompt_template"`
	IsCrossMargin        bool      `db:"is_cross_margin" json:"is_cross_margin"`
	KlineIntervals       string    `db:"kline_intervals" json:"kline_intervals"`       // K线时间间隔配置，格式如 "3m,4h" 或 "5m,1h,1d"
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
}

// UserSignalSource 用户信号源
type UserSignalSource struct {
	ID          string    `db:"id" json:"id"`
	UserID      string    `db:"user_id" json:"user_id"`
	CoinPoolURL string    `db:"coin_pool_url" json:"coin_pool_url"`
	OITopURL    string    `db:"oi_top_url" json:"oi_top_url"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}