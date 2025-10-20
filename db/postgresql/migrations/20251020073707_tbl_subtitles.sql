-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_subtitles (
    id BIGSERIAL PRIMARY KEY,
    episode_id BIGINT NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
    src VARCHAR(500) NOT NULL,
    label VARCHAR(50),
    lang VARCHAR(10),
    is_default BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP DEFAULT NOW(),
    created_by BIGINT,
    updated_at TIMESTAMP,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_subtitles_episode_id ON tbl_subtitles(episode_id);

-- +goose Down
DROP TABLE IF EXISTS tbl_subtitles;
