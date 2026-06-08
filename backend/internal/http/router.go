package http

import (
	"strconv"

	"rsl-learning-generator/backend/internal/models"
	"rsl-learning-generator/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type createCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedBy   int64  `json:"created_by"`
}

type updateCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type createThemeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
}

type updateThemeRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	OrderIndex  int    `json:"order_index"`
	Status      string `json:"status"`
}

type addTranslatedWordRequest struct {
	TranslatedWordID int64 `json:"translated_word_id"`
	DifficultyLevel  int   `json:"difficulty_level"`
	IsRequired       *bool `json:"is_required"`
}

type reviewExerciseRequest struct {
	ReviewerID int64  `json:"reviewer_id"`
	Comment    string `json:"comment"`
}

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
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})

	app.Get("/api/courses", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.courses c").
			Select("c.id, c.title, c.description, c.status, c.created_by, u.full_name as creator_name, c.created_at, c.updated_at").
			Joins("join learning.users u on u.id = c.created_by").
			Order("c.id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})

	app.Post("/api/courses", func(c *fiber.Ctx) error {
		var req createCourseRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, 400, "не удалось прочитать тело запроса")
		}
		if req.Title == "" {
			return errorJSON(c, 400, "название курса обязательно")
		}
		if req.CreatedBy == 0 {
			req.CreatedBy = 2
		}
		var id int64
		err := db.Raw(`
			insert into learning.courses (title, description, status, created_by)
			values (?, ?, 'draft', ?)
			returning id
		`, req.Title, req.Description, req.CreatedBy).Scan(&id).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		audit(db, req.CreatedBy, "create", "course", id)
		return c.Status(201).JSON(fiber.Map{"id": id, "status": "created"})
	})

	app.Get("/api/courses/:id/themes", func(c *fiber.Ctx) error {
		courseID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id курса")
		}
		var rows []map[string]any
		err = db.Table("learning.themes").
			Select("id, course_id, title, description, order_index, status, created_at, updated_at").
			Where("course_id = ?", courseID).
			Order("order_index, id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})

	app.Post("/api/courses/:id/themes", func(c *fiber.Ctx) error {
		courseID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id курса")
		}
		var req createThemeRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, 400, "не удалось прочитать тело запроса")
		}
		if req.Title == "" {
			return errorJSON(c, 400, "название темы обязательно")
		}
		if req.OrderIndex == 0 {
			var nextIndex int
			err = db.Raw("select coalesce(max(order_index), 0) + 1 from learning.themes where course_id = ?", courseID).Scan(&nextIndex).Error
			if err != nil {
				return errorJSON(c, 500, err.Error())
			}
			req.OrderIndex = nextIndex
		}
		var id int64
		err = db.Raw(`
			insert into learning.themes (course_id, title, description, order_index, status)
			values (?, ?, ?, ?, 'draft')
			returning id
		`, courseID, req.Title, req.Description, req.OrderIndex).Scan(&id).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		audit(db, 2, "create", "theme", id)
		return c.Status(201).JSON(fiber.Map{"id": id, "status": "created"})
	})

	app.Get("/api/courses/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id курса")
		}
		var rows []map[string]any
		err = db.Table("learning.courses c").
			Select("c.id, c.title, c.description, c.status, c.created_by, u.full_name as creator_name, c.created_at, c.updated_at").
			Joins("join learning.users u on u.id = c.created_by").
			Where("c.id = ?", id).
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		if len(rows) == 0 {
			return errorJSON(c, 404, "курс не найден")
		}
		return c.JSON(rows[0])
	})

	app.Put("/api/courses/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id курса")
		}
		var req updateCourseRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, 400, "не удалось прочитать тело запроса")
		}
		result := db.Exec(`
			update learning.courses
			set title = coalesce(nullif(?, ''), title),
				description = ?,
				status = coalesce(nullif(?, ''), status),
				updated_at = now()
			where id = ?
		`, req.Title, req.Description, req.Status, id)
		if result.Error != nil {
			return errorJSON(c, 500, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, 404, "курс не найден")
		}
		audit(db, 2, "update", "course", id)
		return c.JSON(fiber.Map{"id": id, "status": "updated"})
	})

	app.Post("/api/courses/:id/publish", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id курса")
		}
		result := db.Exec("update learning.courses set status = 'published', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, 500, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, 404, "курс не найден")
		}
		audit(db, 2, "publish", "course", id)
		return c.JSON(fiber.Map{"id": id, "status": "published"})
	})

	app.Get("/api/themes/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		var rows []map[string]any
		err = db.Table("learning.themes").
			Select("id, course_id, title, description, order_index, status, created_at, updated_at").
			Where("id = ?", id).
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		if len(rows) == 0 {
			return errorJSON(c, 404, "тема не найдена")
		}
		return c.JSON(rows[0])
	})

	app.Put("/api/themes/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		var req updateThemeRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, 400, "не удалось прочитать тело запроса")
		}
		result := db.Exec(`
			update learning.themes
			set title = coalesce(nullif(?, ''), title),
				description = ?,
				order_index = case when ? = 0 then order_index else ? end,
				status = coalesce(nullif(?, ''), status),
				updated_at = now()
			where id = ?
		`, req.Title, req.Description, req.OrderIndex, req.OrderIndex, req.Status, id)
		if result.Error != nil {
			return errorJSON(c, 500, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, 404, "тема не найдена")
		}
		audit(db, 2, "update", "theme", id)
		return c.JSON(fiber.Map{"id": id, "status": "updated"})
	})

	app.Post("/api/themes/:id/publish", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		result := db.Exec("update learning.themes set status = 'published', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, 500, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, 404, "тема не найдена")
		}
		audit(db, 2, "publish", "theme", id)
		return c.JSON(fiber.Map{"id": id, "status": "published"})
	})

	app.Get("/api/themes/:id/translated-words", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		var rows []map[string]any
		err = db.Table("learning.theme_translated_words ttw").
			Select("ttw.id, ttw.theme_id, ttw.translated_word_id, ttw.difficulty_level, ttw.is_required, tw.display_text, w.name as word_name, c.name as concept_name, g.name as gesture_name").
			Joins("join linguistic.translated_words tw on tw.id = ttw.translated_word_id").
			Joins("join linguistic.words w on w.id = tw.word_id").
			Joins("join linguistic.concepts c on c.id = tw.concept_id").
			Joins("join linguistic.gestures g on g.id = tw.gesture_id").
			Where("ttw.theme_id = ?", themeID).
			Order("ttw.id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})

	app.Post("/api/themes/:id/translated-words", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		var req addTranslatedWordRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, 400, "не удалось прочитать тело запроса")
		}
		if req.TranslatedWordID == 0 {
			return errorJSON(c, 400, "translated_word_id обязателен")
		}
		if req.DifficultyLevel == 0 {
			req.DifficultyLevel = 1
		}
		isRequired := true
		if req.IsRequired != nil {
			isRequired = *req.IsRequired
		}
		var count int64
		err = db.Table("linguistic.translated_words").Where("id = ?", req.TranslatedWordID).Count(&count).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		if count == 0 {
			return errorJSON(c, 404, "переводное слово не найдено")
		}
		var id int64
		err = db.Raw(`
			insert into learning.theme_translated_words (theme_id, translated_word_id, difficulty_level, is_required)
			values (?, ?, ?, ?)
			on conflict (theme_id, translated_word_id)
			do update set difficulty_level = excluded.difficulty_level, is_required = excluded.is_required
			returning id
		`, themeID, req.TranslatedWordID, req.DifficultyLevel, isRequired).Scan(&id).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		audit(db, 2, "add_translated_word", "theme", themeID)
		return c.Status(201).JSON(fiber.Map{"id": id, "status": "saved"})
	})

	app.Delete("/api/themes/:id/translated-words/:translatedWordId", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		translatedWordID, err := parseID(c, "translatedWordId")
		if err != nil {
			return errorJSON(c, 400, "некорректный id переводного слова")
		}
		result := db.Exec("delete from learning.theme_translated_words where theme_id = ? and translated_word_id = ?", themeID, translatedWordID)
		if result.Error != nil {
			return errorJSON(c, 500, result.Error.Error())
		}
		audit(db, 2, "remove_translated_word", "theme", themeID)
		return c.JSON(fiber.Map{"status": "deleted", "deleted": result.RowsAffected})
	})

	app.Post("/api/themes/:id/generate", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		result, err := generator.Generate(themeID, 2)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error(), "result": result})
		}
		return c.JSON(result)
	})

	app.Get("/api/themes/:id/exercises", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, 400, "некорректный id темы")
		}
		var exercises []models.Exercise
		err = db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("position_index")
		}).Where("theme_id = ?", themeID).Order("id").Find(&exercises).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(exercises)
	})

	app.Put("/api/exercises/:id/approve", func(c *fiber.Ctx) error {
		return reviewExercise(c, db, "approved")
	})

	app.Put("/api/exercises/:id/reject", func(c *fiber.Ctx) error {
		return reviewExercise(c, db, "rejected")
	})

	app.Get("/api/generation-runs", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.generation_runs").
			Select("id, theme_id, started_by, status, found_examples, generated_exercises, rejected_examples, duration_ms, error_message, created_at").
			Order("id desc").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})

	app.Get("/api/audit", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.audit_logs a").
			Select("a.id, a.user_id, u.full_name as user_name, a.action, a.entity_type, a.entity_id, a.created_at").
			Joins("left join learning.users u on u.id = a.user_id").
			Order("a.id desc").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, 500, err.Error())
		}
		return c.JSON(rows)
	})
}

func reviewExercise(c *fiber.Ctx, db *gorm.DB, decision string) error {
	id, err := parseID(c, "id")
	if err != nil {
		return errorJSON(c, 400, "некорректный id упражнения")
	}
	var req reviewExerciseRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, 400, "не удалось прочитать тело запроса")
	}
	if req.ReviewerID == 0 {
		req.ReviewerID = 2
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec("update learning.exercises set status = ?, updated_at = now() where id = ?", decision, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Exec(`
			insert into learning.exercise_reviews (exercise_id, reviewer_id, decision, comment)
			values (?, ?, ?, ?)
		`, id, req.ReviewerID, decision, req.Comment).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorJSON(c, 404, "упражнение не найдено")
		}
		return errorJSON(c, 500, err.Error())
	}
	audit(db, req.ReviewerID, decision, "exercise", id)
	return c.JSON(fiber.Map{"id": id, "status": decision})
}

func parseID(c *fiber.Ctx, name string) (int64, error) {
	return strconv.ParseInt(c.Params(name), 10, 64)
}

func errorJSON(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func audit(db *gorm.DB, userID int64, action string, entityType string, entityID int64) {
	db.Exec(`
		insert into learning.audit_logs (user_id, action, entity_type, entity_id)
		values (?, ?, ?, ?)
	`, userID, action, entityType, entityID)
}