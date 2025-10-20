package auth

import (
	"fmt"
	"log"
	"os"
	custom_log "rerng_addicted_api/pkg/logs"
	"rerng_addicted_api/pkg/responses"
	"rerng_addicted_api/pkg/utils"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type AuthRepo interface {
	Login(username string, password string) (*LoginResponse, *responses.ErrorResponse)
}

type AuthRepoImpl struct {
	DBPool *sqlx.DB
}

func NewAuthRepoImpl(db_pool *sqlx.DB) *AuthRepoImpl {
	return &AuthRepoImpl{
		DBPool: db_pool,
	}
}

func (au *AuthRepoImpl) Login(username string, password string) (*LoginResponse, *responses.ErrorResponse) {
	var users []User

	// prepare sql
	sql := `
		SELECT
			user_uuid
		FROM tbl_users
		WHERE deleted_at IS NULL 
		AND user_name = $1 
		AND password = $2
	`

	// execute request
	if err := au.DBPool.Select(&users, sql, username, password); err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("login_failed", fmt.Errorf("username_or_password_invalid"))
	}

	if len(users) == 0 {
		custom_log.NewCustomLog("login_failed", "no_user_found", "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("login_failed", fmt.Errorf("username_or_password_invalid"))
	}

	user := users[0]

	hours := utils.GetenvInt("JWT_EXP_HOUR", 7)
	expirationTime := time.Now().Add(time.Duration(hours) * time.Hour)

	// create the JWT claims
	login_session, _ := uuid.NewV7()
	claims := jwt.MapClaims{
		"user_uuid":     user.UserUUID,
		"login_session": login_session.String(),
		"exp":           expirationTime.Unix(),
	}

	errs := godotenv.Load()
	if errs != nil {
		log.Fatalf("Error loading .env file")
	}
	secret_key := os.Getenv("JWT_SECRET_KEY")

	// prepare sql
	update_sql := `
		UPDATE tbl_users SET
			login_session = $1
		WHERE deleted_at IS NULL
		AND user_uuid = $2 
	`

	// execute request
	_, err := au.DBPool.Exec(update_sql, login_session, user.UserUUID)
	if err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("login_failed", fmt.Errorf("error_database"))
	}

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret_key))
	if err != nil {
		custom_log.NewCustomLog("login_failed", err.Error(), "error")
		err_msg := &responses.ErrorResponse{}
		return nil, err_msg.NewErrorResponse("login_failed", fmt.Errorf("error_create_token"))
	}

	return &LoginResponse{
		Auth: Auth{
			Token:     tokenString,
			TokenType: "JWT",
		},
	}, nil
}

func (au *AuthRepoImpl) GetUserByUUID(user_uuid string) (*UserInfo, error) {
	var user_info UserInfo

	// prepare sql
	sql := `
		SELECT
			id, user_uuid, user_name,
			role_id, login_session, status_id
		FROM tbl_users
		WHERE deleted_at IS NULL 
		AND user_uuid = $1
	`

	// execute request
	if err := au.DBPool.Get(&user_info, sql, user_uuid); err != nil {
		custom_log.NewCustomLog("get_userinfo_failed", err.Error(), "error")
		return nil, err
	}

	return &user_info, nil
}
