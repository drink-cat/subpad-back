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
	cols, err := tableColumns(db, "subpad_info")
	if err != nil {
		return err
	}
	_, hasSnake := cols["fee_addr"]
	if _, ok := cols["feeAddr"]; ok && !hasSnake {
		if err = renameColumn(db, "subpad_info", "feeAddr", "fee_addr"); err != nil {
			return err
		}
		hasSnake = true
	}
	if _, ok := cols["user_addr"]; ok && !hasSnake {
		if err = renameColumn(db, "subpad_info", "user_addr", "fee_addr"); err != nil {
			return err
		}
		return nil
	}
	if _, ok := cols["user_addr"]; ok {
		if err = db.Exec("ALTER TABLE `subpad_info` DROP COLUMN `user_addr`").Error; err != nil {
			return fmt.Errorf("drop subpad_info.user_addr: %w", err)
		}
	}
	return nil
}

func tableColumns(db *gorm.DB, table string) (map[string]struct{}, error) {
	var rows []struct {
		Name string `gorm:"column:name"`
	}
	var err error
	switch db.Dialector.Name() {
	case "mysql":
		err = db.Raw(`
			SELECT COLUMN_NAME AS name FROM information_schema.COLUMNS
			WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
		`, table).Scan(&rows).Error
	default:
		err = db.Raw("SELECT name FROM pragma_table_info(?)", table).Scan(&rows).Error
	}
	if err != nil {
		return nil, fmt.Errorf("list %s columns: %w", table, err)
	}
	cols := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		cols[row.Name] = struct{}{}
	}
	return cols, nil
}

func renameColumn(db *gorm.DB, table, oldName, newName string) error {
	var err error
	switch db.Dialector.Name() {
	case "mysql":
		err = db.Exec(
			fmt.Sprintf("ALTER TABLE `%s` CHANGE COLUMN `%s` `%s` varchar(100)", table, oldName, newName),
		).Error
	default:
		err = db.Exec(
			fmt.Sprintf("ALTER TABLE `%s` RENAME COLUMN `%s` TO `%s`", table, oldName, newName),
		).Error
	}
	if err != nil {
		return fmt.Errorf("rename %s.%s: %w", table, oldName, err)
	}
	return nil
}
