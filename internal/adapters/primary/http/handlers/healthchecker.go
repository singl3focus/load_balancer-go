package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/singl3focus/load_balancer-go/internal/domain"
)

type HealthCheckParams struct {
	Interval time.Duration
}

type HealthChecker struct {
	client   *http.Client
	callback HealthCheckCallback
	stopChan chan struct{}
	servers  []*domain.Server
	mu       sync.RWMutex

	Interval time.Duration
}

type HealthCheckCallback func(s *domain.Server, alive bool)

func NewHealthChecker(
	interval time.Duration,
	callback HealthCheckCallback,
	servers []*domain.Server,
) *HealthChecker {
	return &HealthChecker{
		client: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives: true,
			},
		},
		stopChan: make(chan struct{}),
		Interval: interval,
		callback: callback,
	}
}

func (hc *HealthChecker) StartHealthCheck() {
	ticker := time.NewTicker(hc.Interval)
	go func() {
		defer close(hc.stopChan)
		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopChan:
				ticker.Stop()
				return
			default:
				// т.к. при infinity loop CPU сильно нагружается, необходимо использовать sleep
				// (50 для быстрой реакции, 500 - для мин. нагрузки CPU)
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func (hc *HealthChecker) StopHealthCheck() {
	select {
	case hc.stopChan <- struct{}{}:
	default: // Если канал закрыт, не вызовет паники
	}
}

func (hc *HealthChecker) UpdateServers(servers []*domain.Server) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	temp := make([]*domain.Server, len(servers))
	copy(temp, servers)
	hc.servers = temp
}

func (hc *HealthChecker) checkAll() {
	hc.mu.Lock()
	servers := hc.servers
	hc.mu.Unlock()

	var wg sync.WaitGroup
	for _, s := range servers {
		wg.Add(1)
		go func(server *domain.Server) {
			defer wg.Done()
			resp, err := hc.client.Get(server.Url)
			alive := err == nil

			if resp != nil && resp.StatusCode >= 500 {
				alive = false
			}
			if resp != nil {
				defer resp.Body.Close()
			}

			hc.callback(server, alive)
		}(s)
	}
	wg.Wait()
}