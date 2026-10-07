package syncer

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/model"
)

func TestValidRejectsShortContract(t *testing.T) {
	cfg := testCfg()
	cfg.LaunchContract = "0x111"
	if err := Valid(cfg); err == nil {
		t.Fatal("expected invalid contract")
	}
}

func TestScanTenBlocksFromBegin(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	addr := common.HexToAddress(cfg.LaunchContract)
	client := &stubClient{
		head: 800,
		logs: []types.Log{
			makeLog(addr, 705, 0, false),
			makeLog(addr, 714, 1, false),
			makeLog(addr, 715, 0, false),
		},
	}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	if len(client.queries) == 0 {
		t.Fatal("no queries")
	}
	first := client.queries[0]
	if first.FromBlock.Cmp(big.NewInt(705)) != 0 || first.ToBlock.Cmp(big.NewInt(714)) != 0 {
		t.Fatalf("first range = %s-%s", first.FromBlock, first.ToBlock)
	}
	if len(first.Addresses) != 1 || first.Addresses[0] != addr {
		t.Fatalf("addresses = %v", first.Addresses)
	}
	last := client.queries[len(client.queries)-1]
	if last.ToBlock.Cmp(big.NewInt(799)) != 0 {
		t.Fatalf("last to = %s", last.ToBlock)
	}
	for _, q := range client.queries {
		span := q.ToBlock.Int64() - q.FromBlock.Int64() + 1
		if span <= 0 || span > int64(BlocksPerScan) {
			t.Fatalf("batch span = %d", span)
		}
	}
	events, err := store.SyncEvent.List(context.Background(), model.SyncEventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 || events[0].BlockNumber != 715 || events[1].BlockNumber != 714 || events[2].BlockNumber != 705 {
		t.Fatalf("events = %+v", events)
	}
	if events[2].Topics == "" || events[2].Data != "0x0102" || events[2].Removed != model.SyncEventActive {
		t.Fatalf("event = %+v", events[2])
	}
	cursor, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 799 {
		t.Fatalf("cursor = %d", cursor.BlockNumber)
	}
}

func TestScanResumesAndSkipsDuplicate(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	addr := common.HexToAddress(cfg.LaunchContract)
	client := &stubClient{head: 800, logs: []types.Log{makeLog(addr, 705, 0, false)}}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	if err := store.SyncCursor.Upsert(context.Background(), int(cfg.ChainID), 704); err != nil {
		t.Fatal(err)
	}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	events, err := store.SyncEvent.List(context.Background(), model.SyncEventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}
	cursor, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 799 {
		t.Fatalf("cursor = %d", cursor.BlockNumber)
	}
}

func TestScanStopsAtSafeHead(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	client := &stubClient{head: 710}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	if client.last.FromBlock.Cmp(big.NewInt(705)) != 0 || client.last.ToBlock.Cmp(big.NewInt(709)) != 0 {
		t.Fatalf("range = %s-%s", client.last.FromBlock, client.last.ToBlock)
	}
	cursor, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 709 {
		t.Fatalf("cursor = %d", cursor.BlockNumber)
	}
}

func TestScanWaitsWhenHeadIsUnconfirmed(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	client := &stubClient{head: 705}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	if client.last.FromBlock != nil {
		t.Fatal("scanned before confirmation")
	}
	_, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cursor err = %v", err)
	}
}

func TestScanKeepsCursorWhenRPCFails(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	client := &stubClient{head: 800, fail: true}
	if err := Scan(context.Background(), client, store, cfg); err == nil {
		t.Fatal("expected rpc error")
	}
	_, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("cursor err = %v", err)
	}
}

func TestScanContinuesFromCursor(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	if err := store.SyncCursor.Upsert(context.Background(), int(cfg.ChainID), 714); err != nil {
		t.Fatal(err)
	}
	client := &stubClient{head: 800}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	if len(client.queries) == 0 {
		t.Fatal("no queries")
	}
	first := client.queries[0]
	if first.FromBlock.Cmp(big.NewInt(715)) != 0 || first.ToBlock.Cmp(big.NewInt(724)) != 0 {
		t.Fatalf("first range = %s-%s", first.FromBlock, first.ToBlock)
	}
	cursor, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 799 {
		t.Fatalf("cursor = %d", cursor.BlockNumber)
	}
}

func TestScanKeepsCursorWhenLaterBatchFails(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	client := &stubClient{head: 800, failAfter: 2}
	if err := Scan(context.Background(), client, store, cfg); err == nil {
		t.Fatal("expected rpc error")
	}
	cursor, err := store.SyncCursor.GetByChainID(context.Background(), int(cfg.ChainID))
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 714 {
		t.Fatalf("cursor = %d", cursor.BlockNumber)
	}
}

func TestScanStoresRemovedLog(t *testing.T) {
	store := openStore(t)
	cfg := testCfg()
	addr := common.HexToAddress(cfg.LaunchContract)
	client := &stubClient{head: 800, logs: []types.Log{makeLog(addr, 705, 2, true)}}
	if err := Scan(context.Background(), client, store, cfg); err != nil {
		t.Fatal(err)
	}
	events, err := store.SyncEvent.List(context.Background(), model.SyncEventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Removed != model.SyncEventRemoved || events[0].LogIndex != 2 {
		t.Fatalf("events = %+v", events)
	}
}

type stubClient struct {
	head      uint64
	logs      []types.Log
	queries   []ethereum.FilterQuery
	last      ethereum.FilterQuery
	fail      bool
	failAfter int
	calls     int
}

func (s *stubClient) BlockNumber(context.Context) (uint64, error) {
	return s.head, nil
}

func (s *stubClient) FilterLogs(_ context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	s.calls++
	s.queries = append(s.queries, q)
	s.last = q
	if s.fail || (s.failAfter > 0 && s.calls >= s.failAfter) {
		return nil, errors.New("rpc")
	}
	from := q.FromBlock.Uint64()
	to := q.ToBlock.Uint64()
	out := make([]types.Log, 0)
	for _, lg := range s.logs {
		if lg.BlockNumber < from || lg.BlockNumber > to {
			continue
		}
		if len(q.Addresses) > 0 && lg.Address != q.Addresses[0] {
			continue
		}
		out = append(out, lg)
	}
	return out, nil
}

func testCfg() config.SyncLogConfig {
	return config.SyncLogConfig{
		Name:                 "本地网",
		ChainID:              31337,
		RPCURL:               "http://127.0.0.1:8545",
		BeginBlock:           705,
		Confirmations:        1,
		TimerIntervalSeconds: 60,
		LaunchContract:       "0x88D1aF96098a928eE278f162c1a84f339652f95b",
	}
}

func makeLog(addr common.Address, block uint64, index uint, removed bool) types.Log {
	return types.Log{
		Address:     addr,
		Topics:      []common.Hash{common.HexToHash("0xabc")},
		Data:        []byte{1, 2},
		BlockNumber: block,
		TxHash:      common.BigToHash(new(big.Int).SetUint64(block)),
		TxIndex:     1,
		BlockHash:   common.BigToHash(new(big.Int).SetUint64(block + 1000)),
		Index:       index,
		Removed:     removed,
	}
}

func openStore(t *testing.T) *model.Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err = model.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return model.NewStore(db)
}
