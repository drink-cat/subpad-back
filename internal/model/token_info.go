package model

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type TokenInfo struct {
	ID               int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	SubpadID         *int64 `gorm:"index:idx_token_info_subpad_id" json:"subpad_id"`
	PoolID           string `gorm:"size:100;index:idx_token_info_pool_id" json:"pool_id"`
	Creator          string `gorm:"size:100;index:idx_token_info_creator" json:"creator"`
	ChainID          int    `gorm:"column:chainid" json:"chainid"`
	TokenAddr        string `gorm:"size:100;index:idx_token_info_token_addr" json:"token_addr"`
	TokenName        string `gorm:"size:200" json:"token_name"`
	TokenSymbol      string `gorm:"size:64" json:"token_symbol"`
	QuoteTokenAddr   string `gorm:"size:100" json:"quote_token_addr"`
	QuoteTokenSymbol string `gorm:"size:64" json:"quote_token_symbol"`
	LaunchSupply     int64  `json:"launch_supply"`
	TickSpacing      int    `json:"tick_spacing"`
}

func (TokenInfo) TableName() string { return "token_info" }

type TokenInfoFilter struct {
	SubpadID  *int64
	PoolID    string
	Creator   string
	ChainID   *int
	TokenAddr string
	Page
}

type TokenInfoRepo struct {
	db *gorm.DB
}

func (r *TokenInfoRepo) Create(ctx context.Context, row *TokenInfo) error {
	return createRow(ctx, r.db, "token_info", row)
}

func (r *TokenInfoRepo) Get(ctx context.Context, id int64) (*TokenInfo, error) {
	return getRow[TokenInfo](ctx, r.db, "token_info", id)
}

func (r *TokenInfoRepo) Update(ctx context.Context, row *TokenInfo) error {
	if row == nil {
		return fmt.Errorf("update token_info: nil row")
	}
	return updateRow(ctx, r.db, "token_info", row.ID, row)
}

func (r *TokenInfoRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[TokenInfo](ctx, r.db, "token_info", id)
}

func (r *TokenInfoRepo) List(ctx context.Context, f TokenInfoFilter) ([]TokenInfo, error) {
	q := r.db.WithContext(ctx).Model(&TokenInfo{})
	if f.SubpadID != nil {
		q = q.Where("subpad_id = ?", *f.SubpadID)
	}
	if f.PoolID != "" {
		q = q.Where("pool_id = ?", f.PoolID)
	}
	if f.Creator != "" {
		q = q.Where("creator = ?", f.Creator)
	}
	if f.ChainID != nil {
		q = q.Where("chainid = ?", *f.ChainID)
	}
	if f.TokenAddr != "" {
		q = q.Where("token_addr = ?", f.TokenAddr)
	}
	rows := make([]TokenInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list token_info: %w", err)
	}
	return rows, nil
}
