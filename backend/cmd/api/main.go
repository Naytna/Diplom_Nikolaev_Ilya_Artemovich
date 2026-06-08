package main

import (
	"log"

	"rsl-learning-generator/backend/internal/config"
	"rsl-learning-generator/backend/internal/database"
	httpserver "rsl-learning-generator/backend/internal/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	httpserver.RegisterRoutes(app, db)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
