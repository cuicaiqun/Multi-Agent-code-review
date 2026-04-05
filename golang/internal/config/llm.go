package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMClient 大模型API客户端
// 通过OpenAI兼容接口调用MiniMax或其他大模型
type LLMClient struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewLLMClient 创建LLM客户端，从环境变量读取配置
func NewLLMClient() *LLMClient {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	baseURL := os.Getenv("MINIMAX_API_BASE")
	model := os.Getenv("LLM_MODEL")

	if baseURL == "" {
		baseURL = "https://api.minimax.chat/v1"
	}
	if model == "" {
		model = "MiniMax-M2.7"
	}

	return &LLMClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Chat 发送聊天请求
func (c *LLMClient) Chat(systemPrompt, userMessage string) (string, error) {
	if c.APIKey == "" {
		return "", fmt.Errorf("未配置LLM API Key，请设置环境变量 MINIMAX_API_KEY")
	}

	reqBody := chatRequest{
		Model: c.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.1,
		MaxTokens:   4096,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API调用失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API返回空结果")
	}

	return chatResp.Choices[0].Message.Content, nil
}
