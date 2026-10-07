package model

import (
	"fmt"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	err := db.AutoMigrate(
		&UserInfo{},
		&TokenInfo{},
		&SubpadInfo{},
		&FeeInfo{},
		&SyncEvent{},
		&SyncCursor{},
	)
	if err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}
	return nil
}
