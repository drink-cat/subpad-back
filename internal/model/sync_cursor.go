package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SyncCursor struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChainID     int       `gorm:"column:chainid;uniqueIndex:uk_sync_cursor_chainid" json:"chainid"`
	BlockNumber int64     `json:"block_number"`
	CreatedAt   time.Time `gorm:"type:datetime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:datetime" json:"updated_at"`
}

func (SyncCursor) TableName() string { return "sync_cursor" }

type SyncCursorFilter struct {
	ChainID *int `form:"chainid"`
	Page
}

type SyncCursorRepo struct {
	db *gorm.DB
}

func (r *SyncCursorRepo) Create(ctx context.Context, row *SyncCursor) error {
	return createRow(ctx, r.db, "sync_cursor", row)
}

func (r *SyncCursorRepo) Get(ctx context.Context, id int64) (*SyncCursor, error) {
	return getRow[SyncCursor](ctx, r.db, "sync_cursor", id)
}

func (r *SyncCursorRepo) GetByChainID(ctx context.Context, chainID int) (*SyncCursor, error) {
	var row SyncCursor
	err := r.db.WithContext(ctx).Where("chainid = ?", chainID).First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get sync_cursor: %w", err)
	}
	return &row, nil
}

func (r *SyncCursorRepo) Update(ctx context.Context, row *SyncCursor) error {
	if row == nil {
		return fmt.Errorf("update sync_cursor: nil row")
	}
	return updateRow(ctx, r.db, "sync_cursor", row.ID, row)
}

func (r *SyncCursorRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[SyncCursor](ctx, r.db, "sync_cursor", id)
}

func (r *SyncCursorRepo) List(ctx context.Context, f SyncCursorFilter) ([]SyncCursor, error) {
	q := r.db.WithContext(ctx).Model(&SyncCursor{})
	if f.ChainID != nil {
		q = q.Where("chainid = ?", *f.ChainID)
	}
	rows := make([]SyncCursor, 0)
	if err := f.Apply(q.Order("chainid ASC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list sync_cursor: %w", err)
	}
	return rows, nil
}

// Upsert 按链写入已扫区块。同一条链再次写入时只更新区块高度。
func (r *SyncCursorRepo) Upsert(ctx context.Context, chainID int, blockNumber int64) error {
	row := &SyncCursor{ChainID: chainID, BlockNumber: blockNumber}
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "chainid"}},
		DoUpdates: clause.AssignmentColumns([]string{"block_number", "updated_at"}),
	}).Create(row).Error
	if err != nil {
		return fmt.Errorf("upsert sync_cursor: %w", err)
	}
	return nil
}
