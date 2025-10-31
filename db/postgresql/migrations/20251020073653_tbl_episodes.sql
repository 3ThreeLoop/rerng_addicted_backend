-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_episodes (
    id BIGINT PRIMARY KEY,
    series_id BIGINT NOT NULL,
    number NUMERIC(5,2) NOT NULL,
    sub INT DEFAULT 0,
    src VARCHAR(500) NOT NULL,

    status_id INTEGER NOT NULL DEFAULT 1,
    "order" INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by BIGINT,
    updated_at TIMESTAMP,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_episodes_series_id ON tbl_episodes(series_id);

-- +goose Down
DROP TABLE IF EXISTS tb_episodes;
