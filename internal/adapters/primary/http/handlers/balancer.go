package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"github.com/singl3focus/load_balancer-go/internal/config"
	"github.com/singl3focus/load_balancer-go/internal/domain"
	"github.com/singl3focus/load_balancer-go/pkg/logger"
)

const (
	MaxAttempts            = 3
	DefaultErrorStatusCode = http.StatusBadGateway
)

type OptionFunc func(*LoadBalancer)

func WithUsingHealthCheck() OptionFunc {
	return func(lb *LoadBalancer) {
		lb.UsingHealthCheck = true
	}
}

var _ http.Handler = &LoadBalancer{}

type LoadBalancer struct {
	mu            sync.RWMutex
	servers       []*domain.Server
	strategy      domain.Strategy
	healthChecker *HealthChecker
	logger        logger.Logger

	UsingHealthCheck bool
}

func NewLoadBalancer(
	logger logger.Logger,
	strategy domain.Strategy,
	servers []*domain.Server,
	hcParams HealthCheckParams,
	opts ...OptionFunc,
) *LoadBalancer {
	lb := &LoadBalancer{
		strategy: strategy,
		servers:  servers,
		logger:   logger,
		healthChecker: NewHealthChecker(
			hcParams.Interval,
			func(s *domain.Server, alive bool) {
				s.SetAlive(alive)
				logger.Info("Server health status changed", "(url)", s.Url, "(alive)", alive)
			},
			servers,
		),
	}

	for _, opt := range opts {
		opt(lb)
	}

	if lb.UsingHealthCheck {
		lb.healthChecker.StartHealthCheck()
	}

	return lb
}

// AddServer добавляет сервер в пул
func (lb *LoadBalancer) AddServer(server *domain.Server) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.servers = append(lb.servers, server)
	lb.strategy.UpdateServers(lb.servers)
	lb.healthChecker.UpdateServers(lb.servers)
}

// RemoveServer удаляет сервер из пула
func (lb *LoadBalancer) RemoveServer(id int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	for i, s := range lb.servers {
		if s.ID == id {
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
			lb.strategy.UpdateServers(lb.servers)
			lb.healthChecker.UpdateServers(lb.servers)
			return
		}
	}
}

// ServeHTTP  обработчик запросов
func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for attempts := 0; attempts < MaxAttempts; attempts++ {
		server, err := lb.strategy.NextServer()
		if err != nil {
			switch {
			case errors.Is(err, domain.ErrNoAvailableServers):
				lb.handleError(w, http.StatusServiceUnavailable, domain.ErrNoAvailableServers.Error())
				lb.logger.Error("No avaliable servers")
			default:
				lb.handleError(w, DefaultErrorStatusCode, "internal server error")
				lb.logger.Error("Undefined get server error", "(err)", err.Error())
			}

			return
		}

		if server.IsAvailable() {
			server.HandleRequest(w, r)
			return
		}

		lb.logger.Warn("Server unavailable",
			"(attempt)", attempts+1, "(server_id)", server.ID, "(url)", server.Url)
	}

	// [DEV] с точки зрения безопасности юзер не должен знать что он обращается к балансировщику и
	// подробное описание ошибки.(т.к. это может привести к уязвимостям)
	lb.handleError(w, http.StatusServiceUnavailable, "all servers failed after maximum attempts")
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (lb *LoadBalancer) handleError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := ErrorResponse{
		Code:    code,
		Message: msg,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		lb.logger.Error("Failed to encode error response", "error", err)
	}
}

func (lb *LoadBalancer) ApplyConfig(newConfig config.BalancerConfig, strategy domain.Strategy) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	oldServersLen := len(lb.servers)

	// Удаление отсутствующих серверов
	existingIDs := make(map[int]bool)
	for _, s := range newConfig.Servers {
		existingIDs[s.ID] = true
	}

	for i := 0; i < len(lb.servers); i++ {
		if !existingIDs[lb.servers[i].ID] {
			lb.servers = append(lb.servers[:i], lb.servers[i+1:]...)
			i--
		}
	}

	// Добавление/обновление серверов
	for _, s := range newConfig.Servers { // TODO: O(n^2) - что если серверов будет много?
		found := false
		for _, existing := range lb.servers {
			if existing.ID == s.ID {
				existing.SetURL(s.URL)
				found = true
				break
			}
		}
		if !found {
			lb.servers = append(lb.servers, domain.NewServer(s.ID, s.URL))
		}
	}

	// Обновление данных в стратегии
	strategy.UpdateServers(lb.servers)
	lb.strategy = strategy

	lb.logger.Info("Balancer config updated",
		"(added)", len(lb.servers)-oldServersLen,
		"(removed)", oldServersLen-len(lb.servers),
		"(total)", len(lb.servers),
	)
}

// GracefulShutdown корректно завершает работу балансировщика
func (lb *LoadBalancer) GracefulShutdown(ctx context.Context) error {    
    if lb.UsingHealthCheck {
        lb.healthChecker.StopHealthCheck()
        lb.logger.Debug("Health checks stopped")
    }

	if err := lb.strategy.Shutdown(ctx); err != nil {
		lb.logger.Error("Strategy shutdown failed", "error", err)
	}
    
    // ожидание завершения серверов
    var wg sync.WaitGroup
    for _, srv := range lb.servers {
        wg.Add(1)
        go func(s *domain.Server) {
            defer wg.Done()
            s.CloseConnections()
        }(srv)
    }

    shutdownCompleted := make(chan struct{})
    go func() {
        wg.Wait()
        close(shutdownCompleted)
    }()

    select {
    case <-shutdownCompleted:
    case <-ctx.Done():
        lb.logger.Warn("Shutdown timed out, force closing")
        return ctx.Err()
    }
    
    return nil
}