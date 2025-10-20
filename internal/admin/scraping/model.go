package scraping

type Serie struct {
	ID            int    `json:"id"`
	Title         string `json:"title"`
	EpisodesCount int    `json:"episodesCount"`
	Label         string `json:"label"`
	FavoriteID    int    `json:"favoriteID"`
	Thumbnail     string `json:"thumbnail"`
}

type Episode struct {
	ID     int     `json:"id"`
	Number float64 `json:"number"`
	Sub    int     `json:"sub"`
}

type SerieDetail struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ReleaseDate   string    `json:"releaseDate"`
	Trailer       string    `json:"trailer"`
	Country       string    `json:"country"`
	Status        string    `json:"status"`
	Type          string    `json:"type"`
	NextEpDateID  int       `json:"nextEpDateID"`
	Episodes      []Episode `json:"episodes"`
	EpisodesCount int       `json:"episodesCount"`
	Label         *string   `json:"label"`
	FavoriteID    int       `json:"favoriteID"`
	Thumbnail     string    `json:"thumbnail"`
}

type SerieDeepDetail struct {
	ID            int           `json:"id"`
	Title         string        `json:"title"`
	Description   string        `json:"description"`
	ReleaseDate   string        `json:"releaseDate"`
	Trailer       string        `json:"trailer"`
	Country       string        `json:"country"`
	Status        string        `json:"status"`
	Type          string        `json:"type"`
	NextEpDateID  int           `json:"nextEpDateID"`
	Episodes      []EpisodeDeep `json:"episodes"`
	EpisodesCount int           `json:"episodesCount"`
	Label         *string       `json:"label"`
	FavoriteID    int           `json:"favoriteID"`
	Thumbnail     string        `json:"thumbnail"`
}

type EpisodeDeep struct {
	ID        int        `json:"id"`
	Number    float64    `json:"number"`
	Sub       int        `json:"sub"`
	Source    string     `json:"src"`
	Subtitles []Subtitle `json:"subtitles"`
}

type Subtitle struct {
	Src     string `json:"src"`
	Label   string `json:"label"`
	Lang    string `json:"land"`
	Default bool   `json:"default"`
}

type SeriesResponse struct {
	Series []Serie `json:"series"`
}

type SeriesDeepDetailsResponse struct {
	SeriesDeepDetails []SerieDeepDetail `json:"series_deep_details"`
}

type SeriesDetailsResponse struct {
	SeriesDetails []SerieDetail `json:"series_details"`
}
