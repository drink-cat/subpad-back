package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

func createRow[T any](ctx context.Context, db *gorm.DB, table string, row *T) error {
	if row == nil {
		return fmt.Errorf("create %s: nil row", table)
	}
	if err := db.WithContext(ctx).Create(row).Error; err != nil {
		return fmt.Errorf("create %s: %w", table, err)
	}
	return nil
}

func getRow[T any](ctx context.Context, db *gorm.DB, table string, id int64) (*T, error) {
	if id == 0 {
		return nil, fmt.Errorf("get %s: id is required", table)
	}
	var row T
	if err := db.WithContext(ctx).First(&row, id).Error; err != nil {
		return nil, fmt.Errorf("get %s: %w", table, err)
	}
	return &row, nil
}

// updateRow 按主键整行写回。零值也会写入，调用前应先查出原记录再修改。
func updateRow[T any](ctx context.Context, db *gorm.DB, table string, id int64, row *T) error {
	if row == nil || id == 0 {
		return fmt.Errorf("update %s: id is required", table)
	}
	var n int64
	if err := db.WithContext(ctx).Model(new(T)).Where("id = ?", id).Count(&n).Error; err != nil {
		return fmt.Errorf("update %s: %w", table, err)
	}
	if n == 0 {
		return fmt.Errorf("update %s: %w", table, gorm.ErrRecordNotFound)
	}
	if err := db.WithContext(ctx).Save(row).Error; err != nil {
		return fmt.Errorf("update %s: %w", table, err)
	}
	return nil
}

func deleteRow[T any](ctx context.Context, db *gorm.DB, table string, id int64) error {
	if id == 0 {
		return fmt.Errorf("delete %s: id is required", table)
	}
	res := db.WithContext(ctx).Delete(new(T), id)
	if res.Error != nil {
		return fmt.Errorf("delete %s: %w", table, res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("delete %s: %w", table, gorm.ErrRecordNotFound)
	}
	return nil
}
