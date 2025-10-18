package config

import (
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

func NewFiber() *fiber.App {
	// app instance
	app := fiber.New()

	// middlewares
	app.Use(fiberlogger.New())

	return app
}
