package model

import (
	"fmt"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	if err := renameSubpadFeeAddr(db); err != nil {
		return err
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

func renameSubpadFeeAddr(db *gorm.DB) error {
	m := db.Migrator()
	if !m.HasTable(&SubpadInfo{}) {
		return nil
	}
	hasOld := m.HasColumn(&SubpadInfo{}, "user_addr")
	hasNew := m.HasColumn(&SubpadInfo{}, "feeAddr")
	if hasOld && !hasNew {
		if err := m.RenameColumn(&SubpadInfo{}, "user_addr", "feeAddr"); err != nil {
			return fmt.Errorf("rename subpad_info.user_addr: %w", err)
		}
		return nil
	}
	if hasOld && hasNew {
		if err := m.DropColumn(&SubpadInfo{}, "user_addr"); err != nil {
			return fmt.Errorf("drop subpad_info.user_addr: %w", err)
		}
	}
	return nil
}
