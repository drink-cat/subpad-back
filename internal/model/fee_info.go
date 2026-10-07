package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	FeeTypePlatform     = "platform"
	FeeTypeTokenCreator = "tokencreator"
	FeeTypeSubpad       = "subpad"
)

type FeeInfo struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChainID    int       `gorm:"column:chainid;uniqueIndex:uk_fee_info_chain_tx_log,priority:1" json:"chainId"`
	PoolID     string    `gorm:"size:100;index:idx_fee_info_pool_id" json:"poolId"`
	TxHash     string    `gorm:"size:100;uniqueIndex:uk_fee_info_chain_tx_log,priority:2;index:idx_fee_info_tx_hash" json:"txHash"`
	LogIndex   int       `gorm:"uniqueIndex:uk_fee_info_chain_tx_log,priority:3" json:"logIndex"`
	FeeType    string    `gorm:"size:32" json:"feeType"`
	FeeToken   string    `gorm:"size:100" json:"feeToken"`
	FeeDecimal int       `json:"feeDecimal"`
	FeeAmount  Amount    `gorm:"type:varchar(80)" json:"feeAmount"`
	FeeTo      string    `gorm:"size:100;index:idx_fee_info_fee_to" json:"feeTo"`
	CreatedAt  time.Time `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt  time.Time `gorm:"type:datetime" json:"updatedAt"`
}

func (FeeInfo) TableName() string { return "fee_info" }

type FeeInfoFilter struct {
	ChainID *int   `form:"chainId"`
	PoolID  string `form:"poolId"`
	TxHash  string `form:"txHash"`
	FeeType string `form:"feeType"`
	FeeTo   string `form:"feeTo"`
	Page
}

type FeeInfoRepo struct {
	db *gorm.DB
}

func (r *FeeInfoRepo) Create(ctx context.Context, row *FeeInfo) error {
	return createRow(ctx, r.db, "fee_info", row)
}

func (r *FeeInfoRepo) Get(ctx context.Context, id int64) (*FeeInfo, error) {
	return getRow[FeeInfo](ctx, r.db, "fee_info", id)
}

func (r *FeeInfoRepo) GetByLog(ctx context.Context, chainID int, txHash string, logIndex int) (*FeeInfo, error) {
	if txHash == "" {
		return nil, fmt.Errorf("get fee_info: tx_hash is required")
	}
	var row FeeInfo
	err := r.db.WithContext(ctx).
		Where("chainid = ? AND tx_hash = ? AND log_index = ?", chainID, txHash, logIndex).
		First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get fee_info: %w", err)
	}
	return &row, nil
}

func (r *FeeInfoRepo) Update(ctx context.Context, row *FeeInfo) error {
	if row == nil {
		return fmt.Errorf("update fee_info: nil row")
	}
	return updateRow(ctx, r.db, "fee_info", row.ID, row)
}

func (r *FeeInfoRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[FeeInfo](ctx, r.db, "fee_info", id)
}

func (r *FeeInfoRepo) List(ctx context.Context, f FeeInfoFilter) ([]FeeInfo, error) {
	q := r.db.WithContext(ctx).Model(&FeeInfo{})
	if f.ChainID != nil {
		q = q.Where("chainid = ?", *f.ChainID)
	}
	if f.PoolID != "" {
		q = q.Where("pool_id = ?", f.PoolID)
	}
	if f.TxHash != "" {
		q = q.Where("tx_hash = ?", f.TxHash)
	}
	if f.FeeType != "" {
		q = q.Where("fee_type = ?", f.FeeType)
	}
	if f.FeeTo != "" {
		q = q.Where("fee_to = ?", f.FeeTo)
	}
	rows := make([]FeeInfo, 0)
	if err := f.Apply(q.Order("id DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list fee_info: %w", err)
	}
	return rows, nil
}
