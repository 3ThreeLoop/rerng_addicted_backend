package scraping

import (
	"fmt"
	"rerng_addicted_api/internal/admin/serie"
	custom_log "rerng_addicted_api/pkg/logs"
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/responses"

	"github.com/jmoiron/sqlx"
)

type ScrapingServiceCreator interface {
	Search(keyword string) (*SeriesResponse, *responses.ErrorResponse)
	ViewDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse)
	GetDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse)
	GetEpisodes(key int, ep_num int) (*EpisodesResponse, *responses.ErrorResponse)
	Seed()
}

type ScrapingService struct {
	DBPool       *sqlx.DB
	ScrapingRepo *ScrapingRepoImpl
	UserContext  *types.UserContext
}

func NewScrapingService(db_pool *sqlx.DB, user_context *types.UserContext) *ScrapingService {
	return &ScrapingService{
		DBPool:       db_pool,
		ScrapingRepo: NewScrapingRepoImpl(db_pool, user_context),
	}
}

func (sc *ScrapingService) Search(keyword string) (*SeriesResponse, *responses.ErrorResponse) {
	return sc.ScrapingRepo.Search(keyword)
}

func (sc *ScrapingService) ViewDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse) {
	return sc.ScrapingRepo.ViewDetail(key)
}

func (sc *ScrapingService) GetDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse) {
	return sc.ScrapingRepo.GetDetail(key)
}

func (sc *ScrapingService) GetEpisodes(key int, ep_num int) (*EpisodesResponse, *responses.ErrorResponse) {
	resp, err := sc.ScrapingRepo.GetEpisodes(key, ep_num)

	if err == nil && len(resp.Episodes) > 0 {
		serie_repo := serie.NewSerieRepoImpl(sc.DBPool, sc.UserContext)
		if insert_err := serie_repo.InsertEpisode(sc.DBPool, key, resp.Episodes[0]); insert_err != nil {
			custom_log.NewCustomLog("scraping_failed", insert_err.Error(), "error")
			return nil, (&responses.ErrorResponse{}).NewErrorResponse("scraping_failed", fmt.Errorf("database_error"))
		}
	}

	return resp, err
}

func (sc *ScrapingService) Seed() {
	series := sc.ScrapingRepo.Seed()
	serie_repo := serie.NewSerieRepoImpl(sc.DBPool, sc.UserContext)
	for _, serie := range series {
		_, err := serie_repo.Create(serie)
		if err != nil {
			fmt.Println("Error Inserting Data : ", err.Err)
		}
	}

}
