package http

import (
	"strconv"

	"rsl-learning-generator/backend/internal/models"
	"rsl-learning-generator/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	generator := services.NewGenerationService(db)

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/api/translated-words", func(c *fiber.Ctx) error {
		search := c.Query("search")
		var rows []map[string]any
		query := db.Table("linguistic.translated_words tw").
			Select("tw.id, tw.display_text, w.name as word_name, c.name as concept_name, g.name as gesture_name").
			Joins("join linguistic.words w on w.id = tw.word_id").
			Joins("join linguistic.concepts c on c.id = tw.concept_id").
			Joins("join linguistic.gestures g on g.id = tw.gesture_id").
			Order("tw.id")
		if search != "" {
			query = query.Where("tw.display_text ilike ? or w.name ilike ? or g.name ilike ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
		if err := query.Find(&rows).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(rows)
	})

	app.Post("/api/themes/:id/generate", func(c *fiber.Ctx) error {
		themeID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "некорректный id темы"})
		}
		result, err := generator.Generate(themeID, 2)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error(), "result": result})
		}
		return c.JSON(result)
	})

	app.Get("/api/themes/:id/exercises", func(c *fiber.Ctx) error {
		themeID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "некорректный id темы"})
		}
		var exercises []models.Exercise
		err = db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("position_index")
		}).Where("theme_id = ?", themeID).Order("id").Find(&exercises).Error
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(exercises)
	})
}
