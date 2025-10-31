package scraping

import "rerng_addicted_api/internal/admin/serie"

type SeriesResponse struct {
	Series []serie.Serie `json:"series"`
}

type SeriesDeepDetailsResponse struct {
	SeriesDeepDetails []serie.SerieDeepDetail `json:"series_deep_details"`
}

type SeriesDetailsResponse struct {
	SeriesDetails []serie.SerieDetail `json:"series_details"`
}

type EpisodesResponse struct {
	Episodes []serie.EpisodeDeep `json:"episodes"`
}
