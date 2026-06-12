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
	if err := database.EnsureDemoAuthData(db); err != nil {
		log.Fatal(err)
	}
	if err := database.EnsureGenerationRunsSchema(db); err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()

		contentType := string(c.Response().Header.ContentType())
		if contentType == "" || contentType == fiber.MIMEApplicationJSON {
			c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
		}

		return err
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	httpserver.RegisterRoutes(app, db, cfg)

	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
