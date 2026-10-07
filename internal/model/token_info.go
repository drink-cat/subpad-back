package model

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type TokenInfo struct {
	ID               int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           int64  `gorm:"index:idx_token_info_user_id" json:"userId"`
	SubpadID         *int64 `gorm:"index:idx_token_info_subpad_id" json:"subpadId"`
	PoolID           string `gorm:"size:100;index:idx_token_info_pool_id" json:"poolId"`
	Creator          string `gorm:"size:100;index:idx_token_info_creator" json:"creator"`
	ChainID          int    `gorm:"column:chainid" json:"chainId"`
	TokenAddr        string `gorm:"size:100;index:idx_token_info_token_addr" json:"tokenAddr"`
	TokenName        string `gorm:"size:200" json:"tokenName"`
	TokenSymbol      string `gorm:"size:64;index:idx_token_info_token_symbol" json:"tokenSymbol"`
	QuoteTokenAddr   string `gorm:"size:100" json:"quoteTokenAddr"`
	QuoteTokenSymbol string `gorm:"size:64" json:"quoteTokenSymbol"`
	LaunchSupply     Amount `gorm:"type:varchar(80)" json:"launchSupply"`
	TickSpacing      int    `json:"tickSpacing"`
}

func (TokenInfo) TableName() string { return "token_info" }

type TokenInfoFilter struct {
	UserID      *int64 `form:"userId"`
	SubpadID    *int64 `form:"subpadId"`
	PoolID      string `form:"poolId"`
	Creator     string `form:"creator"`
	ChainID     *int   `form:"chainId"`
	TokenAddr   string `form:"tokenAddr"`
	TokenSymbol string `form:"tokenSymbol"`
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
	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
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
	if f.TokenSymbol != "" {
		q = q.Where("token_symbol = ?", f.TokenSymbol)
	}
	rows := make([]TokenInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list token_info: %w", err)
	}
	return rows, nil
}

func (r *TokenInfoRepo) GetByPool(ctx context.Context, chainID int, poolID string) (*TokenInfo, error) {
	if poolID == "" {
		return nil, fmt.Errorf("get token_info: pool_id is required")
	}
	var row TokenInfo
	err := r.db.WithContext(ctx).Where("chainid = ? AND pool_id = ?", chainID, poolID).First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get token_info: %w", err)
	}
	return &row, nil
}

// FindPending 找这条链上同符号、还没写上代币地址的最新一行。发币接口会先落库，日志到达后再补齐。
func (r *TokenInfoRepo) FindPending(ctx context.Context, chainID int, symbol string) (*TokenInfo, error) {
	if symbol == "" {
		return nil, nil
	}
	var row TokenInfo
	err := r.db.WithContext(ctx).
		Where("chainid = ? AND token_symbol = ? AND token_addr = ?", chainID, symbol, "").
		Order("id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find token_info: %w", err)
	}
	return &row, nil
}
