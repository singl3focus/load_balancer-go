package postgres

import (
	"fmt"

	"github.com/singl3focus/load_balancer-go/internal/adapters/secondary/storage"
	"github.com/singl3focus/load_balancer-go/internal/domain"
)

// GetConfig возвращает конфигурацию бакета по ключу
func (d *Database) GetConfig(key string) (domain.BucketConfig, error) {
    const op = "storage.Postgres.GetConfig"
    
    query := fmt.Sprintf(
        `SELECT capacity, fill_rate FROM %s WHERE key = $1`,
        rateLimitConfigTable,
    )

    var config domain.BucketConfig
    err := d.db.QueryRow(query, key).Scan(&config.Capacity, &config.FillRate)
    if err != nil {
        return domain.BucketConfig{}, fmt.Errorf("%s: %w", op, err)
    }
    
    return config, nil
}

// StoreConfig сохраняет или обновляет конфигурацию бакета
func (d *Database) StoreConfig(key string, cfg domain.BucketConfig) error {
    const op = "storage.Postgres.StoreConfig"
    
    query := fmt.Sprintf(`
        INSERT INTO %s (key, capacity, fill_rate, updated_at)
        VALUES ($1, $2, $3, NOW())
        ON CONFLICT (key) 
        DO UPDATE SET
            capacity = EXCLUDED.capacity,
            fill_rate = EXCLUDED.fill_rate,
            updated_at = EXCLUDED.updated_at
    `, rateLimitConfigTable)

    _, err := d.db.Exec(query,
		key, cfg.Capacity, cfg.FillRate.Microseconds(), // Храним в микросекундах
	)
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }
    
    return nil
}

// UpdateConfig обновляет конфиг, если действия не произошло, то возвращается ошибка  
func (d *Database) UpdateConfig(key string, cfg domain.BucketConfig) error {
    const op = "storage.Postgres.UpdateConfig"

    query := fmt.Sprintf(`
        UPDATE %s 
        SET capacity = $1,
            fill_rate = $2,
            updated_at = NOW()
        WHERE key = $3
    `, rateLimitConfigTable)

    result, err := d.db.Exec(
        query,
        cfg.Capacity,
        cfg.FillRate.Microseconds(),
        key,
    )

    if rows, _ := result.RowsAffected(); rows == 0 {
        return storage.ErrEmptyAction
    }

    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    return nil
}

// DeleteConfig удаляет конфиг, если действия не произошло, то возвращается ошибка  
func (d *Database) DeleteConfig(key string) error {
    const op = "storage.Postgres.DeleteConfig"

    query := fmt.Sprintf(`DELETE FROM %s WHERE key = $1`, rateLimitConfigTable)

    result, err := d.db.Exec(query, key)
    if err != nil {
        return fmt.Errorf("%s: %w", op, err)
    }

    if rows, _ := result.RowsAffected(); rows == 0 {
        return storage.ErrEmptyAction
    }

    return nil
}