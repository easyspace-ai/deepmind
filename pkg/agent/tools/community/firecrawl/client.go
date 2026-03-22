package firecrawl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ============================================
// Firecrawl Client（一比一复刻 DeerFlow）
// ============================================

// Client Firecrawl API 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建 Firecrawl 客户端
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Web []SearchResultWeb `json:"web,omitempty"`
}

// SearchResultWeb 搜索结果
type SearchResultWeb struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

// Search 执行搜索
func (c *Client) Search(query string, limit int) (*SearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("firecrawl api key not configured")
	}

	reqBody := SearchRequest{
		Query: query,
		Limit: limit,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.firecrawl.dev/v0/search", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ScrapeRequest 爬取请求
type ScrapeRequest struct {
	Formats []string `json:"formats"`
}

// ScrapeResponse 爬取响应
type ScrapeResponse struct {
	Markdown string         `json:"markdown,omitempty"`
	Metadata ScrapeMetadata `json:"metadata,omitempty"`
}

// ScrapeMetadata 爬取元数据
type ScrapeMetadata struct {
	Title string `json:"title,omitempty"`
}

// Scrape 爬取网页
func (c *Client) Scrape(url string, formats []string) (*ScrapeResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("firecrawl api key not configured")
	}

	reqBody := ScrapeRequest{
		Formats: formats,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.firecrawl.dev/v0/scrape/%s", url), bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (status %d): %s", resp.StatusCode, string(body))
	}

	var result ScrapeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
