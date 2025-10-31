package serie

import (
	share "rerng_addicted_api/pkg/model"
	"github.com/jmoiron/sqlx"
)

type SerieService struct {
	DBPool       *sqlx.DB
	SerieRepo *SerieRepoImpl
	UserContext  *share.UserContext
}

func NewSerieService(db_pool *sqlx.DB, userCtx *share.UserContext) *SerieService {
	return &SerieService{
		DBPool: db_pool,
		UserContext: userCtx,
		SerieRepo: NewSerieRepoImpl(db_pool, userCtx),
	}
}

// TODO: add your service methods here
