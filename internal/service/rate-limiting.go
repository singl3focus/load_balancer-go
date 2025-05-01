package service

import (
	"sync"
	"time"

	"github.com/singl3focus/load_balancer-go/internal/domain"
)

type Limiter struct {
    buckets     *sync.Map
    defaultCfg  domain.BucketConfig
    storage     domain.Storage
}

func NewLimiter(defaultCfg domain.BucketConfig, storage domain.Storage) *Limiter {
    return &Limiter{
        buckets:    new(sync.Map),
        defaultCfg: defaultCfg,
        storage:    storage,
    }
}

func (l *Limiter) Allow(key string) bool {
    cfg, err := l.storage.GetConfig(key)
    if err != nil {
        cfg = l.defaultCfg
    }

    actual, _ := l.buckets.LoadOrStore(key, domain.NewBucket(cfg.Capacity, cfg.FillRate))
    bucket := actual.(*domain.TokenBucket)
    return bucket.Allow()
}

func (l *Limiter) UpdateConfig(key string, cfg domain.BucketConfig) {
    _ = l.storage.StoreConfig(key, cfg)
    l.buckets.Delete(key) // удаляем существующий bucket
}

func (l *Limiter) Cleanup(interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        for range ticker.C {
            l.buckets.Range(func(key, value interface{}) bool {
                bucket := value.(*domain.TokenBucket)
                if time.Since(bucket.LastUsing()) > 1 * time.Hour {
                    l.buckets.Delete(key)
                }
                return true
            })
        }
    }()
}