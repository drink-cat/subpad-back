package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type SubpadInfo struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      int64     `gorm:"index:idx_subpad_info_user_id" json:"user_id"`
	UserAddr    string    `gorm:"size:100" json:"user_addr"`
	Brand       string    `gorm:"size:100;index:idx_subpad_info_brand" json:"brand"`
	NameFull    string    `gorm:"size:200" json:"name_full"`
	Status      int       `json:"status"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime" json:"updated_at"`
}

func (SubpadInfo) TableName() string { return "subpad_info" }

type SubpadInfoFilter struct {
	UserID *int64
	Brand  string
	Status *int
	Page
}

type SubpadInfoRepo struct {
	db *gorm.DB
}

func (r *SubpadInfoRepo) Create(ctx context.Context, row *SubpadInfo) error {
	return createRow(ctx, r.db, "subpad_info", row)
}

func (r *SubpadInfoRepo) Get(ctx context.Context, id int64) (*SubpadInfo, error) {
	return getRow[SubpadInfo](ctx, r.db, "subpad_info", id)
}

func (r *SubpadInfoRepo) GetByBrand(ctx context.Context, brand string) (*SubpadInfo, error) {
	if brand == "" {
		return nil, fmt.Errorf("get subpad_info: brand is required")
	}
	var row SubpadInfo
	err := r.db.WithContext(ctx).Where("brand = ?", brand).Order("id ASC").First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get subpad_info: %w", err)
	}
	return &row, nil
}

func (r *SubpadInfoRepo) Update(ctx context.Context, row *SubpadInfo) error {
	if row == nil {
		return fmt.Errorf("update subpad_info: nil row")
	}
	return updateRow(ctx, r.db, "subpad_info", row.ID, row)
}

func (r *SubpadInfoRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[SubpadInfo](ctx, r.db, "subpad_info", id)
}

func (r *SubpadInfoRepo) List(ctx context.Context, f SubpadInfoFilter) ([]SubpadInfo, error) {
	q := r.db.WithContext(ctx).Model(&SubpadInfo{})
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.Brand != "" {
		q = q.Where("brand = ?", f.Brand)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	rows := make([]SubpadInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list subpad_info: %w", err)
	}
	return rows, nil
}
