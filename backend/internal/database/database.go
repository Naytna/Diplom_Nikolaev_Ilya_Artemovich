package database

import (
	"rsl-learning-generator/backend/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	expertDemoPasswordHash  = "$2a$10$Am7pYLO.rTA7JR/R6v2ek.0IHku6orwM6rEGsOEGBoyvehwp5Us5i"
	studentDemoPasswordHash = "$2a$10$FxM3LCuF.Y6m3niPzCQNquiUv3uLNUzUGSAhO7opFVdQh3g7X8Jne"
)

func Connect(cfg config.Config) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
}

func EnsureDemoAuthData(db *gorm.DB) error {
	updates := []struct {
		roleCode string
		hash     string
	}{
		{roleCode: "expert", hash: expertDemoPasswordHash},
		{roleCode: "student", hash: studentDemoPasswordHash},
	}

	for _, update := range updates {
		if err := db.Exec(`
			update learning.users as u
			set password_hash = ?
			from learning.roles as r
			where r.id = u.role_id
				and r.code = ?
				and u.password_hash <> ?
		`, update.hash, update.roleCode, update.hash).Error; err != nil {
			return err
		}
	}

	return nil
}
