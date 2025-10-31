package serie

import (
	"fmt"
	custom_log "rerng_addicted_api/pkg/logs"
	share "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/responses"
	"strings"

	"github.com/fatih/color"
	"github.com/jmoiron/sqlx"
)

type SerieRepo interface {
	Create(serie_detail SerieDeepDetail) (*SeriesDeepDetailsResponse, *responses.ErrorResponse)
	InsertEpisode(tx *sqlx.Tx, serie_id int, ep EpisodeDeep) error
	InsertSubtitle(tx *sqlx.Tx, episode_id int, sub Subtitle) error
}

type SerieRepoImpl struct {
	DBPool      *sqlx.DB
	UserContext *share.UserContext
}

func NewSerieRepoImpl(db_pool *sqlx.DB, userCtx *share.UserContext) *SerieRepoImpl {
	return &SerieRepoImpl{
		DBPool:      db_pool,
		UserContext: userCtx,
	}
}

func (sc *SerieRepoImpl) Create(serie_detail SerieDeepDetail) (*SeriesDeepDetailsResponse, *responses.ErrorResponse) {
	// begin transaction
	tx, err := sc.DBPool.Beginx()
	if err != nil {
		color.Red("❌ Failed to start transaction: %v", err)
		custom_log.NewCustomLog("insert_serie_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("insert_serie_failed", fmt.Errorf("technical_error"))
	}

	color.Yellow("\n💾 Upserting series and related data...")
	// upsert data
	_, err = tx.NamedExec(`
		INSERT INTO tbl_series 
			(id, title, description, release_date, trailer, country, status, type, next_ep_date_id,
			episodes_count, label, favorite_id, thumbnail)
		VALUES 
			(:id, :title, :description, :release_date, :trailer, :country, :status, :type, :next_ep_date_id,
			:episodes_count, :label, :favorite_id, :thumbnail)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			release_date = EXCLUDED.release_date,
			trailer = EXCLUDED.trailer,
			country = EXCLUDED.country,
			status = EXCLUDED.status,
			type = EXCLUDED.type,
			next_ep_date_id = EXCLUDED.next_ep_date_id,
			episodes_count = EXCLUDED.episodes_count,
			label = EXCLUDED.label,
			favorite_id = EXCLUDED.favorite_id,
			thumbnail = EXCLUDED.thumbnail
	`, serie_detail)
	if err != nil {
		color.Red("❌ Failed to upsert series %d: %v", serie_detail.ID, err)
		tx.Rollback()
		custom_log.NewCustomLog("insert_serie_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("insert_serie_failed", fmt.Errorf("database_error"))
	}

	// insert episodes
	for _, ep := range serie_detail.Episodes {
		if err := sc.InsertEpisode(tx, serie_detail.ID, ep); err != nil {
			color.Red("❌ %v", err)
			tx.Rollback()
			custom_log.NewCustomLog("insert_episode_failed", err.Error(), "error")
			err_msg := &responses.ErrorResponse{}
			return nil, err_msg.NewErrorResponse("insert_episode_failed", fmt.Errorf("database_error"))
		}
	}

	// commit transaction
	if err := tx.Commit(); err != nil {
		color.Red("❌ Transaction commit failed: %v", err)
		custom_log.NewCustomLog("insert_serie_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("insert_serie_failed", fmt.Errorf("database_error"))
	}

	color.Green("\n✅ Upserted series %d successfully", serie_detail.ID)
	return &SeriesDeepDetailsResponse{
		SeriesDeepDetails: []SerieDeepDetail{serie_detail},
	}, nil
}

func (sc *SerieRepoImpl) InsertEpisode(execer sqlx.Ext, serie_id int, ep EpisodeDeep) error {
	// determine status_id based on source URL
	status_id := 1
	if !strings.Contains(ep.Source, ".m3u8") && !strings.Contains(ep.Source, ".mp4") {
		status_id = 2
	}
	_, err := sqlx.NamedExec(execer, `
		INSERT INTO tbl_episodes (id, series_id, number, sub, src, status_id)
		VALUES (:id, :series_id, :number, :sub, :src, :status_id)
		ON CONFLICT (id) DO UPDATE SET
			series_id = EXCLUDED.series_id,
			number = EXCLUDED.number,
			sub = EXCLUDED.sub,
			src = EXCLUDED.src,
			status_id = EXCLUDED.status_id
	`, map[string]interface{}{
		"id":        ep.ID,
		"series_id": serie_id,
		"number":    ep.Number,
		"sub":       ep.Sub,
		"src":       ep.Source,
		"status_id": status_id,
	})
	if err != nil {
		return fmt.Errorf("failed upserting episode %d: %w", ep.ID, err)
	}

	for _, sub := range ep.Subtitles {
		if err := sc.InsertSubtitle(execer, ep.ID, sub); err != nil {
			return err
		}
	}
	return nil
}

func (sc *SerieRepoImpl) InsertSubtitle(execer sqlx.Ext, episode_id int, sub Subtitle) error {
	_, err := sqlx.NamedExec(execer, `
		INSERT INTO tbl_subtitles (episode_id, src, label, lang, is_default)
		VALUES (:episode_id, :src, :label, :lang, :is_default)
		ON CONFLICT (episode_id, lang) DO UPDATE SET
			src = EXCLUDED.src,
			label = EXCLUDED.label,
			is_default = EXCLUDED.is_default
	`, map[string]interface{}{
		"episode_id": episode_id,
		"src":        sub.Src,
		"label":      sub.Label,
		"lang":       sub.Lang,
		"is_default": sub.Default,
	})
	if err != nil {
		return fmt.Errorf("failed upserting subtitle for episode %d: %w", episode_id, err)
	}
	return nil
}
