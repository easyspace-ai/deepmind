// Package deerflow provides a Go client for interacting with the DeerFlow API.
//
// Example usage:
//
//	client := deerflow.NewClient(
//	    deerflow.WithBaseURL("http://localhost:8001"),
//	)
//
//	// Simple chat
//	response, err := client.Chat(context.Background(), "Hello, world!")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(response)
//
//	// Chat with thread
//	response, err := client.Chat(context.Background(), "Hello",
//	    deerflow.WithThreadID("my-thread-id"))
//
//	// Streaming chat
//	stream, err := client.Stream(context.Background(), "Tell me a story")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer stream.Close()
//
//	for event := range stream.Events() {
//	    switch e := event.(type) {
//	    case *deerflow.MessageEvent:
//	        fmt.Print(e.Content)
//	    case *deerflow.ToolCallEvent:
//	        fmt.Printf("\n[Tool: %s]\n", e.Name)
//	    }
//	}
package deerflow

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://localhost:8001"
	defaultTimeout = 60 * time.Second
)

// Client is the DeerFlow API client.
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// WithHTTPClient sets the HTTP client for the client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithAPIKey sets the API key for the client.
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
	}
}

// NewClient creates a new DeerFlow client.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ============================================
// Chat Messages API
// ============================================

// ChatRequest is the request for sending a chat message.
type ChatRequest struct {
	Message  string                 `json:"message"`
	ThreadID string                 `json:"thread_id,omitempty"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// ChatResponse is the response from a chat message.
type ChatResponse struct {
	ThreadID string                 `json:"thread_id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Chat sends a message to DeerFlow and waits for a complete response.
func (c *Client) Chat(ctx context.Context, message string, opts ...ChatOption) (*ChatResponse, error) {
	req := &ChatRequest{
		Message: message,
	}

	for _, opt := range opts {
		opt(req)
	}

	// Create thread first if no thread ID
	threadID := req.ThreadID
	if threadID == "" {
		thread, err := c.CreateThread(ctx)
		if err != nil {
			return nil, fmt.Errorf("create thread: %w", err)
		}
		threadID = thread.ThreadID
	}

	// Use streaming to get the full response
	stream, err := c.Stream(ctx, message, append(opts, WithThreadID(threadID))...)
	if err != nil {
		return nil, fmt.Errorf("start stream: %w", err)
	}
	defer stream.Close()

	var content strings.Builder
	for event := range stream.Events() {
		switch e := event.(type) {
		case *MessageEvent:
			content.WriteString(e.Content)
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	return &ChatResponse{
		ThreadID: threadID,
		Content:  content.String(),
	}, nil
}

// ChatOption configures a chat request.
type ChatOption func(*ChatRequest)

// WithThreadID sets the thread ID for the chat.
func WithThreadID(threadID string) ChatOption {
	return func(r *ChatRequest) {
		r.ThreadID = threadID
	}
}

// WithContext sets additional context for the chat.
func WithContext(ctx map[string]interface{}) ChatOption {
	return func(r *ChatRequest) {
		r.Context = ctx
	}
}

// ============================================
// Threads API
// ============================================

// Thread represents a conversation thread.
type Thread struct {
	ThreadID  string                 `json:"thread_id"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
	Values    map[string]interface{} `json:"values,omitempty"`
}

// CreateThread creates a new conversation thread.
func (c *Client) CreateThread(ctx context.Context) (*Thread, error) {
	u := fmt.Sprintf("%s/api/langgraph/threads", c.baseURL)

	reqBody := map[string]interface{}{}
	jsonBody, _ := json.Marshal(reqBody)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var thread Thread
	if err := json.NewDecoder(resp.Body).Decode(&thread); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &thread, nil
}

// GetThread gets an existing thread by ID.
func (c *Client) GetThread(ctx context.Context, threadID string) (*Thread, error) {
	u := fmt.Sprintf("%s/api/langgraph/threads/%s", c.baseURL, threadID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("thread not found: %s", threadID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var thread Thread
	if err := json.NewDecoder(resp.Body).Decode(&thread); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &thread, nil
}

// DeleteThread deletes a thread by ID.
func (c *Client) DeleteThread(ctx context.Context, threadID string) error {
	u := fmt.Sprintf("%s/api/langgraph/threads/%s", c.baseURL, threadID)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", u, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// ListThreadsRequest is the request for listing threads.
type ListThreadsRequest struct {
	Limit     int    `json:"limit,omitempty"`
	Offset    int    `json:"offset,omitempty"`
	SortBy    string `json:"sortBy,omitempty"`
	SortOrder string `json:"sortOrder,omitempty"`
}

// ListThreads lists threads with pagination.
func (c *Client) ListThreads(ctx context.Context, req *ListThreadsRequest) ([]*Thread, error) {
	if req == nil {
		req = &ListThreadsRequest{Limit: 50}
	}

	u := fmt.Sprintf("%s/api/langgraph/threads/search", c.baseURL)

	jsonBody, _ := json.Marshal(req)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var threads []*Thread
	if err := json.NewDecoder(resp.Body).Decode(&threads); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return threads, nil
}

// ============================================
// Streaming API
// ============================================

// Stream represents a streaming response.
type Stream struct {
	events chan Event
	err    error
	closer io.Closer
}

// Event is a streaming event.
type Event interface {
	EventType() string
}

// MessageEvent is an event containing message content.
type MessageEvent struct {
	Content string `json:"content"`
}

// EventType returns the event type.
func (e *MessageEvent) EventType() string { return "message" }

// ToolCallEvent is an event containing a tool call.
type ToolCallEvent struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// EventType returns the event type.
func (e *ToolCallEvent) EventType() string { return "tool_call" }

// ToolResultEvent is an event containing a tool result.
type ToolResultEvent struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
}

// EventType returns the event type.
func (e *ToolResultEvent) EventType() string { return "tool_result" }

// MetadataEvent is an event containing metadata.
type MetadataEvent struct {
	RunID    string                 `json:"run_id"`
	ThreadID string                 `json:"thread_id"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// EventType returns the event type.
func (e *MetadataEvent) EventType() string { return "metadata" }

// FinishEvent is an event indicating the stream has finished.
type FinishEvent struct {
	Status string `json:"status"`
}

// EventType returns the event type.
func (e *FinishEvent) EventType() string { return "finish" }

// Events returns the channel of events.
func (s *Stream) Events() <-chan Event {
	return s.events
}

// Err returns any error that occurred during streaming.
func (s *Stream) Err() error {
	return s.err
}

// Close closes the stream.
func (s *Stream) Close() error {
	if s.closer != nil {
		return s.closer.Close()
	}
	return nil
}

// StreamRequest is the request for streaming chat.
type StreamRequest struct {
	AssistantID string                 `json:"assistant_id,omitempty"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Messages    []interface{}          `json:"messages,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
}

// Stream starts a streaming chat with DeerFlow.
func (c *Client) Stream(ctx context.Context, message string, opts ...ChatOption) (*Stream, error) {
	req := &ChatRequest{
		Message: message,
	}

	for _, opt := range opts {
		opt(req)
	}

	// Create thread first if no thread ID
	threadID := req.ThreadID
	if threadID == "" {
		thread, err := c.CreateThread(ctx)
		if err != nil {
			return nil, fmt.Errorf("create thread: %w", err)
		}
		threadID = thread.ThreadID
	}

	// Build stream request
	streamReq := &StreamRequest{
		AssistantID: "lead_agent",
		Messages: []interface{}{
			map[string]interface{}{
				"type": "human",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": message,
					},
				},
			},
		},
		Context: req.Context,
	}

	u := fmt.Sprintf("%s/api/langgraph/threads/%s/runs/stream", c.baseURL, threadID)
	jsonBody, _ := json.Marshal(streamReq)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", u, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	stream := &Stream{
		events: make(chan Event, 100),
		closer: resp.Body,
	}

	// Start reading SSE in goroutine
	go func() {
		defer close(stream.events)
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		var eventType string
		var dataBuffer bytes.Buffer

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				// End of event
				if eventType != "" && dataBuffer.Len() > 0 {
					event, err := parseEvent(eventType, dataBuffer.Bytes())
					if err == nil {
						stream.events <- event
					}
				}
				eventType = ""
				dataBuffer.Reset()
				continue
			}

			if strings.HasPrefix(line, "event: ") {
				eventType = strings.TrimPrefix(line, "event: ")
			} else if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				dataBuffer.WriteString(data)
			}
		}

		if err := scanner.Err(); err != nil {
			stream.err = err
		}
	}()

	return stream, nil
}

func parseEvent(eventType string, data []byte) (Event, error) {
	switch eventType {
	case "metadata":
		var meta MetadataEvent
		if err := json.Unmarshal(data, &meta); err == nil {
			return &meta, nil
		}
	case "created":
		// Parse as metadata
		var created map[string]interface{}
		if err := json.Unmarshal(data, &created); err == nil {
			meta := &MetadataEvent{}
			if runID, ok := created["run_id"].(string); ok {
				meta.RunID = runID
			}
			if threadID, ok := created["thread_id"].(string); ok {
				meta.ThreadID = threadID
			}
			return meta, nil
		}
	case "updates":
		// Parse updates for messages
		var updates map[string]interface{}
		if err := json.Unmarshal(data, &updates); err == nil {
			if root, ok := updates["__root__"].(map[string]interface{}); ok {
				if msgs, ok := root["messages"].([]interface{}); ok && len(msgs) > 0 {
					lastMsg := msgs[len(msgs)-1]
					if msgMap, ok := lastMsg.(map[string]interface{}); ok {
						if content, ok := msgMap["content"].(string); ok {
							return &MessageEvent{Content: content}, nil
						}
					}
				}
			}
		}
	case "finish":
		var finish FinishEvent
		if err := json.Unmarshal(data, &finish); err == nil {
			return &finish, nil
		}
	case "events":
		var eventMap map[string]interface{}
		if err := json.Unmarshal(data, &eventMap); err == nil {
			if evt, ok := eventMap["event"].(string); ok {
				if evt == "on_tool_end" {
					if name, ok := eventMap["name"].(string); ok {
						return &ToolResultEvent{
							Name:   name,
							Result: eventMap["data"],
						}, nil
					}
				}
			}
		}
	}

	// Return a generic metadata event as fallback
	return &MetadataEvent{
		Extra: map[string]interface{}{
			"raw_type": eventType,
			"raw_data": string(data),
		},
	}, nil
}

// ============================================
// Models API
// ============================================

// Model represents an available model.
type Model struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Model            string `json:"model"`
	DisplayName      string `json:"display_name"`
	SupportsThinking bool   `json:"supports_thinking"`
}

// ListModelsResponse is the response for listing models.
type ListModelsResponse struct {
	Models []Model `json:"models"`
}

// ListModels lists available models.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	u := fmt.Sprintf("%s/api/models", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var listResp ListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return listResp.Models, nil
}

// ============================================
// Skills API
// ============================================

// Skill represents an available skill.
type Skill struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	License     *string `json:"license"`
	Category    string  `json:"category"`
	Enabled     bool    `json:"enabled"`
}

// ListSkillsResponse is the response for listing skills.
type ListSkillsResponse struct {
	Skills []Skill `json:"skills"`
}

// ListSkills lists available skills.
func (c *Client) ListSkills(ctx context.Context) ([]Skill, error) {
	u := fmt.Sprintf("%s/api/skills", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var listResp ListSkillsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return listResp.Skills, nil
}

// ============================================
// Agents API
// ============================================

// Agent represents an available agent.
type Agent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Model       string `json:"model"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListAgentsResponse is the response for listing agents.
type ListAgentsResponse struct {
	Agents []Agent `json:"agents"`
}

// ListAgents lists available agents.
func (c *Client) ListAgents(ctx context.Context) ([]Agent, error) {
	u := fmt.Sprintf("%s/api/agents", c.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var listResp ListAgentsResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return listResp.Agents, nil
}

// GetAgent gets an agent by name.
func (c *Client) GetAgent(ctx context.Context, name string) (*Agent, error) {
	u := fmt.Sprintf("%s/api/agents/%s", c.baseURL, url.PathEscape(name))

	httpReq, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setAuthHeader(httpReq)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("agent not found: %s", name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var agent Agent
	if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &agent, nil
}

func (c *Client) setAuthHeader(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
}
