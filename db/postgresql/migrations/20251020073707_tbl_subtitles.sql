-- +goose Up
CREATE TABLE IF NOT EXISTS tbl_subtitles (
    id BIGSERIAL PRIMARY KEY,
    episode_id BIGINT NOT NULL,
    src VARCHAR(500) NOT NULL,
    label VARCHAR(50),
    lang VARCHAR(10),
    is_default BOOLEAN DEFAULT FALSE,

    status_id INTEGER NOT NULL DEFAULT 1,
    "order" INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT NOW(),
    created_by BIGINT,
    updated_at TIMESTAMP,
    updated_by BIGINT,
    deleted_at TIMESTAMP,
    deleted_by BIGINT
);

CREATE INDEX IF NOT EXISTS idx_subtitles_episode_id ON tbl_subtitles(episode_id);

ALTER TABLE tbl_subtitles
ADD CONSTRAINT uq_episode_lang UNIQUE (episode_id, lang);

-- +goose Down
DROP TABLE IF EXISTS tbl_subtitles;
