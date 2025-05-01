package service

import (
	"context"
	"sync/atomic"

	"github.com/singl3focus/load_balancer-go/internal/domain"
)

func GetStrategy(strategyName string, servers []*domain.Server) domain.Strategy {
	if servers == nil {
		servers = make([]*domain.Server, 0)
	}

	var strategy domain.Strategy
	switch strategyName {
	case "round_robin":
		strategy = NewRoundRobinStrategy(servers)
	default: // "least_connections" or other
		strategy = NewLeastConnectionsStrategy(servers)
	}

	return strategy
}


var _ domain.Strategy = &RoundRobinStrategy{}

// RoundRobinStrategy
type RoundRobinStrategy struct {
	servers atomic.Value // []*Server
	index   uint32
}

func NewRoundRobinStrategy(servers []*domain.Server) *RoundRobinStrategy {
    s := &RoundRobinStrategy{}
    s.servers.Store(servers)
    return s
}

// NextServer возвращает первый доступный сервер, начиная с последующего по индексу.
// Если все сервера недоступны, возвращается ошибка ErrNoAvailableServers.
func (r *RoundRobinStrategy) NextServer() (*domain.Server, error) {
	servers := r.servers.Load().([]*domain.Server)
	if len(servers) == 0 {
		return nil, domain.ErrNoAvailableServers
	}

	for i := 0; i < len(servers); i++ {
		idx := atomic.AddUint32(&r.index, 1) % uint32(len(servers))

		server := servers[idx]
		if server.IsAvailable() {
			return server, nil
		}
	}
	return nil, domain.ErrNoAvailableServers
}

func (r *RoundRobinStrategy) UpdateServers(servers []*domain.Server) {
	r.servers.Store(servers)
	atomic.StoreUint32(&r.index, 0)
}

func (r *RoundRobinStrategy) Shutdown(ctx context.Context) error {
	r.servers.Store(make([]*domain.Server, 0))
    return nil
}

var _ domain.Strategy = &LeastConnectionsStrategy{}

// LeastConnectionsStrategy
type LeastConnectionsStrategy struct {
	servers atomic.Value // []*Server
}

func NewLeastConnectionsStrategy(servers []*domain.Server) *LeastConnectionsStrategy {
    s := &LeastConnectionsStrategy{}
    s.servers.Store(servers)
    return s
}

func (r *LeastConnectionsStrategy) UpdateServers(servers []*domain.Server) {
	r.servers.Store(servers)
}

func (l *LeastConnectionsStrategy) NextServer() (*domain.Server, error) {
	servers := l.servers.Load().([]*domain.Server)
	if len(servers) == 0 {
		return nil, domain.ErrNoAvailableServers
	}

	var best *domain.Server
	for _, s := range servers {
		if !s.IsAvailable() {
			continue
		}
		if best == nil || s.ActiveConnections() < best.ActiveConnections() {
			best = s
		}
	}

	return best, nil
}

func (l *LeastConnectionsStrategy) Shutdown(ctx context.Context) error {
    l.servers.Store(make([]*domain.Server, 0))
    return nil
}