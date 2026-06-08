package main

import (
	"log"

	"rsl-learning-generator/backend/internal/config"
	"rsl-learning-generator/backend/internal/database"
	httpserver "rsl-learning-generator/backend/internal/http"

	"github.com/gofiber/fiber/v2"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()
	httpserver.RegisterRoutes(app, db)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
