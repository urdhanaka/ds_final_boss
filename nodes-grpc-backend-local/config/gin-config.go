package config

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewGin() *gin.Engine {
	g := gin.Default()
	g.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowWebSockets: true,
		AllowHeaders:    []string{"*"},
	}))

	return g
}
