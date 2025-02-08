package config

import (
	"github.com/gofiber/fiber/v3"
    fiberlogger "github.com/gofiber/fiber/v3/middleware/logger"
)

func NewFiber() *fiber.App {
	app := fiber.New()
	app.Use(fiberlogger.New())

    return app
}
