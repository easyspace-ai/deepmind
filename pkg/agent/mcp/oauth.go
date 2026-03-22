package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/weibaohui/nanobot-go/pkg/config"
	"go.uber.org/zap"
)

// ============================================
// MCP OAuth（一比一复刻 DeerFlow）
// ============================================

// OAuthToken 缓存的 OAuth Token
type OAuthToken struct {
	AccessToken string
	TokenType   string
	ExpiresAt   time.Time
}

// OAuthTokenManager OAuth Token 管理器
// 一比一复刻 DeerFlow 的 OAuthTokenManager
type OAuthTokenManager struct {
	oauthByServer map[string]*config.OAuthConfig
	tokens        map[string]*OAuthToken
	locks         map[string]*sync.Mutex
	logger        *zap.Logger
}

// NewOAuthTokenManager 从扩展配置创建 OAuthTokenManager
func NewOAuthTokenManager(extensionsConfig *config.ExtensionsConfig, logger *zap.Logger) *OAuthTokenManager {
	oauthByServer := make(map[string]*config.OAuthConfig)
	locks := make(map[string]*sync.Mutex)

	for serverName, serverConfig := range extensionsConfig.MCPServers {
		if serverConfig.OAuth != nil && serverConfig.OAuth.Enabled {
			oauthByServer[serverName] = serverConfig.OAuth
			locks[serverName] = &sync.Mutex{}
		}
	}

	return &OAuthTokenManager{
		oauthByServer: oauthByServer,
		tokens:        make(map[string]*OAuthToken),
		locks:         locks,
		logger:        logger,
	}
}

// HasOAuthServers 检查是否有 OAuth 服务器
func (m *OAuthTokenManager) HasOAuthServers() bool {
	return len(m.oauthByServer) > 0
}

// OAuthServerNames 获取 OAuth 服务器名称列表
func (m *OAuthTokenManager) OAuthServerNames() []string {
	names := make([]string, 0, len(m.oauthByServer))
	for name := range m.oauthByServer {
		names = append(names, name)
	}
	return names
}

// GetAuthorizationHeader 获取 Authorization header
// 一比一复刻 DeerFlow 的 get_authorization_header
func (m *OAuthTokenManager) GetAuthorizationHeader(serverName string) (string, error) {
	oauth := m.oauthByServer[serverName]
	if oauth == nil {
		return "", nil
	}

	// 检查缓存的 token 是否有效
	token := m.tokens[serverName]
	if token != nil && !m.isExpiring(token, oauth) {
		return fmt.Sprintf("%s %s", token.TokenType, token.AccessToken), nil
	}

	// 需要刷新或获取新 token
	lock := m.locks[serverName]
	if lock == nil {
		lock = &sync.Mutex{}
		m.locks[serverName] = lock
	}

	lock.Lock()
	defer lock.Unlock()

	// 再次检查（双重检查锁定）
	token = m.tokens[serverName]
	if token != nil && !m.isExpiring(token, oauth) {
		return fmt.Sprintf("%s %s", token.TokenType, token.AccessToken), nil
	}

	// 获取新 token
	fresh, err := m.fetchToken(oauth)
	if err != nil {
		return "", err
	}

	m.tokens[serverName] = fresh

	if m.logger != nil {
		m.logger.Info("Refreshed OAuth access token for MCP server",
			zap.String("server_name", serverName))
	}

	return fmt.Sprintf("%s %s", fresh.TokenType, fresh.AccessToken), nil
}

// isExpiring 检查 token 是否即将过期
// 一比一复刻 DeerFlow 的 _is_expiring
func (m *OAuthTokenManager) isExpiring(token *OAuthToken, oauth *config.OAuthConfig) bool {
	now := time.Now().UTC()
	refreshSkew := oauth.RefreshSkewSeconds
	if refreshSkew < 0 {
		refreshSkew = 0
	}
	return token.ExpiresAt.Before(now.Add(time.Duration(refreshSkew) * time.Second))
}

// fetchToken 获取新 token
// 一比一复刻 DeerFlow 的 _fetch_token
func (m *OAuthTokenManager) fetchToken(oauth *config.OAuthConfig) (*OAuthToken, error) {
	data := url.Values{}
	data.Set("grant_type", oauth.GrantType)

	// 添加额外参数
	for k, v := range oauth.ExtraTokenParams {
		data.Set(k, v)
	}

	if oauth.Scopes != "" {
		data.Set("scope", oauth.Scopes)
	}
	if oauth.Audience != "" {
		data.Set("audience", oauth.Audience)
	}

	switch oauth.GrantType {
	case "client_credentials":
		if oauth.ClientID == "" || oauth.ClientSecret == "" {
			return nil, fmt.Errorf("OAuth client_credentials requires client_id and client_secret")
		}
		data.Set("client_id", oauth.ClientID)
		data.Set("client_secret", oauth.ClientSecret)

	case "refresh_token":
		if oauth.RefreshToken == "" {
			return nil, fmt.Errorf("OAuth refresh_token grant requires refresh_token")
		}
		data.Set("refresh_token", oauth.RefreshToken)
		if oauth.ClientID != "" {
			data.Set("client_id", oauth.ClientID)
		}
		if oauth.ClientSecret != "" {
			data.Set("client_secret", oauth.ClientSecret)
		}

	default:
		return nil, fmt.Errorf("unsupported OAuth grant type: %s", oauth.GrantType)
	}

	// 创建 HTTP 客户端
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// 发送请求
	resp, err := client.PostForm(oauth.TokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("OAuth token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OAuth token request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var payload map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to parse OAuth token response: %w", err)
	}

	// 提取 token 字段
	tokenField := oauth.TokenField
	if tokenField == "" {
		tokenField = "access_token"
	}
	accessToken, ok := payload[tokenField].(string)
	if !ok || accessToken == "" {
		return nil, fmt.Errorf("OAuth token response missing '%s'", tokenField)
	}

	// 提取 token type
	tokenTypeField := oauth.TokenTypeField
	if tokenTypeField == "" {
		tokenTypeField = "token_type"
	}
	defaultTokenType := oauth.DefaultTokenType
	if defaultTokenType == "" {
		defaultTokenType = "Bearer"
	}
	tokenType, _ := payload[tokenTypeField].(string)
	if tokenType == "" {
		tokenType = defaultTokenType
	}

	// 提取 expires_in
	expiresInField := oauth.ExpiresInField
	if expiresInField == "" {
		expiresInField = "expires_in"
	}
	expiresInRaw := payload[expiresInField]
	expiresIn := 3600 // 默认 1 小时
	switch v := expiresInRaw.(type) {
	case float64:
		expiresIn = int(v)
	case int:
		expiresIn = v
	}
	if expiresIn < 1 {
		expiresIn = 1
	}

	expiresAt := time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)

	return &OAuthToken{
		AccessToken: accessToken,
		TokenType:   tokenType,
		ExpiresAt:   expiresAt,
	}, nil
}

// GetInitialOAuthHeaders 获取初始 OAuth Authorization headers
// 一比一复刻 DeerFlow 的 get_initial_oauth_headers
func GetInitialOAuthHeaders(extensionsConfig *config.ExtensionsConfig, logger *zap.Logger) (map[string]string, error) {
	tokenManager := NewOAuthTokenManager(extensionsConfig, logger)
	if !tokenManager.HasOAuthServers() {
		return map[string]string{}, nil
	}

	headers := make(map[string]string)
	for _, serverName := range tokenManager.OAuthServerNames() {
		header, err := tokenManager.GetAuthorizationHeader(serverName)
		if err == nil && header != "" {
			headers[serverName] = header
		}
	}

	return headers, nil
}
