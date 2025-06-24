package handlers

import (
	"errors"
	"net/http"
	"nodes-grpc-be/services"
	"strings"

	"github.com/gin-gonic/gin"
)

func Authenticate(jwtService *services.JwtService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// no Authorization header
		if authHeader == "" {
			res := NewErrorResponse(errors.New("unauthorized"), "no token provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		}

		// bad format
		if !strings.Contains(authHeader, "Bearer ") {
			res := NewErrorResponse(errors.New("unauthorized"), "invalid token format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		}

		headerValue := strings.ReplaceAll(authHeader, "Bearer ", "")
		token, err := jwtService.ValidateToken(headerValue)
		if err != nil {
			res := NewErrorResponse(errors.New("unauthorized"), "invalid token format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		}
		if !token.Valid {
			res := NewErrorResponse(errors.New("unauthorized"), "access denied")
			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		}

		c.Set("token", headerValue)

		c.Next()
	}
}
