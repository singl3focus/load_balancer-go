package domain

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Server struct {
	proxy             *httputil.ReverseProxy
	mu                sync.RWMutex 
	
	ID                int          // Уникальный идентификатор сервера
	Url               string       // URL бэкенда
	alive             bool         // Флаг доступности сервера
	activeConnections int32        // Текущее количество активных соединений
}

func NewServer(id int, serverUrl string) *Server {
	parsedUrl, err := url.Parse(serverUrl)
	if err != nil {
		panic(err)
	}

	return &Server{
		ID:    id,
		Url:   serverUrl,
		proxy: httputil.NewSingleHostReverseProxy(parsedUrl),
		alive: true,
	}
}

// Обработка запроса с отслеживанием активных соединений
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&s.activeConnections, 1)
	defer atomic.AddInt32(&s.activeConnections, -1)

	s.proxy.ServeHTTP(w, r)
}

// CheckHealth проверка health check для сервера
func (s *Server) CheckHealth() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

// SetAlive установка состояния сервера
func (s *Server) SetAlive(alive bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alive = alive
}

// SetURL
func (s *Server) SetURL(serverUrl string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    parsed, _ := url.Parse(serverUrl)
    s.proxy = httputil.NewSingleHostReverseProxy(parsed)
    s.Url = serverUrl
}

// ActiveConnections получение текущего количества активных соединений
func (s *Server) ActiveConnections() int {
	return int(atomic.LoadInt32(&s.activeConnections))
}

// IsAvailable демонстрирует доступность сервера
func (s *Server) IsAvailable() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.alive
}

func (s *Server) CloseConnections() {
	s.SetAlive(false)

	if tr, ok := s.proxy.Transport.(interface { CloseIdleConnections() }); ok {
        tr.CloseIdleConnections()
    }
}