package middlewares

import (
	"net/http"
	"strings"
)

// ApplyCorsHeaders 应用CORS头，公开函数可被其他包直接使用
func ApplyCorsHeaders(w http.ResponseWriter, r *http.Request, allowOrigins string) bool {
	// 设置CORS头
	if allowOrigins == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else if origin := r.Header.Get("Origin"); origin != "" {
		// 检查请求的Origin是否在允许列表中
		origins := strings.Split(allowOrigins, ",")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
	}
	// 设置其他CORS头
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	// 处理OPTIONS预检请求
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}

	return false
}

// CorsMiddleware 创建CORS中间件
// 此函数为标准的http中间件格式，可用于http.Handler链
func CorsMiddleware(allowOrigins string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 应用CORS头，如果是OPTIONS请求则直接返回
		if ApplyCorsHeaders(w, r, allowOrigins) {
			return
		}

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// CreateCorsHandlerFunc 创建一个http.HandlerFunc形式的CORS处理函数
// 此函数可以直接用于http.Server的Handler字段
func CreateCorsHandlerFunc(allowOrigins string, defaultHandler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 应用CORS头，如果是OPTIONS请求则直接返回
		if ApplyCorsHeaders(w, r, allowOrigins) {
			return
		}
		// 继续处理请求
		defaultHandler.ServeHTTP(w, r)
	}
}
