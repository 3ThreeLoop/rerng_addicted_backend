package auth

import (
	"rerng_addicted_api/pkg/responses"

	"github.com/jmoiron/sqlx"
)

type AuthServiceCreator interface {
	Login(username string, password string) (*LoginResponse, *responses.ErrorResponse)
}

type AuthService struct {
	DBPool   *sqlx.DB
	AuthRepo *AuthRepoImpl
}

func NewAuthService(db_pool *sqlx.DB) *AuthService {
	return &AuthService{
		DBPool:   db_pool,
		AuthRepo: NewAuthRepoImpl(db_pool),
	}
}

func (au *AuthService) Login(username string, password string) (*LoginResponse, *responses.ErrorResponse) {
	return au.AuthRepo.Login(username, password)
}
