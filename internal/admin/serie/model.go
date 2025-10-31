package serie

import (
	"fmt"
	custom_log "rerng_addicted_api/pkg/logs"
	"rerng_addicted_api/pkg/utils"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Serie struct {
	ID            int    `db:"id" json:"id"`
	Title         string `db:"title" json:"title"`
	EpisodesCount int    `db:"episodes_count" json:"episodes_count"`
	Label         string `db:"label" json:"label"`
	FavoriteID    int    `db:"favorite_id" json:"favorite_id"`
	Thumbnail     string `db:"thumbnail" json:"thumbnail"`
}

type Episode struct {
	ID     int     `db:"id" json:"id"`
	Number float64 `db:"number" json:"number"`
	Sub    int     `db:"sub" json:"sub"`
}

type SerieDetail struct {
	ID            int       `db:"id" json:"id"`
	Title         string    `db:"title" json:"title"`
	Description   string    `db:"description" json:"description"`
	ReleaseDate   string    `db:"release_date" json:"release_date"`
	Trailer       string    `db:"trailer" json:"trailer"`
	Country       string    `db:"country" json:"country"`
	Status        string    `db:"status" json:"status"`
	Type          string    `db:"type" json:"type"`
	NextEpDateID  int       `db:"next_ep_date_id" json:"next_ep_date_id"`
	Episodes      []Episode `json:"episodes"`
	EpisodesCount int       `db:"episodes_count" json:"episodes_count"`
	Label         *string   `db:"label" json:"label"`
	FavoriteID    int       `db:"favorite_id" json:"favorite_id"`
	Thumbnail     string    `db:"thumbnail" json:"thumbnail"`
}

type SerieDeepDetail struct {
	ID            int           `db:"id" json:"id"`
	Title         string        `db:"title" json:"title"`
	Description   string        `db:"description" json:"description"`
	ReleaseDate   *time.Time    `db:"release_date" json:"release_date"`
	Trailer       string        `db:"trailer" json:"trailer"`
	Country       string        `db:"country" json:"country"`
	Status        string        `db:"status" json:"status"`
	Type          string        `db:"type" json:"type"`
	NextEpDateID  int           `db:"next_ep_date_id" json:"next_ep_date_id"`
	Episodes      []EpisodeDeep `json:"episodes"`
	EpisodesCount int           `db:"episodes_count" json:"episodes_count"`
	Label         *string       `db:"label" json:"label"`
	FavoriteID    int           `db:"favorite_id" json:"favorite_id"`
	Thumbnail     string        `db:"thumbnail" json:"thumbnail"`
}

type EpisodeDeep struct {
	ID        int        `db:"id" json:"id"`
	SeriesID  int        `db:"series_id" json:"series_id"`
	Number    float64    `db:"number" json:"number"`
	Sub       int        `db:"sub" json:"sub"`
	Source    string     `db:"src" json:"src"`
	Subtitles []Subtitle `json:"subtitles"`
}

type Subtitle struct {
	Src     string `db:"src" json:"src"`
	Label   string `db:"label" json:"label"`
	Lang    string `db:"lang" json:"lang"`
	Default bool   `db:"is_default" json:"is_default"`
}

type SeriesDeepDetailsResponse struct {
	SeriesDeepDetails []SerieDeepDetail `json:"series_deep_details"`
}

type SerieJSON struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	EpisodesCount int    `json:"episodesCount"`
	Label         string `json:"label"`
	FavoriteID    int    `json:"favoriteId"`
	Thumbnail     string `json:"thumbnail"`
}

type EpisodeJSON struct {
	ID     int     `json:"id"`
	Number float64 `json:"number"`
	Sub    int     `json:"sub"`
}

type SerieDetailJSON struct {
	ID            int           `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ReleaseDate   string        `json:"releaseDate"`
	Trailer       string        `json:"trailer"`
	Country       string        `json:"country"`
	Status        string        `json:"status"`
	Type          string        `json:"type"`
	NextEpDateID  int           `json:"nextEpDateId"`
	Episodes      []EpisodeJSON `json:"episodes"`
	EpisodesCount int           `json:"episodesCount"`
	Label         *string       `json:"label"`
	FavoriteID    int           `json:"favoriteId"`
	Thumbnail     string        `json:"thumbnail"`
}

type SerieDeepDetailJSON struct {
	ID            int               `json:"id"`
	Title         string            `json:"title"`
	Description   string            `json:"description"`
	ReleaseDate   string            `json:"releaseDate"`
	Trailer       string            `json:"trailer"`
	Country       string            `json:"country"`
	Status        string            `json:"status"`
	Type          string            `json:"type"`
	NextEpDateID  int               `json:"nextEpDateId"`
	Episodes      []EpisodeDeepJSON `json:"episodes"`
	EpisodesCount int               `json:"episodesCount"`
	Label         *string           `json:"label"`
	FavoriteID    int               `json:"favoriteId"`
	Thumbnail     string            `json:"thumbnail"`
}

type EpisodeDeepJSON struct {
	ID        int            `json:"id"`
	SeriesID  int            `json:"seriesId"`
	Number    float64        `json:"number"`
	Sub       int            `json:"sub"`
	Source    string         `json:"src"`
	Subtitles []SubtitleJSON `json:"subtitles"`
}

type SubtitleJSON struct {
	Src     string `json:"src"`
	Label   string `json:"label"`
	Lang    string `json:"land"`
	Default bool   `json:"Default"`
}

type NewSerieRequest struct {
	ID            int       `db:"id" json:"id" validate:"required"`
	Title         string    `db:"title" json:"title" validate:"required,min=1,max=255"`
	Description   string    `db:"description" json:"description" validate:"required,min=5"`
	ReleaseDate   string    `db:"release_date" json:"release_date" validate:"omitempty,datetime=2006-01-02"`
	Trailer       string    `db:"trailer" json:"trailer" validate:"omitempty,url"`
	Country       string    `db:"country" json:"country" validate:"required,min=2,max=100"`
	Status        string    `db:"status" json:"status" validate:"required,oneof=ongoing completed canceled upcoming"`
	Type          string    `db:"type" json:"type" validate:"required,oneof=movie series ova special"`
	Episodes      []Episode `json:"episodes" validate:"omitempty,dive"`
	EpisodesCount int       `db:"episodes_count" json:"episodes_count" validate:"required,gte=0"`
	Label         *string   `db:"label" json:"label" validate:"omitempty,max=100"`
	FavoriteID    int       `db:"favorite_id" json:"favorite_id" validate:"gte=0"`
	Thumbnail     string    `db:"thumbnail" json:"thumbnail" validate:"required,url"`
}

func (s *NewSerieRequest) bind(c *fiber.Ctx, v *utils.Validator) error {
	if err := c.BodyParser(s); err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		return fmt.Errorf("%s", utils.Translate("invalid_body", nil, c))
	}

	if err := v.Validate(s, c); err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		return err
	}

	return nil
}
