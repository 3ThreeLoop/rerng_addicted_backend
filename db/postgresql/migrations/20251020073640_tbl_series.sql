-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_series (
    id BIGINT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    release_date TIMESTAMP,
    trailer VARCHAR(500),
    country VARCHAR(100),
    status VARCHAR(50),
    type VARCHAR(50),
    next_ep_date_id BIGINT DEFAULT 0,
    episodes_count INT DEFAULT 0,
    label VARCHAR(100),
    favorite_id BIGINT DEFAULT 0,
    thumbnail VARCHAR(500),

    status_id INTEGER NOT NULL DEFAULT 1,
    "order" INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by BIGINT,
    updated_at TIMESTAMP,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

-- +goose Down
DROP TABLE IF EXISTS tbl_series;
