package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	ProviderGemini       = "gemini"
	DefaultGeminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
	DefaultGeminiModel   = "gemini-3-pro-preview"
)

type GeminiClient struct {
	*Client
}

// NewGeminiClient 创建 Gemini 客户端（向前兼容）
//
// Deprecated: 推荐使用 NewGeminiClientWithOptions 以获得更好的灵活性
func NewGeminiClient() AIClient {
	return NewGeminiClientWithOptions()
}

// NewGeminiClientWithOptions 创建 Gemini 客户端（支持选项模式）
//
// 使用示例：
//
//	// 基础用法
//	client := mcp.NewGeminiClientWithOptions()
//
//	// 自定义配置
//	client := mcp.NewGeminiClientWithOptions(
//	    mcp.WithAPIKey("AIza..."),
//	    mcp.WithLogger(customLogger),
//	    mcp.WithTimeout(60*time.Second),
//	)
func NewGeminiClientWithOptions(opts ...ClientOption) AIClient {
	// 1. 创建 Gemini 预设选项
	geminiOpts := []ClientOption{
		WithProvider(ProviderGemini),
		WithModel(DefaultGeminiModel),
		WithBaseURL(DefaultGeminiBaseURL),
		WithUseFullURL(true), // Gemini 使用自定义 URL 结构
	}

	// 2. 合并用户选项（用户选项优先级更高）
	allOpts := append(geminiOpts, opts...)

	// 3. 创建基础客户端
	baseClient := NewClient(allOpts...).(*Client)

	// 4. 创建 Gemini 客户端
	geminiClient := &GeminiClient{
		Client: baseClient,
	}

	// 5. 设置 hooks 指向 GeminiClient（实现动态分派）
	baseClient.hooks = geminiClient

	return geminiClient
}

func (geminiClient *GeminiClient) SetAPIKey(apiKey string, customURL string, customModel string) {
	geminiClient.APIKey = apiKey

	if len(apiKey) > 8 {
		geminiClient.logger.Infof("🔧 [MCP] Gemini API Key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])
	}
	if customURL != "" {
		geminiClient.BaseURL = customURL
		geminiClient.logger.Infof("🔧 [MCP] Gemini 使用自定义 BaseURL: %s", customURL)
	} else {
		geminiClient.logger.Infof("🔧 [MCP] Gemini 使用默认 BaseURL: %s", geminiClient.BaseURL)
	}
	if customModel != "" {
		geminiClient.Model = customModel
		geminiClient.logger.Infof("🔧 [MCP] Gemini 使用自定义 Model: %s", customModel)
	} else {
		geminiClient.logger.Infof("🔧 [MCP] Gemini 使用默认 Model: %s", geminiClient.Model)
	}
}

// buildUrl 构建 Gemini 特定的 URL（API Key 作为查询参数）
func (geminiClient *GeminiClient) buildUrl() string {
	// Gemini API 格式: https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent?key={apiKey}
	return fmt.Sprintf("%s/models/%s:generateContent?key=%s", geminiClient.BaseURL, geminiClient.Model, geminiClient.APIKey)
}

// setAuthHeader Gemini 不需要 Authorization header（API Key 在 URL 中）
func (geminiClient *GeminiClient) setAuthHeader(reqHeaders http.Header) {
	// Gemini 通过 URL 参数传递 API Key，不需要设置 Authorization header
}

// buildMCPRequestBody 构建 Gemini 格式的请求体
func (geminiClient *GeminiClient) buildMCPRequestBody(systemPrompt, userPrompt string) map[string]any {
	// Gemini API 请求格式
	contents := []map[string]any{}

	// 如果有 system prompt，添加到 systemInstruction
	var systemInstruction map[string]any
	if systemPrompt != "" {
		systemInstruction = map[string]any{
			"parts": []map[string]string{
				{"text": systemPrompt},
			},
		}
	}

	// 添加用户消息
	contents = append(contents, map[string]any{
		"role": "user",
		"parts": []map[string]string{
			{"text": userPrompt},
		},
	})

	// 构建请求体
	requestBody := map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"temperature":  geminiClient.config.Temperature,
			"maxOutputTokens": geminiClient.MaxTokens,
		},
	}

	// 添加 system instruction（如果存在）
	if systemInstruction != nil {
		requestBody["systemInstruction"] = systemInstruction
	}

	return requestBody
}

// parseMCPResponse 解析 Gemini API 响应
func (geminiClient *GeminiClient) parseMCPResponse(body []byte) (string, error) {
	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		geminiClient.logger.Errorf("❌ [%s] JSON 反序列化失败: %v", geminiClient.String(), err)
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Candidates) == 0 {
		geminiClient.logger.Errorf("❌ [%s] API 返回空 candidates", geminiClient.String())
		return "", fmt.Errorf("API返回空响应")
	}

	if len(result.Candidates[0].Content.Parts) == 0 {
		geminiClient.logger.Errorf("❌ [%s] API 返回空 parts", geminiClient.String())
		return "", fmt.Errorf("API返回空响应")
	}

	text := result.Candidates[0].Content.Parts[0].Text
	geminiClient.logger.Infof("✓ [%s] 成功解析响应, 文本长度: %d chars", geminiClient.String(), len(text))
	return text, nil
}

// buildRequestBodyFromRequest 从 Request 对象构建 Gemini 格式的请求体
func (geminiClient *GeminiClient) buildRequestBodyFromRequest(req *Request) map[string]any {
	// 转换 Message 为 Gemini API 格式
	contents := make([]map[string]any, 0)
	var systemInstruction map[string]any

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			// Gemini 使用 systemInstruction 而不是 system message
			systemInstruction = map[string]any{
				"parts": []map[string]string{
					{"text": msg.Content},
				},
			}
		} else {
			// user 或 assistant 消息
			contents = append(contents, map[string]any{
				"role": msg.Role,
				"parts": []map[string]string{
					{"text": msg.Content},
				},
			})
		}
	}

	// 构建 generationConfig
	generationConfig := map[string]any{}

	if req.Temperature != nil {
		generationConfig["temperature"] = *req.Temperature
	} else {
		generationConfig["temperature"] = geminiClient.config.Temperature
	}

	if req.MaxTokens != nil {
		generationConfig["maxOutputTokens"] = *req.MaxTokens
	} else {
		generationConfig["maxOutputTokens"] = geminiClient.MaxTokens
	}

	if req.TopP != nil {
		generationConfig["topP"] = *req.TopP
	}

	if len(req.Stop) > 0 {
		generationConfig["stopSequences"] = req.Stop
	}

	// 构建请求体
	requestBody := map[string]any{
		"contents":         contents,
		"generationConfig": generationConfig,
	}

	// 添加 system instruction（如果存在）
	if systemInstruction != nil {
		requestBody["systemInstruction"] = systemInstruction
	}

	return requestBody
}

// callWithRequest 单次调用 AI API（使用 Request 对象）- 重写以支持 Gemini 格式
func (geminiClient *GeminiClient) callWithRequest(req *Request) (string, error) {
	// 打印当前 AI 配置
	geminiClient.logger.Infof("📡 [%s] 开始请求 AI Server: BaseURL=%s, Model=%s", geminiClient.String(), geminiClient.BaseURL, geminiClient.Model)
	geminiClient.logger.Debugf("[%s] Messages count: %d", geminiClient.String(), len(req.Messages))

	// 构建 Gemini 格式的请求体
	geminiClient.logger.Infof("🔧 [%s] 构建 Gemini 格式请求体...", geminiClient.String())
	requestBody := geminiClient.buildRequestBodyFromRequest(req)

	// 序列化请求体
	jsonData, err := geminiClient.marshalRequestBody(requestBody)
	if err != nil {
		geminiClient.logger.Errorf("❌ [%s] 序列化请求体失败: %v", geminiClient.String(), err)
		return "", err
	}
	geminiClient.logger.Infof("✓ [%s] 请求体构建成功, 大小: %d bytes", geminiClient.String(), len(jsonData))

	// 构建 URL
	url := geminiClient.buildUrl()
	geminiClient.logger.Infof("🌐 [%s] 请求 URL: %s", geminiClient.String(), url)

	// 输出请求体内容（用于调试）
	geminiClient.logger.Debugf("📤 [%s] 请求体: %s", geminiClient.String(), string(jsonData))

	// 创建 HTTP 请求
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		geminiClient.logger.Errorf("❌ [%s] 创建 HTTP 请求失败: %v", geminiClient.String(), err)
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	geminiClient.logger.Infof("📨 [%s] 发送 HTTP 请求...", geminiClient.String())

	// 发送 HTTP 请求
	resp, err := geminiClient.httpClient.Do(httpReq)
	if err != nil {
		geminiClient.logger.Errorf("❌ [%s] 发送 HTTP 请求失败: %v", geminiClient.String(), err)
		return "", fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		geminiClient.logger.Errorf("❌ [%s] 读取响应体失败: %v", geminiClient.String(), err)
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	// 输出响应体内容（用于调试）
	geminiClient.logger.Infof("📥 [%s] 收到响应: status=%d, size=%d bytes", geminiClient.String(), resp.StatusCode, len(body))
	geminiClient.logger.Debugf("📥 [%s] 响应体: %s", geminiClient.String(), string(body))

	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		geminiClient.logger.Errorf("❌ [%s] API 返回错误 (status %d): %s", geminiClient.String(), resp.StatusCode, string(body))
		return "", fmt.Errorf("API返回错误 (status %d): %s", resp.StatusCode, string(body))
	}

	// 解析响应
	geminiClient.logger.Infof("🔍 [%s] 解析 API 响应...", geminiClient.String())
	result, err := geminiClient.parseMCPResponse(body)
	if err != nil {
		geminiClient.logger.Errorf("❌ [%s] 解析响应失败: %v", geminiClient.String(), err)
		return "", fmt.Errorf("fail to parse AI server response: %w", err)
	}

	geminiClient.logger.Infof("✅ [%s] AI 请求成功完成, 响应长度: %d chars", geminiClient.String(), len(result))
	return result, nil
}
