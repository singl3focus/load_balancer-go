-- +goose Up

CREATE TABLE rate_limit_configs (
    key        TEXT PRIMARY KEY,
    capacity   BIGINT NOT NULL,
    fill_rate  BIGINT NOT NULL,  -- В микросекундах
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
