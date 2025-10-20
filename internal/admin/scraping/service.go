package scraping

import (
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/responses"

	"github.com/jmoiron/sqlx"
)

type ScrapingServiceCreator interface {
	Search(keyword string) (*SeriesResponse, *responses.ErrorResponse)
	GetDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse)
	GetDeepDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse)
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

func (sc *ScrapingService) GetDetail(key string) (*SeriesDetailsResponse, *responses.ErrorResponse) {
	return sc.ScrapingRepo.GetDetail(key)
}

func (sc *ScrapingService) GetDeepDetail(key string) (*SeriesDeepDetailsResponse, *responses.ErrorResponse) {
	return sc.ScrapingRepo.GetDeepDetail(key)
}
