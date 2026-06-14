package http

import (
	"strconv"
	"strings"
	"time"

	"rsl-learning-generator/backend/internal/auth"
	"rsl-learning-generator/backend/internal/config"
	"rsl-learning-generator/backend/internal/models"
	"rsl-learning-generator/backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type createCourseRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
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
	Comment string `json:"comment"`
}

const demoRoleHeader = "X-Demo-Role"

func requireDemoRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := ensureDemoRoleAccess(c, allowedRoles...); err != nil {
			return err
		}
		return c.Next()
	}
}

func currentDemoRole(c *fiber.Ctx) string {
	role := strings.TrimSpace(strings.ToLower(c.Get(demoRoleHeader)))

	switch role {
	case "expert", "learner":
		return role
	default:
		return "guest"
	}
}

func ensureDemoRoleAccess(c *fiber.Ctx, allowedRoles ...string) error {
	role := currentDemoRole(c)

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return nil
		}
	}

	if len(allowedRoles) == 1 && allowedRoles[0] == "expert" {
		return errorJSON(c, fiber.StatusForbidden, "экспертная часть доступна только пользователю с ролью эксперта")
	}

	return errorJSON(c, fiber.StatusForbidden, "рабочая тетрадь доступна только обучающемуся или эксперту")
}

func RegisterRoutes(app *fiber.App, db *gorm.DB, cfg config.Config) {
	generator := services.NewGenerationService(db)
	tokenManager := auth.NewTokenManager(cfg.AuthSecret, time.Duration(cfg.AuthTokenTTLHours)*time.Hour)

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/api/auth/login", func(c *fiber.Ctx) error {
		var req loginRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
		}

		if req.Username == "" || req.Password == "" {
			return errorJSON(c, fiber.StatusBadRequest, "укажите username и password")
		}

		user, found, err := loadDemoUser(db, req.Username)
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		if !found || !auth.CheckPassword(user.PasswordHash, req.Password) {
			return errorJSON(c, fiber.StatusUnauthorized, "неверный логин или пароль")
		}

		token, err := tokenManager.IssueToken(user)
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, "не удалось выпустить токен")
		}

		return c.JSON(fiber.Map{
			"token": token,
			"user":  publicUser(user),
		})
	})

	app.Get("/api/auth/me", auth.RequireRoles(tokenManager, "expert", "student"), func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		return c.JSON(publicUser(user))
	})

	app.Get("/api/public/courses", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.courses c").
			Select(`
				c.id,
				c.title,
				c.description,
				c.status,
				c.created_at,
				count(t.id) as themes_count
			`).
			Joins("left join learning.themes t on t.course_id = c.id and t.status = 'published'").
			Where("c.status = 'published'").
			Group("c.id, c.title, c.description, c.status, c.created_at").
			Order("c.id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	app.Get("/api/public/courses/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		var courseRows []map[string]any
		err = db.Table("learning.courses").
			Select("id, title, description, status, created_at, updated_at").
			Where("id = ? and status = 'published'", id).
			Find(&courseRows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		if len(courseRows) == 0 {
			return errorJSON(c, fiber.StatusNotFound, "опубликованный курс не найден")
		}
		var themes []map[string]any
		err = db.Table("learning.themes").
			Select("id, course_id, title, description, order_index, status, created_at, updated_at").
			Where("course_id = ? and status = 'published'", id).
			Order("order_index, id").
			Find(&themes).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(fiber.Map{
			"course": courseRows[0],
			"themes": themes,
		})
	})

	app.Get("/api/public/themes/:id/textbook", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		return publicThemeExercises(c, db, themeID, "textbook")
	})

	app.Get("/api/public/themes/:id/workbook", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}

		if roleErr := ensureDemoRoleAccess(c, "learner", "expert"); roleErr != nil {
			return roleErr
		}

		return publicThemeExercises(c, db, themeID, "workbook")
	})

	app.Get("/api/public/themes/:id/vocabulary", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		return publicThemeVocabulary(c, db, themeID)
	})

	app.Get("/api/public/themes/:id/textbook-full", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		return publicThemeFull(c, db, themeID, "textbook")
	})

	app.Get("/api/public/themes/:id/workbook-full", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}

		if roleErr := ensureDemoRoleAccess(c, "learner", "expert"); roleErr != nil {
			return roleErr
		}

		return publicThemeFull(c, db, themeID, "workbook")
	})

	expert := app.Group("/api", auth.RequireRoles(tokenManager, "expert"), requireDemoRoles("expert"))

	expert.Get("/translated-words", func(c *fiber.Ctx) error {
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
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	expert.Get("/courses", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.courses c").
			Select("c.id, c.title, c.description, c.status, c.created_by, u.full_name as creator_name, c.created_at, c.updated_at").
			Joins("join learning.users u on u.id = c.created_by").
			Order("c.id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	expert.Post("/courses", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		var req createCourseRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
		}
		if req.Title == "" {
			return errorJSON(c, fiber.StatusBadRequest, "название курса обязательно")
		}

		var id int64
		err := db.Raw(`
			insert into learning.courses (title, description, status, created_by)
			values (?, ?, 'draft', ?)
			returning id
		`, req.Title, req.Description, user.ID).Scan(&id).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		audit(db, user.ID, "create", "course", id)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "status": "created"})
	})

	expert.Get("/courses/:id/themes", func(c *fiber.Ctx) error {
		courseID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		var rows []map[string]any
		err = db.Table("learning.themes").
			Select("id, course_id, title, description, order_index, status, created_at, updated_at").
			Where("course_id = ?", courseID).
			Order("order_index, id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	expert.Post("/courses/:id/themes", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		courseID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		var req createThemeRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
		}
		if req.Title == "" {
			return errorJSON(c, fiber.StatusBadRequest, "название темы обязательно")
		}
		if req.OrderIndex == 0 {
			var nextIndex int
			err = db.Raw("select coalesce(max(order_index), 0) + 1 from learning.themes where course_id = ?", courseID).Scan(&nextIndex).Error
			if err != nil {
				return errorJSON(c, fiber.StatusInternalServerError, err.Error())
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
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		audit(db, user.ID, "create", "theme", id)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "status": "created"})
	})

	expert.Get("/courses/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		var rows []map[string]any
		err = db.Table("learning.courses c").
			Select("c.id, c.title, c.description, c.status, c.created_by, u.full_name as creator_name, c.created_at, c.updated_at").
			Joins("join learning.users u on u.id = c.created_by").
			Where("c.id = ?", id).
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return errorJSON(c, fiber.StatusNotFound, "курс не найден")
		}
		return c.JSON(rows[0])
	})

	expert.Put("/courses/:id", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		var req updateCourseRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
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
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "курс не найден")
		}
		audit(db, user.ID, "update", "course", id)
		return c.JSON(fiber.Map{"id": id, "status": "updated"})
	})

	expert.Post("/courses/:id/publish", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		result := db.Exec("update learning.courses set status = 'published', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "курс не найден")
		}
		audit(db, user.ID, "publish", "course", id)
		return c.JSON(fiber.Map{"id": id, "status": "published"})
	})

	expert.Post("/courses/:id/unpublish", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}
		result := db.Exec("update learning.courses set status = 'draft', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "курс не найден")
		}

		audit(db, user.ID, "unpublish", "course", id)
		return c.JSON(fiber.Map{"id": id, "status": "draft"})
	})

	expert.Delete("/courses/:id", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id курса")
		}

		err = deleteDraftCourse(db, id, user.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errorJSON(c, fiber.StatusNotFound, "курс не найден")
			}
			if fiberErr, ok := err.(*fiber.Error); ok {
				return errorJSON(c, fiberErr.Code, fiberErr.Message)
			}
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{"id": id, "status": "deleted"})
	})

	expert.Get("/themes/:id", func(c *fiber.Ctx) error {
		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		var rows []map[string]any
		err = db.Table("learning.themes").
			Select("id, course_id, title, description, order_index, status, created_at, updated_at").
			Where("id = ?", id).
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		if len(rows) == 0 {
			return errorJSON(c, fiber.StatusNotFound, "тема не найдена")
		}
		return c.JSON(rows[0])
	})

	expert.Put("/themes/:id", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		var req updateThemeRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
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
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "тема не найдена")
		}
		audit(db, user.ID, "update", "theme", id)
		return c.JSON(fiber.Map{"id": id, "status": "updated"})
	})

	expert.Post("/themes/:id/publish", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		result := db.Exec("update learning.themes set status = 'published', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "тема не найдена")
		}
		audit(db, user.ID, "publish", "theme", id)
		return c.JSON(fiber.Map{"id": id, "status": "published"})
	})

	expert.Post("/themes/:id/unpublish", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}

		result := db.Exec("update learning.themes set status = 'draft', updated_at = now() where id = ?", id)
		if result.Error != nil {
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return errorJSON(c, fiber.StatusNotFound, "тема не найдена")
		}

		audit(db, user.ID, "unpublish", "theme", id)
		return c.JSON(fiber.Map{"id": id, "status": "draft"})
	})

	expert.Delete("/themes/:id", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		id, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}

		err = deleteDraftTheme(db, id, user.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return errorJSON(c, fiber.StatusNotFound, "тема не найдена")
			}
			if fiberErr, ok := err.(*fiber.Error); ok {
				return errorJSON(c, fiberErr.Code, fiberErr.Message)
			}
			return errorJSON(c, fiber.StatusBadRequest, err.Error())
		}

		return c.JSON(fiber.Map{"id": id, "status": "deleted"})
	})

	expert.Get("/themes/:id/translated-words", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
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
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	expert.Post("/themes/:id/translated-words", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		var req addTranslatedWordRequest
		if err := c.BodyParser(&req); err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
		}
		if req.TranslatedWordID == 0 {
			return errorJSON(c, fiber.StatusBadRequest, "translated_word_id обязателен")
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
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		if count == 0 {
			return errorJSON(c, fiber.StatusNotFound, "переводное слово не найдено")
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
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		audit(db, user.ID, "add_translated_word", "theme", themeID)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id, "status": "saved"})
	})

	expert.Delete("/themes/:id/translated-words/:translatedWordId", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		translatedWordID, err := parseID(c, "translatedWordId")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id переводного слова")
		}
		result := db.Exec("delete from learning.theme_translated_words where theme_id = ? and translated_word_id = ?", themeID, translatedWordID)
		if result.Error != nil {
			return errorJSON(c, fiber.StatusInternalServerError, result.Error.Error())
		}
		audit(db, user.ID, "remove_translated_word", "theme", themeID)
		return c.JSON(fiber.Map{"status": "deleted", "deleted": result.RowsAffected})
	})

	expert.Post("/themes/:id/generate", func(c *fiber.Ctx) error {
		user, ok := auth.CurrentUser(c)
		if !ok {
			return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
		}

		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		result, err := generator.Generate(themeID, user.ID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error(), "result": result})
		}
		audit(db, user.ID, "generate", "theme", themeID)
		return c.JSON(result)
	})

	expert.Get("/themes/:id/exercises", func(c *fiber.Ctx) error {
		themeID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id темы")
		}
		var exercises []models.Exercise
		err = db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
			return db.Order("position_index")
		}).Where("theme_id = ?", themeID).Order("id").Find(&exercises).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(exercises)
	})

	expert.Put("/exercises/:id/approve", func(c *fiber.Ctx) error {
		return reviewExercise(c, db, "approved")
	})

	expert.Put("/exercises/:id/reject", func(c *fiber.Ctx) error {
		return reviewExercise(c, db, "rejected")
	})

	expert.Get("/generation-runs", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.generation_runs").
			Select("id, theme_id, started_by, status, found_examples, generated_exercises, rejected_examples, skipped_examples, duration_ms, error_message, created_at").
			Order("id desc").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})

	expert.Get("/generation-runs/:id/rejections", func(c *fiber.Ctx) error {
		runID, err := parseID(c, "id")
		if err != nil {
			return errorJSON(c, fiber.StatusBadRequest, "некорректный id запуска генерации")
		}

		var rows []models.GenerationRejection
		err = db.
			Where("generation_run_id = ?", runID).
			Order("id").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}

		return c.JSON(rows)
	})

	expert.Get("/audit", func(c *fiber.Ctx) error {
		var rows []map[string]any
		err := db.Table("learning.audit_logs a").
			Select("a.id, a.user_id, u.full_name as user_name, a.action, a.entity_type, a.entity_id, a.created_at").
			Joins("left join learning.users u on u.id = a.user_id").
			Order("a.id desc").
			Find(&rows).Error
		if err != nil {
			return errorJSON(c, fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(rows)
	})
}

func loadDemoUser(db *gorm.DB, username string) (auth.User, bool, error) {
	var row struct {
		ID           int64
		Role         string
		FullName     string
		PasswordHash string
	}

	err := db.Table("learning.users u").
		Select("u.id, r.code as role, u.full_name, u.password_hash").
		Joins("join learning.roles r on r.id = u.role_id").
		Where("r.code = ? and r.code in ('expert', 'student')", username).
		Take(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return auth.User{}, false, nil
		}
		return auth.User{}, false, err
	}

	return auth.User{
		ID:           row.ID,
		Role:         row.Role,
		FullName:     row.FullName,
		PasswordHash: row.PasswordHash,
	}, true, nil
}

func publicUser(user auth.User) fiber.Map {
	return fiber.Map{
		"id":        user.ID,
		"role":      user.Role,
		"full_name": user.FullName,
	}
}

func deleteDraftCourse(db *gorm.DB, courseID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var course struct {
			ID     int64
			Status string
		}

		err := tx.Table("learning.courses").
			Select("id, status").
			Where("id = ?", courseID).
			Take(&course).Error
		if err != nil {
			return err
		}

		if course.Status != "draft" {
			return fiber.NewError(fiber.StatusBadRequest, "можно удалить только черновой курс")
		}

		var publishedThemesCount int64
		err = tx.Table("learning.themes").
			Where("course_id = ? and status <> 'draft'", courseID).
			Count(&publishedThemesCount).Error
		if err != nil {
			return err
		}

		if publishedThemesCount > 0 {
			return fiber.NewError(fiber.StatusBadRequest, "нельзя удалить курс, в котором есть опубликованные темы")
		}

		var themeIDs []int64
		err = tx.Table("learning.themes").
			Where("course_id = ?", courseID).
			Pluck("id", &themeIDs).Error
		if err != nil {
			return err
		}

		for _, themeID := range themeIDs {
			if err := deleteThemeData(tx, themeID); err != nil {
				return err
			}
		}

		if err := tx.Exec("delete from learning.themes where course_id = ?", courseID).Error; err != nil {
			return err
		}

		result := tx.Exec("delete from learning.courses where id = ?", courseID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		audit(tx, userID, "delete", "course", courseID)
		return nil
	})
}

func deleteDraftTheme(db *gorm.DB, themeID int64, userID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var theme struct {
			ID     int64
			Status string
		}

		err := tx.Table("learning.themes").
			Select("id, status").
			Where("id = ?", themeID).
			Take(&theme).Error
		if err != nil {
			return err
		}

		if theme.Status != "draft" {
			return fiber.NewError(fiber.StatusBadRequest, "можно удалить только черновую тему")
		}

		if err := deleteThemeData(tx, themeID); err != nil {
			return err
		}

		result := tx.Exec("delete from learning.themes where id = ?", themeID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		audit(tx, userID, "delete", "theme", themeID)
		return nil
	})
}

func deleteThemeData(tx *gorm.DB, themeID int64) error {
	if err := tx.Exec(`
		delete from learning.exercise_reviews
		where exercise_id in (
			select id from learning.exercises where theme_id = ?
		)
	`, themeID).Error; err != nil {
		return err
	}

	if err := tx.Exec(`
		delete from learning.exercise_segments
		where exercise_id in (
			select id from learning.exercises where theme_id = ?
		)
	`, themeID).Error; err != nil {
		return err
	}

	if err := tx.Exec("delete from learning.exercises where theme_id = ?", themeID).Error; err != nil {
		return err
	}

	if err := tx.Exec("delete from learning.generation_rejections where theme_id = ?", themeID).Error; err != nil {
		return err
	}

	if err := tx.Exec("delete from learning.generation_runs where theme_id = ?", themeID).Error; err != nil {
		return err
	}

	if err := tx.Exec("delete from learning.theme_translated_words where theme_id = ?", themeID).Error; err != nil {
		return err
	}

	return nil
}

func reviewExercise(c *fiber.Ctx, db *gorm.DB, decision string) error {
	user, ok := auth.CurrentUser(c)
	if !ok {
		return errorJSON(c, fiber.StatusUnauthorized, "требуется авторизация")
	}

	id, err := parseID(c, "id")
	if err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "некорректный id упражнения")
	}
	var req reviewExerciseRequest
	if err := c.BodyParser(&req); err != nil {
		return errorJSON(c, fiber.StatusBadRequest, "не удалось прочитать тело запроса")
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
		`, id, user.ID, decision, req.Comment).Error
	})
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorJSON(c, fiber.StatusNotFound, "упражнение не найдено")
		}
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	audit(db, user.ID, decision, "exercise", id)
	return c.JSON(fiber.Map{"id": id, "status": decision})
}

func parseID(c *fiber.Ctx, name string) (int64, error) {
	return strconv.ParseInt(c.Params(name), 10, 64)
}

func errorJSON(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{"error": message})
}

func publicThemeExercises(c *fiber.Ctx, db *gorm.DB, themeID int64, targetMode string) error {
	var themeRows []map[string]any
	err := db.Table("learning.themes t").
		Select("t.id, t.course_id, t.title, t.description, t.order_index, t.status").
		Joins("join learning.courses c on c.id = t.course_id").
		Where("t.id = ? and t.status = 'published' and c.status = 'published'", themeID).
		Find(&themeRows).Error
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	if len(themeRows) == 0 {
		return errorJSON(c, fiber.StatusNotFound, "опубликованная тема не найдена")
	}

	var exercises []models.Exercise
	err = db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
		return db.Order("position_index")
	}).
		Where("theme_id = ? and target_mode = ? and status = 'approved'", themeID, targetMode).
		Order("id").
		Find(&exercises).Error
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"theme":       themeRows[0],
		"target_mode": targetMode,
		"exercises":   exercises,
	})
}

func publicThemeVocabulary(c *fiber.Ctx, db *gorm.DB, themeID int64) error {
	theme, found, err := loadPublishedTheme(db, themeID)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	if !found {
		return errorJSON(c, fiber.StatusNotFound, "опубликованная тема не найдена")
	}

	vocabulary, err := loadThemeVocabulary(db, themeID)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"theme": theme,
		"items": vocabulary,
	})
}

func publicThemeFull(c *fiber.Ctx, db *gorm.DB, themeID int64, targetMode string) error {
	theme, found, err := loadPublishedTheme(db, themeID)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}
	if !found {
		return errorJSON(c, fiber.StatusNotFound, "опубликованная тема не найдена")
	}

	vocabulary, err := loadThemeVocabulary(db, themeID)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	exercises, err := loadPublishedThemeExercises(db, themeID, targetMode)
	if err != nil {
		return errorJSON(c, fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{
		"theme":       theme,
		"target_mode": targetMode,
		"vocabulary":  vocabulary,
		"exercises":   exercises,
	})
}

func loadPublishedTheme(db *gorm.DB, themeID int64) (map[string]any, bool, error) {
	var rows []map[string]any
	err := db.Table("learning.themes t").
		Select(`
			t.id,
			t.course_id,
			t.title,
			t.description,
			t.order_index,
			t.status,
			c.title as course_title
		`).
		Joins("join learning.courses c on c.id = t.course_id").
		Where("t.id = ? and t.status = 'published' and c.status = 'published'", themeID).
		Find(&rows).Error
	if err != nil {
		return nil, false, err
	}
	if len(rows) == 0 {
		return nil, false, nil
	}
	return rows[0], true, nil
}

func loadThemeVocabulary(db *gorm.DB, themeID int64) ([]map[string]any, error) {
	var rows []map[string]any
	err := db.Table("learning.theme_translated_words ttw").
		Select(`
			ttw.id,
			ttw.theme_id,
			ttw.translated_word_id,
			ttw.difficulty_level,
			ttw.is_required,
			tw.display_text,
			w.name as word_name,
			c.name as concept_name,
			c.description as concept_description,
			g.name as gesture_name,
			g.video_url,
			g.description as gesture_description
		`).
		Joins("join linguistic.translated_words tw on tw.id = ttw.translated_word_id").
		Joins("join linguistic.words w on w.id = tw.word_id").
		Joins("join linguistic.concepts c on c.id = tw.concept_id").
		Joins("left join linguistic.gestures g on g.id = tw.gesture_id").
		Where("ttw.theme_id = ?", themeID).
		Order("ttw.difficulty_level, w.name, tw.id").
		Find(&rows).Error
	return rows, err
}

func loadPublishedThemeExercises(db *gorm.DB, themeID int64, targetMode string) ([]models.Exercise, error) {
	var exercises []models.Exercise
	err := db.Preload("Segments", func(db *gorm.DB) *gorm.DB {
		return db.Order("position_index")
	}).
		Where("theme_id = ? and target_mode = ? and status = 'approved'", themeID, targetMode).
		Order("id").
		Find(&exercises).Error
	return exercises, err
}

func audit(db *gorm.DB, userID int64, action string, entityType string, entityID int64) {
	db.Exec(`
		insert into learning.audit_logs (user_id, action, entity_type, entity_id)
		values (?, ?, ?, ?)
	`, userID, action, entityType, entityID)
}
