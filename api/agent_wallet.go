package api

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nofx/crypto"
	"strings"
	"time"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// AgentWallet Agent 钱包模型
type AgentWallet struct {
	ID                     int        `json:"id"`
	MainWallet             string     `json:"main_wallet"`
	AgentAddress           string     `json:"agent_address"`
	EncryptedPrivateKey    string     `json:"-"` // 不返回给前端
	AuthorizationSignature string     `json:"-"` // 不返回给前端
	Status                 string     `json:"status"`
	HyperliquidChain       string     `json:"hyperliquid_chain"`
	BuilderFeeAuthorized   bool       `json:"builder_fee_authorized"`              // 是否已授权 Builder Fee
	BuilderFeeMaxRate      int        `json:"builder_fee_max_rate"`                // Builder Fee 最大费率（基点，100 = 0.1%）
	BuilderFeeAuthorizedAt *time.Time `json:"builder_fee_authorized_at,omitempty"` // Builder Fee 授权时间
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// CreateAgentWalletRequest 创建 Agent 钱包请求
type CreateAgentWalletRequest struct {
	MainWallet       string `json:"main_wallet" binding:"required"`
	HyperliquidChain string `json:"hyperliquid_chain"` // "Mainnet" or "Testnet"
}

// CreateAgentWalletResponse 创建 Agent 钱包响应
type CreateAgentWalletResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	AgentAddress string `json:"agent_address,omitempty"`
	MainWallet   string `json:"main_wallet,omitempty"`
	Status       string `json:"status,omitempty"`
}

// GetAgentWalletResponse 查询 Agent 钱包响应
type GetAgentWalletResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message,omitempty"`
	Data    *AgentWallet `json:"data,omitempty"`
}

// handleCreateAgentWallet 创建 Agent 钱包（后端生成私钥）
func (s *Server) handleCreateAgentWallet(c *gin.Context) {
	var req CreateAgentWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, CreateAgentWalletResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 规范化主钱包地址（小写）
	mainWallet := strings.ToLower(req.MainWallet)
	if !strings.HasPrefix(mainWallet, "0x") || len(mainWallet) != 42 {
		c.JSON(http.StatusBadRequest, CreateAgentWalletResponse{
			Success: false,
			Message: "Invalid Ethereum address format",
		})
		return
	}

	// 设置默认链
	hyperliquidChain := req.HyperliquidChain
	if hyperliquidChain == "" {
		hyperliquidChain = "Mainnet"
	}
	if hyperliquidChain != "Mainnet" && hyperliquidChain != "Testnet" {
		c.JSON(http.StatusBadRequest, CreateAgentWalletResponse{
			Success: false,
			Message: "Invalid hyperliquid_chain, must be 'Mainnet' or 'Testnet'",
		})
		return
	}

	// 检查是否已存在
	existingWallet, err := s.getAgentWallet(mainWallet)
	if err == nil && existingWallet != nil {
		c.JSON(http.StatusOK, CreateAgentWalletResponse{
			Success:      true,
			Message:      "Agent wallet already exists",
			AgentAddress: existingWallet.AgentAddress,
			MainWallet:   existingWallet.MainWallet,
			Status:       existingWallet.Status,
		})
		return
	}

	// 1. 生成新的 Agent 钱包私钥
	privateKey, err := ethcrypto.GenerateKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateAgentWalletResponse{
			Success: false,
			Message: "Failed to generate agent wallet: " + err.Error(),
		})
		return
	}

	// 2. 获取 Agent 地址
	agentAddress := ethcrypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// 3. 将私钥转换为十六进制字符串（不含 0x 前缀）
	privateKeyBytes := ethcrypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// 4. 使用 EncryptionManager 加密私钥
	em, err := crypto.GetEncryptionManager()
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateAgentWalletResponse{
			Success: false,
			Message: "Failed to get encryption manager: " + err.Error(),
		})
		return
	}

	encryptedPrivateKey, err := em.EncryptForDatabase(privateKeyHex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateAgentWalletResponse{
			Success: false,
			Message: "Failed to encrypt private key: " + err.Error(),
		})
		return
	}

	// 5. 保存到数据库
	query := `
		INSERT INTO agent_wallets (
			main_wallet,
			agent_address,
			encrypted_private_key,
			status,
			hyperliquid_chain
		) VALUES ($1, $2, $3, $4, $5)
	`

	db := s.database.GetDB().(*sqlx.DB)
	_, err = db.Exec(
		query,
		mainWallet,
		strings.ToLower(agentAddress),
		encryptedPrivateKey,
		"INIT",
		hyperliquidChain,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, CreateAgentWalletResponse{
			Success: false,
			Message: "Failed to save agent wallet: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, CreateAgentWalletResponse{
		Success:      true,
		Message:      "Agent wallet created successfully. Please authorize on Hyperliquid to activate.",
		AgentAddress: strings.ToLower(agentAddress),
		MainWallet:   mainWallet,
		Status:       "INIT",
	})
}

// handleGetAgentWallet 查询 Agent 钱包状态
func (s *Server) handleGetAgentWallet(c *gin.Context) {
	mainWallet := strings.ToLower(c.Query("main_wallet"))
	if mainWallet == "" {
		c.JSON(http.StatusBadRequest, GetAgentWalletResponse{
			Success: false,
			Message: "Missing main_wallet parameter",
		})
		return
	}

	wallet, err := s.getAgentWallet(mainWallet)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, GetAgentWalletResponse{
				Success: false,
				Message: "Agent wallet not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, GetAgentWalletResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GetAgentWalletResponse{
		Success: true,
		Data:    wallet,
	})
}

// getAgentWallet 从数据库获取 Agent 钱包
func (s *Server) getAgentWallet(mainWallet string) (*AgentWallet, error) {
	query := `
		SELECT
			id,
			main_wallet,
			agent_address,
			encrypted_private_key,
			authorization_signature,
			status,
			hyperliquid_chain,
			COALESCE(builder_fee_authorized, false) as builder_fee_authorized,
			COALESCE(builder_fee_max_rate, 0) as builder_fee_max_rate,
			builder_fee_authorized_at,
			created_at,
			updated_at
		FROM agent_wallets
		WHERE main_wallet = $1
	`

	db := s.database.GetDB().(*sqlx.DB)
	wallet := &AgentWallet{}
	err := db.QueryRow(query, strings.ToLower(mainWallet)).Scan(
		&wallet.ID,
		&wallet.MainWallet,
		&wallet.AgentAddress,
		&wallet.EncryptedPrivateKey,
		&wallet.AuthorizationSignature,
		&wallet.Status,
		&wallet.HyperliquidChain,
		&wallet.BuilderFeeAuthorized,
		&wallet.BuilderFeeMaxRate,
		&wallet.BuilderFeeAuthorizedAt,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// getAgentPrivateKey 获取解密后的 Agent 私钥（仅内部使用）
func (s *Server) getAgentPrivateKey(mainWallet string) (*ecdsa.PrivateKey, error) {
	wallet, err := s.getAgentWallet(mainWallet)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent wallet: %w", err)
	}

	// 解密私钥
	em, err := crypto.GetEncryptionManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get encryption manager: %w", err)
	}

	privateKeyHex, err := em.DecryptFromDatabase(wallet.EncryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	// 将十六进制字符串转换为私钥对象
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key hex: %w", err)
	}

	privateKey, err := ethcrypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to ECDSA private key: %w", err)
	}

	return privateKey, nil
}

// AuthorizeAgentRequest 授权 Agent 钱包请求
type AuthorizeAgentRequest struct {
	MainWallet   string `json:"main_wallet" binding:"required"`
	Signature    string `json:"signature" binding:"required"` // MainWallet 签名的 ApproveAgent 消息
	AgentName    string `json:"agent_name"`                   // 可选的 Agent 名称
	Nonce        uint64 `json:"nonce" binding:"required"`     // 用于签名的 nonce
	SignatureRSV struct {
		R string `json:"r" binding:"required"`
		S string `json:"s" binding:"required"`
		V int    `json:"v" binding:"required"`
	} `json:"signature_rsv" binding:"required"` // {r, s, v} 格式的签名
}

// AuthorizeAgentResponse 授权 Agent 钱包响应
type AuthorizeAgentResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  string `json:"status,omitempty"`
}

// handleAuthorizeAgent 授权 Agent 钱包（提交到 Hyperliquid）
func (s *Server) handleAuthorizeAgent(c *gin.Context) {
	var req AuthorizeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthorizeAgentResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 规范化主钱包地址
	mainWallet := strings.ToLower(req.MainWallet)

	// 1. 获取 Agent 钱包
	wallet, err := s.getAgentWallet(mainWallet)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, AuthorizeAgentResponse{
				Success: false,
				Message: "Agent wallet not found. Please create one first.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, AuthorizeAgentResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}

	// 2. 检查是否已授权
	if wallet.Status == "ACTIVE" {
		c.JSON(http.StatusOK, AuthorizeAgentResponse{
			Success: true,
			Message: "Agent wallet already authorized",
			Status:  "ACTIVE",
		})
		return
	}

	// 3. 构建 Hyperliquid API 请求
	hyperliquidAPI := "https://api.hyperliquid.xyz/exchange"
	if wallet.HyperliquidChain == "Testnet" {
		hyperliquidAPI = "https://api.hyperliquid-testnet.xyz/exchange"
	}

	// 构建 ApproveAgent action (flat structure, matching Python SDK)
	action := map[string]interface{}{
		"type":             "approveAgent",
		"signatureChainId": "0x66eee", // Hyperliquid L1 chain ID (matches Python SDK)
		"hyperliquidChain": wallet.HyperliquidChain,
		"agentAddress":     wallet.AgentAddress,
		"nonce":            req.Nonce,
	}

	// Add agentName only if non-empty (Python SDK behavior)
	if req.AgentName != "" {
		action["agentName"] = req.AgentName
	}

	// 构建完整请求体
	requestBody := map[string]interface{}{
		"action": action,
		"nonce":  req.Nonce,
		"signature": map[string]interface{}{
			"r": req.SignatureRSV.R,
			"s": req.SignatureRSV.S,
			"v": req.SignatureRSV.V,
		},
	}

	// 4. 发送到 Hyperliquid
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthorizeAgentResponse{
			Success: false,
			Message: "Failed to marshal request: " + err.Error(),
		})
		return
	}

	// 调试日志：打印发送给 Hyperliquid 的完整 payload
	log.Printf("🔍 [DEBUG] Sending to Hyperliquid API: %s", hyperliquidAPI)
	log.Printf("🔍 [DEBUG] Main Wallet: %s", mainWallet)
	log.Printf("🔍 [DEBUG] Agent Address: %s", wallet.AgentAddress)
	log.Printf("🔍 [DEBUG] Request Payload: %s", string(jsonData))

	resp, err := http.Post(hyperliquidAPI, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthorizeAgentResponse{
			Success: false,
			Message: "Failed to submit to Hyperliquid: " + err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// 5. 解析 Hyperliquid 响应
	var hyperliquidResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&hyperliquidResp); err != nil {
		c.JSON(http.StatusInternalServerError, AuthorizeAgentResponse{
			Success: false,
			Message: "Failed to parse Hyperliquid response: " + err.Error(),
		})
		return
	}

	// 调试日志：打印 Hyperliquid 响应
	respJSON, _ := json.Marshal(hyperliquidResp)
	log.Printf("🔍 [DEBUG] Hyperliquid Response: %s", string(respJSON))

	// 检查 Hyperliquid 是否返回错误
	if status, ok := hyperliquidResp["status"].(string); ok && status == "err" {
		errorMsg := "Unknown error"
		if response, ok := hyperliquidResp["response"].(string); ok {
			errorMsg = response
		}
		log.Printf("❌ [ERROR] Hyperliquid rejected: %s", errorMsg)
		c.JSON(http.StatusBadRequest, AuthorizeAgentResponse{
			Success: false,
			Message: "Hyperliquid rejected authorization: " + errorMsg,
		})
		return
	}

	// 6. 更新数据库状态
	updateQuery := `
		UPDATE agent_wallets
		SET authorization_signature = $1, status = 'ACTIVE'
		WHERE main_wallet = $2
	`

	db := s.database.GetDB().(*sqlx.DB)
	_, err = db.Exec(updateQuery, req.Signature, mainWallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthorizeAgentResponse{
			Success: false,
			Message: "Failed to update database: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, AuthorizeAgentResponse{
		Success: true,
		Message: "Agent wallet authorized successfully!",
		Status:  "ACTIVE",
	})
}

// ConfirmBuilderFeeRequest 确认 Builder Fee 授权请求
type ConfirmBuilderFeeRequest struct {
	MainWallet string `json:"main_wallet" binding:"required"`
	MaxFeeRate int    `json:"max_fee_rate" binding:"required"` // 基点，100 = 0.1%
}

// ConfirmBuilderFeeResponse 确认 Builder Fee 授权响应
type ConfirmBuilderFeeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// handleConfirmBuilderFee 确认 Builder Fee 授权（前端在成功授权后调用）
func (s *Server) handleConfirmBuilderFee(c *gin.Context) {
	var req ConfirmBuilderFeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ConfirmBuilderFeeResponse{
			Success: false,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	// 规范化主钱包地址
	mainWallet := strings.ToLower(req.MainWallet)

	// 1. 获取 Agent 钱包
	wallet, err := s.getAgentWallet(mainWallet)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, ConfirmBuilderFeeResponse{
				Success: false,
				Message: "Agent wallet not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, ConfirmBuilderFeeResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		})
		return
	}

	// 2. 检查是否已确认
	if wallet.BuilderFeeAuthorized {
		c.JSON(http.StatusOK, ConfirmBuilderFeeResponse{
			Success: true,
			Message: "Builder Fee already confirmed",
		})
		return
	}

	// 3. 更新数据库
	updateQuery := `
		UPDATE agent_wallets
		SET builder_fee_authorized = true,
		    builder_fee_max_rate = $1,
		    builder_fee_authorized_at = CURRENT_TIMESTAMP
		WHERE main_wallet = $2
	`

	db := s.database.GetDB().(*sqlx.DB)
	_, err = db.Exec(updateQuery, req.MaxFeeRate, mainWallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ConfirmBuilderFeeResponse{
			Success: false,
			Message: "Failed to update database: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ConfirmBuilderFeeResponse{
		Success: true,
		Message: "Builder Fee authorization confirmed successfully!",
	})
}
