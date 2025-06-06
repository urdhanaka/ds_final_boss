package config

import "github.com/gin-gonic/gin"

func NewGin() *gin.Engine {
	g := gin.Default()

	return g
}
