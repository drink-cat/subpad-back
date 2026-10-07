package model

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

const (
	SyncEventActive  = 0
	SyncEventRemoved = 1
)

type SyncEvent struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ChainID      int       `gorm:"column:chainid;uniqueIndex:uk_sync_event_chain_tx_log,priority:1" json:"chainId"`
	BlockNumber  int64     `json:"blockNumber"`
	BlockHash    string    `gorm:"size:100" json:"blockHash"`
	TxHash       string    `gorm:"size:100;uniqueIndex:uk_sync_event_chain_tx_log,priority:2" json:"txHash"`
	TxIndex      int       `json:"txIndex"`
	LogIndex     int       `gorm:"uniqueIndex:uk_sync_event_chain_tx_log,priority:3" json:"logIndex"`
	ContractAddr string    `gorm:"size:100;index:idx_sync_event_contract_addr" json:"contractAddr"`
	Topics       string    `gorm:"type:text" json:"topics"`
	Data         string    `gorm:"type:mediumtext" json:"data"`
	Removed      int       `gorm:"default:0" json:"removed"`
	CreatedAt    time.Time `gorm:"type:datetime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"type:datetime" json:"updatedAt"`
}

func (SyncEvent) TableName() string { return "sync_event" }

type SyncEventFilter struct {
	ChainID      *int   `form:"chainId"`
	TxHash       string `form:"txHash"`
	ContractAddr string `form:"contractAddr"`
	Removed      *int   `form:"removed"`
	Page
}

type SyncEventRepo struct {
	db *gorm.DB
}

func (r *SyncEventRepo) Create(ctx context.Context, row *SyncEvent) error {
	return createRow(ctx, r.db, "sync_event", row)
}

func (r *SyncEventRepo) Get(ctx context.Context, id int64) (*SyncEvent, error) {
	return getRow[SyncEvent](ctx, r.db, "sync_event", id)
}

func (r *SyncEventRepo) GetByLog(ctx context.Context, chainID int, txHash string, logIndex int) (*SyncEvent, error) {
	if txHash == "" {
		return nil, fmt.Errorf("get sync_event: tx_hash is required")
	}
	var row SyncEvent
	err := r.db.WithContext(ctx).
		Where("chainid = ? AND tx_hash = ? AND log_index = ?", chainID, txHash, logIndex).
		First(&row).Error
	if err != nil {
		return nil, fmt.Errorf("get sync_event: %w", err)
	}
	return &row, nil
}

func (r *SyncEventRepo) Update(ctx context.Context, row *SyncEvent) error {
	if row == nil {
		return fmt.Errorf("update sync_event: nil row")
	}
	return updateRow(ctx, r.db, "sync_event", row.ID, row)
}

func (r *SyncEventRepo) Delete(ctx context.Context, id int64) error {
	return deleteRow[SyncEvent](ctx, r.db, "sync_event", id)
}

func (r *SyncEventRepo) List(ctx context.Context, f SyncEventFilter) ([]SyncEvent, error) {
	q := r.db.WithContext(ctx).Model(&SyncEvent{})
	if f.ChainID != nil {
		q = q.Where("chainid = ?", *f.ChainID)
	}
	if f.TxHash != "" {
		q = q.Where("tx_hash = ?", f.TxHash)
	}
	if f.ContractAddr != "" {
		q = q.Where("contract_addr = ?", f.ContractAddr)
	}
	if f.Removed != nil {
		q = q.Where("removed = ?", *f.Removed)
	}
	rows := make([]SyncEvent, 0)
	if err := f.Apply(q.Order("block_number DESC, log_index DESC")).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("list sync_event: %w", err)
	}
	return rows, nil
}
