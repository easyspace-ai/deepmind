package tavily

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ============================================
// Tavily Client（一比一复刻 DeerFlow）
// ============================================

// Client Tavily API 客户端
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient 创建 Tavily 客户端
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
	Query       string `json:"query"`
	MaxResults  int    `json:"max_results,omitempty"`
	SearchDepth string `json:"search_depth,omitempty"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query   string `json:"query"`
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
		Score   float64 `json:"score,omitempty"`
	} `json:"results"`
}

// Search 执行搜索
func (c *Client) Search(query string, maxResults int) (*SearchResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("tavily api key not configured")
	}

	reqBody := SearchRequest{
		Query:      query,
		MaxResults: maxResults,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.tavily.com/search", bytes.NewReader(jsonBody))
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

// ExtractRequest 提取请求
type ExtractRequest struct {
	URLs []string `json:"urls"`
}

// ExtractResponse 提取响应
type ExtractResponse struct {
	Results []struct {
		URL        string `json:"url"`
		Title      string `json:"title"`
		RawContent string `json:"raw_content"`
	} `json:"results"`
	FailedResults []struct {
		URL   string `json:"url"`
		Error string `json:"error"`
	} `json:"failed_results,omitempty"`
}

// Extract 提取网页内容
func (c *Client) Extract(urls []string) (*ExtractResponse, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("tavily api key not configured")
	}

	reqBody := ExtractRequest{
		URLs: urls,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.tavily.com/extract", bytes.NewReader(jsonBody))
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

	var result ExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
