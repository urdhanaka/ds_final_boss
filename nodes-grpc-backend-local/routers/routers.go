package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetRouters(app *gin.Engine, dbPool *pgxpool.Pool) {
	apiGroup := app.Group("/api")
	v1 := apiGroup.Group("/v1")

	userGroup := v1.Group("/user")
	userGroup.GET("/")
}
