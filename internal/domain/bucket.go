package domain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type TokenBucket struct {
	mu       sync.Mutex
	stopChan chan struct{}

	capacity  int64
	fillRate  time.Duration // Скорость пополнения (например, 1 токен/сек)
	tokens    int64         // Текущее количество токенов
	ticker    *time.Ticker
	lastUsing time.Time
}

type BucketConfig struct {
    Capacity int64         `json:"capacity"`
    FillRate time.Duration `json:"fillrate"`
}

func (b *BucketConfig) UnmarshalJSON(data []byte) error {
    type alias struct {
        Capacity int64  `json:"capacity"`
        FillRate string `json:"fillrate"` // Парсим как строку
    }
    
    var temp alias
    if err := json.Unmarshal(data, &temp); err != nil {
        return err
    }
    
    duration, err := time.ParseDuration(temp.FillRate)
    if err != nil {
        return fmt.Errorf("invalid fillrate format: %w", err)
    }
    
    b.Capacity = temp.Capacity
    b.FillRate = duration
    
    return nil
}

func NewBucket(capacity int64, fillRate time.Duration) *TokenBucket {
	b := &TokenBucket{
		capacity: capacity,
		fillRate: fillRate,
		tokens:   capacity,
		ticker:   time.NewTicker(fillRate),
		stopChan: make(chan struct{}),
	}

	go b.refill()

	return b
}

func (b *TokenBucket) refill() {
	for {
		select {
		case <-b.ticker.C:
			b.mu.Lock()
			if b.tokens < b.capacity {
				b.tokens++
			}
			b.mu.Unlock()
		case <-b.stopChan:
			b.ticker.Stop()
			return
		}
	}
}

func (b *TokenBucket) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastUsing = time.Now()

	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

func (b *TokenBucket) Stop() {
	close(b.stopChan)
}

func (b *TokenBucket) LastUsing() time.Time {
	return b.lastUsing
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
