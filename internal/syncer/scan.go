package syncer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/model"
)

// BlocksPerScan 是定时器每次向前扫的区块数。
const BlocksPerScan = 10

type logClient interface {
	BlockNumber(ctx context.Context) (uint64, error)
	FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error)
}

// Chain 按配置扫一条链上的 LaunchCore 日志。
type Chain struct {
	Cfg   config.SyncLogConfig
	Store *model.Store

	mu  sync.Mutex
	eth *ethclient.Client
}

// Valid 检查这条链能不能扫。合约地址不完整时返回错误。
func Valid(cfg config.SyncLogConfig) error {
	if cfg.ChainID == 0 {
		return fmt.Errorf("chain id is empty")
	}
	if cfg.RPCURL == "" {
		return fmt.Errorf("rpc url is empty")
	}
	if !common.IsHexAddress(cfg.LaunchContract) {
		return fmt.Errorf("launch contract %s is invalid", cfg.LaunchContract)
	}
	return nil
}

func (c *Chain) Run(ctx context.Context) error {
	client, err := c.dial(ctx)
	if err != nil {
		return err
	}
	return Scan(ctx, client, c.Store, c.Cfg)
}

func (c *Chain) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.eth != nil {
		c.eth.Close()
		c.eth = nil
	}
}

func (c *Chain) dial(ctx context.Context) (*ethclient.Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.eth != nil {
		return c.eth, nil
	}
	client, err := ethclient.DialContext(ctx, c.Cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", c.Cfg.Name, err)
	}
	id, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("chain id %s: %w", c.Cfg.Name, err)
	}
	if id.Int64() != c.Cfg.ChainID {
		client.Close()
		return nil, fmt.Errorf("chain id %s: rpc %d, config %d", c.Cfg.Name, id.Int64(), c.Cfg.ChainID)
	}
	c.eth = client
	return c.eth, nil
}

// Scan 从下标的下一块开始，最多扫 BlocksPerScan 个已确认区块。
// 确认高度 = 链头 - confirmations。没有新块时不写库。
func Scan(ctx context.Context, client logClient, store *model.Store, cfg config.SyncLogConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if store == nil {
		return fmt.Errorf("scan %s: mysql is not configured", cfg.Name)
	}
	if err := Valid(cfg); err != nil {
		return fmt.Errorf("scan %s: %w", cfg.Name, err)
	}

	head, err := client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("scan %s block number: %w", cfg.Name, err)
	}
	chainID := int(cfg.ChainID)
	cursor, err := store.SyncCursor.GetByChainID(ctx, chainID)
	hasCursor := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("scan %s cursor: %w", cfg.Name, err)
	}
	var cursorBlock int64
	if hasCursor {
		cursorBlock = cursor.BlockNumber
	}
	from, to, ok := scanRange(hasCursor, cursorBlock, cfg.BeginBlock, head, cfg.Confirmations, BlocksPerScan)
	if !ok {
		return nil
	}

	logs, err := client.FilterLogs(ctx, ethereum.FilterQuery{
		FromBlock: bigInt(from),
		ToBlock:   bigInt(to),
		Addresses: []common.Address{common.HexToAddress(cfg.LaunchContract)},
	})
	if err != nil {
		return fmt.Errorf("scan %s logs %d-%d: %w", cfg.Name, from, to, err)
	}
	for _, lg := range logs {
		if err = saveLog(ctx, store, chainID, lg); err != nil {
			return fmt.Errorf("scan %s save log: %w", cfg.Name, err)
		}
	}
	if err = store.SyncCursor.Upsert(ctx, chainID, to); err != nil {
		return fmt.Errorf("scan %s cursor: %w", cfg.Name, err)
	}
	slog.Info("scanned blocks", "name", cfg.Name, "chainId", chainID, "from", from, "to", to, "logs", len(logs))
	return nil
}

func saveLog(ctx context.Context, store *model.Store, chainID int, lg types.Log) error {
	if !lg.Removed {
		if err := applyLog(ctx, store, chainID, lg); err != nil {
			return err
		}
	}
	txHash := lg.TxHash.Hex()
	logIndex := int(lg.Index)
	_, err := store.SyncEvent.GetByLog(ctx, chainID, txHash, logIndex)
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	removed := model.SyncEventActive
	if lg.Removed {
		removed = model.SyncEventRemoved
	}
	return store.SyncEvent.Create(ctx, &model.SyncEvent{
		ChainID:      chainID,
		BlockNumber:  int64(lg.BlockNumber),
		BlockHash:    lg.BlockHash.Hex(),
		TxHash:       txHash,
		TxIndex:      int(lg.TxIndex),
		LogIndex:     logIndex,
		ContractAddr: lg.Address.Hex(),
		Topics:       encodeTopics(lg.Topics),
		Data:         hexutil.Encode(lg.Data),
		Removed:      removed,
	})
}

func bigInt(n int64) *big.Int {
	return big.NewInt(n)
}

func encodeTopics(topics []common.Hash) string {
	if len(topics) == 0 {
		return ""
	}
	parts := make([]string, len(topics))
	for i, topic := range topics {
		parts[i] = topic.Hex()
	}
	return strings.Join(parts, ",")
}

// scanRange 计算这一轮要扫的闭区间。确认数大于链头时还不能扫。
func scanRange(hasCursor bool, cursor, begin int64, head uint64, confirmations, batch int) (from, to int64, ok bool) {
	if batch <= 0 {
		batch = BlocksPerScan
	}
	if confirmations < 0 {
		confirmations = 0
	}
	if uint64(confirmations) > head {
		return 0, 0, false
	}
	safe := int64(head - uint64(confirmations))
	if hasCursor {
		from = cursor + 1
	} else {
		from = begin
	}
	if from > safe {
		return 0, 0, false
	}
	to = from + int64(batch) - 1
	if to > safe {
		to = safe
	}
	return from, to, true
}
