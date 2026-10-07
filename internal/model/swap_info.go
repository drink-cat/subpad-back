package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type SwapInfo struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChainID        int       `gorm:"column:chainid;uniqueIndex:uk_swap_info_chain_tx_log,priority:1" json:"chainId"`
	PoolID         string    `gorm:"size:100;index:idx_swap_info_pool_id" json:"poolId"`
	TxHash         string    `gorm:"size:100;uniqueIndex:uk_swap_info_chain_tx_log,priority:2;index:idx_swap_info_tx_hash" json:"txHash"`
	LogIndex       int       `gorm:"uniqueIndex:uk_swap_info_chain_tx_log,priority:3" json:"logIndex"`
	Trader         string    `gorm:"size:100;index:idx_swap_info_trader" json:"trader"`
	IsBuy          bool      `json:"isBuy"`
	TokenAddr      string    `gorm:"size:100;index:idx_swap_info_token_addr" json:"tokenAddr"`
	TokenAmount    Amount    `gorm:"type:varchar(80)" json:"tokenAmount"`
	TokenDecimal   int       `json:"tokenDecimal"`
	QuoteTokenAddr string    `gorm:"size:100" json:"quoteTokenAddr"`
	QuoteAmount    Amount    `gorm:"type:varchar(80)" json:"quoteAmount"`
	Fee            Amount    `gorm:"type:varchar(80)" json:"fee"`
	QuoteDecimal   int       `json:"quoteDecimal"`
	Price          Amount    `gorm:"type:varchar(80)" json:"price"`
	CreatedAt      time.Time `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt      time.Time `gorm:"type:datetime" json:"updatedAt"`
}

func (SwapInfo) TableName() string { return "swap_info" }

type SwapInfoFilter struct {
	ChainID   *int   `form:"chainId"`
	PoolID    string `form:"poolId"`
	TxHash    string `form:"txHash"`
	Trader    string `form:"trader"`
	IsBuy     *bool  `form:"isBuy"`
	TokenAddr string `form:"tokenAddr"`
	Page
}

type SwapInfoRepo struct {
	db *gorm.DB
}

func (r *SwapInfoRepo) Create(ctx context.Context, row *SwapInfo) error {
	return createRow(ctx, r.db, "swap_info", row)
}

func (r *SwapInfoRepo) Get(ctx context.Context, id int64) (*SwapInfo, error) {
	return getRow[SwapInfo](ctx, r.db, "swap_info", id)
}

func (r *SwapInfoRepo) GetByLog(ctx context.Context, chainID int, txHash string, logIndex int) (*SwapInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("get swap_info: tx_hash is required")
	}
	var row SwapInfo
	err := r.db.WithContext(ctx).
		Where("chainid = ? AND tx_hash = ? AND log_index = ?", chainID, txHash, logIndex).
		First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get swap_info: %w", err)
	}
	return &row, nil
}

func (r *SwapInfoRepo) Update(ctx context.Context, row *SwapInfo) error {
	if row == nil {
		return fmt.Errorf("update swap_info: nil row")
	}
	return updateRow(ctx, r.db, "swap_info", row.ID, row)
}

func (r *SwapInfoRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[SwapInfo](ctx, r.db, "swap_info", id)
}

func (r *SwapInfoRepo) List(ctx context.Context, f SwapInfoFilter) ([]SwapInfo, error) {
	q := r.db.WithContext(ctx).Model(&SwapInfo{})
	if f.ChainID != nil {
		q = q.Where("chainid = ?", *f.ChainID)
	}
	if f.PoolID != "" {
		q = q.Where("pool_id = ?", f.PoolID)
	}
	if f.TxHash != "" {
		q = q.Where("tx_hash = ?", f.TxHash)
	}
	if f.Trader != "" {
		q = q.Where("trader = ?", f.Trader)
	}
	if f.IsBuy != nil {
		q = q.Where("is_buy = ?", *f.IsBuy)
	}
	if f.TokenAddr != "" {
		q = q.Where("token_addr = ?", f.TokenAddr)
	}
	rows := make([]SwapInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list swap_info: %w", err)
	}
	return rows, nil
}
