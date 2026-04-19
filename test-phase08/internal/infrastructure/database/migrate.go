package database

import (
	"github.com/user/test-phase08/internal/domain/entity"
	"gorm.io/gorm"
)

// RunMigrations runs GORM auto-migrations
func RunMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
	)
}
