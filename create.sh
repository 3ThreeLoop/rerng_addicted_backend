#!/bin/bash

echo "Welcome to the Go module generator!"

# -----------------------
# Ask for module name
# -----------------------
read -p "Enter module name: " MODULE_NAME
while [ -z "$MODULE_NAME" ]; do
  read -p "Module name cannot be empty. Enter module name: " MODULE_NAME
done

# -----------------------
# Ask for category/folder
# -----------------------
read -p "Enter category folder (default: internal/admin): " CATEGORY_DIR
CATEGORY_DIR=${CATEGORY_DIR:-internal/admin}

# -----------------------
# Check go.mod
# -----------------------
if [ ! -f go.mod ]; then
  echo "go.mod not found! Run this script in the project root."
  exit 1
fi

BASE_MODULE=$(grep '^module ' go.mod | awk '{print $2}')

# -----------------------
# Create module folder
# -----------------------
MODULE_PATH="$CATEGORY_DIR/$MODULE_NAME"
mkdir -p "$MODULE_PATH"

# -----------------------
# route.go
# -----------------------
cat <<EOL > "$MODULE_PATH/route.go"
package $MODULE_NAME

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

type ${MODULE_NAME^}Route struct {
	App             *fiber.App
	DBPool          *sqlx.DB
	${MODULE_NAME^}Handler *${MODULE_NAME^}Handler
}

func NewRoute(app *fiber.App, db_pool *sqlx.DB) *${MODULE_NAME^}Route {
	return &${MODULE_NAME^}Route{
		App:             app,
		DBPool:          db_pool,
		${MODULE_NAME^}Handler: New${MODULE_NAME^}Handler(db_pool),
	}
}

func (r *${MODULE_NAME^}Route) Register${MODULE_NAME^}Route() *${MODULE_NAME^}Route {
	//group := r.App.Group("/api/v1/admin/$MODULE_NAME")

	// TODO: add your routes here
	// group.Get("/search", middlewares.NewJwtMiddleware(r.DBPool), r.${MODULE_NAME^}Handler.Search)

	return r
}
EOL

# -----------------------
# handler.go
# -----------------------
cat <<EOL > "$MODULE_PATH/handler.go"
package $MODULE_NAME

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
	share "$BASE_MODULE/pkg/model"
)

type ${MODULE_NAME^}Handler struct {
	DBPool          *sqlx.DB
	${MODULE_NAME^}Service func(c *fiber.Ctx) *${MODULE_NAME^}Service
}

func New${MODULE_NAME^}Handler(db_pool *sqlx.DB) *${MODULE_NAME^}Handler {
	return &${MODULE_NAME^}Handler{
		DBPool: db_pool,
		${MODULE_NAME^}Service: func(c *fiber.Ctx) *${MODULE_NAME^}Service {
			var uCtx share.UserContext
			uCtx, ok := c.Locals("UserContext").(share.UserContext)
			if !ok {
				uCtx = share.UserContext{}
			}
			return New${MODULE_NAME^}Service(db_pool, &uCtx)
		},
	}
}
EOL

# -----------------------
# service.go
# -----------------------
cat <<EOL > "$MODULE_PATH/service.go"
package $MODULE_NAME

import (
	share "$BASE_MODULE/pkg/model"
	"github.com/jmoiron/sqlx"
)

type ${MODULE_NAME^}Service struct {
	DBPool       *sqlx.DB
	${MODULE_NAME^}Repo *${MODULE_NAME^}RepoImpl
	UserContext  *share.UserContext
}

func New${MODULE_NAME^}Service(db_pool *sqlx.DB, userCtx *share.UserContext) *${MODULE_NAME^}Service {
	return &${MODULE_NAME^}Service{
		DBPool: db_pool,
		UserContext: userCtx,
		${MODULE_NAME^}Repo: New${MODULE_NAME^}RepoImpl(db_pool, userCtx),
	}
}

// TODO: add your service methods here
EOL

# -----------------------
# repository.go
# -----------------------
cat <<EOL > "$MODULE_PATH/repository.go"
package $MODULE_NAME

import (
	"github.com/jmoiron/sqlx"
	share "$BASE_MODULE/pkg/model"
)

type ${MODULE_NAME^}Repo interface {
	// TODO: define repository methods
}

type ${MODULE_NAME^}RepoImpl struct {
	DBPool      *sqlx.DB
	UserContext *share.UserContext
}

func New${MODULE_NAME^}RepoImpl(db_pool *sqlx.DB, userCtx *share.UserContext) *${MODULE_NAME^}RepoImpl {
	return &${MODULE_NAME^}RepoImpl{
		DBPool: db_pool,
		UserContext: userCtx,
	}
}
EOL

# -----------------------
# model.go
# -----------------------
cat <<EOL > "$MODULE_PATH/model.go"
package $MODULE_NAME

type ${MODULE_NAME^} struct {
	ID   int
	Name string
	// TODO: add your model fields
}
EOL

echo "✅ Module '$MODULE_NAME' created successfully in '$MODULE_PATH'!"
