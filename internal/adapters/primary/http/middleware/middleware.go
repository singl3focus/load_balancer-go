package middleware

import (
	"net/http"
	"strings"

	"github.com/singl3focus/load_balancer-go/internal/service"
)

type Middleware struct {
    limiter *service.Limiter
}

func NewRateLimitterMiddleware(limiter *service.Limiter) *Middleware {
    return &Middleware{limiter: limiter}
}

func (m *Middleware) Handle(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        key := getClientKey(r)

        if !m.limiter.Allow(key) {
            http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func getClientKey(r *http.Request) string {
    // Приоритет: API Key -> IP. Т.к. API Key более уникальный объект, чем IP.
    if key := r.Header.Get("X-API-Key"); key != "" {
        return key
    }
    return strings.Split(r.RemoteAddr, ":")[0]
}