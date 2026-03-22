package jinaai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ============================================
// Jina AI Client（一比一复刻 DeerFlow）
// ============================================

// Client Jina AI 客户端
type Client struct {
	apiKey string
}

// NewClient 创建 Jina AI 客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
	}
}

// CrawlRequest 爬取请求
type CrawlRequest struct {
	URL string `json:"url"`
}

// Crawl 爬取网页
func (c *Client) Crawl(url string, returnFormat string, timeout int) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	reqBody := CrawlRequest{
		URL: url,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://r.jina.ai/", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Return-Format", returnFormat)
	req.Header.Set("X-Timeout", fmt.Sprintf("%d", timeout))

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("jina api returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if len(body) == 0 || len(bytes.TrimSpace(body)) == 0 {
		return "", fmt.Errorf("jina api returned empty response")
	}

	return string(body), nil
}
