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
	if err := widenAmountColumns(db); err != nil {
		return err
	}
	err := db.AutoMigrate(
		&UserInfo{},
		&TokenInfo{},
		&SubpadInfo{},
		&FeeInfo{},
		&SwapInfo{},
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

// widenAmountColumns 把已经建成 bigint 的金额列改成十进制字符串。链上 uint256 放不进 bigint。
func widenAmountColumns(db *gorm.DB) error {
	if db.Dialector.Name() != "mysql" {
		return nil
	}
	columns := [][2]string{
		{"token_info", "launch_supply"},
		{"fee_info", "fee_amount"},
		{"swap_info", "token_amount"},
		{"swap_info", "quote_amount"},
		{"swap_info", "fee"},
		{"swap_info", "price"},
	}
	for _, item := range columns {
		if err := widenAmountColumn(db, item[0], item[1]); err != nil {
			return err
		}
	}
	return nil
}

func widenAmountColumn(db *gorm.DB, table, column string) error {
	if !db.Migrator().HasTable(table) || !db.Migrator().HasColumn(table, column) {
		return nil
	}
	var dataType string
	err := db.Raw(`
		SELECT DATA_TYPE FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, column).Scan(&dataType).Error
	if err != nil {
		return fmt.Errorf("column type %s.%s: %w", table, column, err)
	}
	if dataType == "" || dataType == "varchar" {
		return nil
	}
	err = db.Exec(fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN `%s` varchar(80) NOT NULL", table, column)).Error
	if err != nil {
		return fmt.Errorf("widen %s.%s: %w", table, column, err)
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
