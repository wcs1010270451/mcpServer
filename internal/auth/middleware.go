package auth

import (
	"log"
	"net/http"
	"strings"

	"McpServer/internal/config"
)

// AuthMiddleware 认证中间件
type AuthMiddleware struct {
	config *config.AuthConfig
}

// NewAuthMiddleware 创建新的认证中间件
func NewAuthMiddleware(authConfig *config.AuthConfig) *AuthMiddleware {
	return &AuthMiddleware{
		config: authConfig,
	}
}

// ValidateAPIKey 验证API密钥
func (am *AuthMiddleware) ValidateAPIKey(apiKey string) bool {
	if !am.config.Enabled {
		return true // 认证未启用，直接通过
	}

	if apiKey == "" {
		return false
	}

	// 检查API密钥是否在允许列表中
	for _, validKey := range am.config.APIKeys {
		if apiKey == validKey {
			return true
		}
	}

	return false
}

// Middleware HTTP中间件函数
func (am *AuthMiddleware) Middleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果认证未启用，直接通过
		if !am.config.Enabled {
			next(w, r)
			return
		}

		// 获取API密钥
		headerName := am.config.HeaderName
		if headerName == "" {
			headerName = "X-API-Key" // 默认头名称
		}

		apiKey := r.Header.Get(headerName)

		// 验证API密钥
		if !am.ValidateAPIKey(apiKey) {
			log.Printf("Authentication failed for request %s %s from %s",
				r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, "Unauthorized: Invalid API Key", http.StatusUnauthorized)
			return
		}

		log.Printf("Authentication successful for request %s %s", r.Method, r.URL.Path)
		next(w, r)
	}
}

// ExtractAPIKey 从请求中提取API密钥
func (am *AuthMiddleware) ExtractAPIKey(r *http.Request) string {
	headerName := am.config.HeaderName
	if headerName == "" {
		headerName = "X-API-Key"
	}

	// 优先从指定头获取
	apiKey := r.Header.Get(headerName)
	if apiKey != "" {
		return apiKey
	}

	// 尝试从Authorization头获取Bearer token
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 尝试从查询参数获取
	return r.URL.Query().Get("api_key")
}

// IsEnabled 检查认证是否启用
func (am *AuthMiddleware) IsEnabled() bool {
	return am.config.Enabled
}

// GetHeaderName 获取API密钥头名称
func (am *AuthMiddleware) GetHeaderName() string {
	if am.config.HeaderName == "" {
		return "X-API-Key"
	}
	return am.config.HeaderName
}
