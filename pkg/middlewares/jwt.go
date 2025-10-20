package middlewares

import (
	// "rerng_addicted_api/internal/front/auth"

	"fmt"
	"net/http"
	"os"
	"rerng_addicted_api/internal/front/auth"
	types "rerng_addicted_api/pkg/model"
	"rerng_addicted_api/pkg/utils"
	"strings"
	"time"

	response "rerng_addicted_api/pkg/http/response"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// func NewJwtMinddleWare(app *fiber.App, DBPool *sqlx.DB)  {
// 	// load environment variables
// 	errs := godotenv.Load()
// 	if errs != nil {
// 		log.Fatalf("Error loading .env file")
// 	}
// 	secret_key := os.Getenv("JWT_SECRET_KEY")

// 	// JWT middleware
// 	app.Use(func(c *fiber.Ctx) error {
// 		// check if the request is upgrading to WebSocket
// 		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
// 			// extract Bearer token from Sec-WebSocket-Protocol
// 			webSocketProtocol := c.Get("Sec-WebSocket-Protocol")
// 			if webSocketProtocol == "" {
// 				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
// 					"error": utils.Translate("missing_ws_auth_protocol", nil, c),
// 				})
// 			}

// 			// split "Bearer, <token>"
// 			parts := strings.Split(webSocketProtocol, ",")
// 			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
// 				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
// 					"error": utils.Translate("invalid_ws_auth_format", nil, c),
// 				})
// 			}

// 			// extract the JWT token from the second part
// 			tokenString := strings.TrimSpace(parts[1])

// 			// parse the JWT token
// 			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
// 				return []byte(secret_key), nil
// 			})
// 			if err != nil || !token.Valid {
// 				return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
// 					"error": utils.Translate("invalid_jwt_token", nil, c),
// 				})
// 			}
// 			c.Locals("jwt_data", token)
// 			// set the response header to echo back the protocol
// 			c.Set("Sec-WebSocket-Protocol", "Bearer")
// 			return c.Next()
// 		}

// 		// apply JWT middleware for HTTP requests
// 		return jwtware.New(jwtware.Config{
// 			SigningKey: jwtware.SigningKey{Key: []byte(secret_key)},
// 			ContextKey: "jwt_data",
// 			ErrorHandler: func(c *fiber.Ctx, err error) error {
// 				if errors.Is(err, jwt.ErrTokenExpired) {
// 					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 						"error": utils.Translate("session_expired", nil, c),
// 					})
// 				}
// 				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
// 					"error": utils.Translate("missing_or_malformed_jwt", nil, c),
// 				})
// 			},
// 		})(c)

// 	})

// 	// user context middleware
// 	app.Use(func(c *fiber.Ctx) error {
// 		// extract the JWT token data
// 		user_token := c.Locals("jwt_data").(*jwt.Token)
// 		pclaim := user_token.Claims.(jwt.MapClaims)

// 		// check if the connection is WebSocket and handle accordingly
// 		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
// 			// for WebSocket, ensure the token contains necessary claims
// 			return handlePlayerContext(c, pclaim, DBPool)
// 		}

// 		// handle regular HTTP requests
// 		return handlePlayerContext(c, pclaim, DBPool)
// 	})
// }

func NewJwtMiddleware(DBPool *sqlx.DB) fiber.Handler {
	_ = godotenv.Load()
	secretKey := os.Getenv("JWT_SECRET_KEY")

	return func(c *fiber.Ctx) error {
		// Handle WebSocket upgrade
		if websocketUpgrade := c.Get("Upgrade"); websocketUpgrade == "websocket" {
			wsProtocol := c.Get("Sec-WebSocket-Protocol")
			if wsProtocol == "" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": utils.Translate("missing_ws_auth_protocol", nil, c),
				})
			}

			parts := strings.Split(wsProtocol, ",")
			if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": utils.Translate("invalid_ws_auth_format", nil, c),
				})
			}

			tokenString := strings.TrimSpace(parts[1])
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": utils.Translate("invalid_jwt_token", nil, c),
				})
			}

			pclaim := token.Claims.(jwt.MapClaims)
			if err := handlePlayerContext(c, pclaim, DBPool); err != nil {
				return err
			}

			c.Set("Sec-WebSocket-Protocol", "Bearer")
			return c.Next()
		}

		// Handle regular HTTP requests
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": utils.Translate("missing_or_malformed_jwt", nil, c),
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.TrimSpace(parts[0]) != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": utils.Translate("missing_or_malformed_jwt", nil, c),
			})
		}

		tokenString := strings.TrimSpace(parts[1])
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secretKey), nil
		})
		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": utils.Translate("invalid_jwt_token", nil, c),
			})
		}

		pclaim := token.Claims.(jwt.MapClaims)
		if err := handlePlayerContext(c, pclaim, DBPool); err != nil {
			return err
		}

		return c.Next()
	}
}

// helper function to handle player context creation and session validation
func handlePlayerContext(c *fiber.Ctx, pclaim jwt.MapClaims, DBPool *sqlx.DB) error {
	// get user_uuid from claims
	user_uuid, ok := pclaim["user_uuid"].(string)
	if !ok || user_uuid == "" {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			utils.Translate("access_denied", nil, c),
			-500,
			fmt.Errorf(
				"%s", utils.Translate(
					"missing_or_invalid_key",
					map[string]interface{}{
						"key": "user_uuid",
					},
					c,
				),
			),
		))
	}

	// get login_session from claims
	login_session, ok := pclaim["login_session"].(string)
	if !ok || login_session == "" {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			utils.Translate("access_denied", nil, c),
			-500,
			fmt.Errorf(
				"%s", utils.Translate(
					"missing_or_invalid_key",
					map[string]interface{}{
						"key": "login_session",
					},
					c,
				),
			),
		))
	}

	// get exp from claims
	exp, ok := pclaim["exp"].(float64)
	if !ok {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			utils.Translate("access_denied", nil, c),
			-500,
			fmt.Errorf(
				"%s", utils.Translate(
					"missing_or_invalid_key",
					map[string]interface{}{
						"key": "exp",
					},
					c,
				),
			),
		))
	}

	// get user info for context
	user_info, err := auth.NewAuthRepoImpl(DBPool).GetUserByUUID(user_uuid)
	if err != nil {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			utils.Translate("access_denied", nil, c),
			-500,
			fmt.Errorf(
				"%s", utils.Translate(
					"get_userinfo_failed",
					nil,
					c,
				),
			),
		))
	}

	// check login session
	if login_session != user_info.LoginSession {
		return c.Status(http.StatusUnprocessableEntity).JSON(response.NewResponseError(
			utils.Translate("access_denied", nil, c),
			-500,
			fmt.Errorf(
				"%s", utils.Translate(
					"session_expired",
					nil,
					c,
				),
			),
		))
	}

	// Create and populate PlayerContext struct
	uCtx := types.UserContext{
		Id:           user_info.ID,
		UserUuid:     user_info.UserUUID,
		UserName:     user_info.UserName,
		LoginSession: login_session,
		RoleId:       uint64(user_info.RoleID),
		Exp:          time.Unix(int64(exp), 0),
		UserAgent:    string(c.Context().UserAgent()),
		Ip:           string(c.Context().RemoteIP().String()),
		StatusId:     user_info.StatusID,
	}

	c.Locals("UserContext", uCtx)
	return nil
}
