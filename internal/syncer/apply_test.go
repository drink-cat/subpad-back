package syncer

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/drink-cat/subpad-back/internal/model"
)

func TestEventTopics(t *testing.T) {
	checks := map[string]string{
		"TokenCreated": "0xab4d761add2b58732098ed4fb3b232cc510b86632778cec01eecd70984f6e9a5",
		"FeeCharged":   "0x14af2eea953773bfd62681db92f9600b65fd66ae0e85fbf49fd1d0cf034f6c50",
		"SwapOnce":     "0x4f87ea67c95796b9f57c6d3d3643dddd224b195ebd5833dbb093de07066b1a67",
	}
	for name, want := range checks {
		if launchABI.Events[name].ID != common.HexToHash(want) {
			t.Fatalf("%s topic = %s", name, launchABI.Events[name].ID.Hex())
		}
	}
}

func TestApplyTokenCreatedUpdatesPending(t *testing.T) {
	store := openStore(t)
	subpadID := int64(7)
	pending := &model.TokenInfo{
		UserID:      3,
		SubpadID:    &subpadID,
		ChainID:     31337,
		TokenSymbol: "AAA",
		TokenName:   "old",
	}
	if err := store.TokenInfo.Create(context.Background(), pending); err != nil {
		t.Fatal(err)
	}
	supply, _ := new(big.Int).SetString("1000000000000000000000000", 10)
	lg := packLog(t, "TokenCreated", 4, []common.Hash{
		common.HexToHash("0xabc"),
		common.BytesToHash(common.HexToAddress("0x1000000000000000000000000000000000000001").Bytes()),
		common.BytesToHash(common.HexToAddress("0x2000000000000000000000000000000000000002").Bytes()),
	}, "Demo", "AAA", common.HexToAddress("0x3000000000000000000000000000000000000003"), "USDC", supply, big.NewInt(0))

	if err := saveLog(context.Background(), store, 31337, lg); err != nil {
		t.Fatal(err)
	}
	if err := saveLog(context.Background(), store, 31337, lg); err != nil {
		t.Fatal(err)
	}
	rows, err := store.TokenInfo.List(context.Background(), model.TokenInfoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != pending.ID || rows[0].UserID != 3 || rows[0].SubpadID == nil || *rows[0].SubpadID != 7 {
		t.Fatalf("tokens = %+v", rows)
	}
	if rows[0].TokenAddr == "" || rows[0].PoolID == "" || rows[0].LaunchSupply.String() != supply.String() || rows[0].QuoteTokenSymbol != "USDC" {
		t.Fatalf("token = %+v", rows[0])
	}
}

func TestApplyTokenCreatedInsertsWhenMissing(t *testing.T) {
	store := openStore(t)
	lg := packLog(t, "TokenCreated", 1, []common.Hash{
		common.HexToHash("0x11"),
		common.BytesToHash(common.HexToAddress("0x1000000000000000000000000000000000000001").Bytes()),
		common.BytesToHash(common.HexToAddress("0x2000000000000000000000000000000000000002").Bytes()),
	}, "Demo", "BBB", common.HexToAddress("0x3000000000000000000000000000000000000003"), "USDC", big.NewInt(10), big.NewInt(0))
	if err := saveLog(context.Background(), store, 31337, lg); err != nil {
		t.Fatal(err)
	}
	rows, err := store.TokenInfo.List(context.Background(), model.TokenInfoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].UserID != 0 || rows[0].SubpadID != nil || rows[0].TokenSymbol != "BBB" {
		t.Fatalf("tokens = %+v", rows)
	}
}

func TestApplySwapAndFee(t *testing.T) {
	store := openStore(t)
	pool := common.HexToHash("0x22")
	trader := common.HexToAddress("0x4000000000000000000000000000000000000004")
	token := common.HexToAddress("0x2000000000000000000000000000000000000002")
	quote := common.HexToAddress("0x3000000000000000000000000000000000000003")
	feeTo := common.HexToAddress("0x5000000000000000000000000000000000000005")
	tx := common.HexToHash("0x33")

	swap := packLog(t, "SwapOnce", 2, []common.Hash{pool, common.BytesToHash(trader.Bytes()), common.BytesToHash(token.Bytes())},
		true, big.NewInt(1000), uint8(18), quote, big.NewInt(200), big.NewInt(2), uint8(6), big.NewInt(50))
	swap.TxHash = tx
	fee := packLog(t, "FeeCharged", 3, []common.Hash{pool, common.BytesToHash(quote.Bytes()), common.BytesToHash(feeTo.Bytes())},
		uint8(0), uint8(6), big.NewInt(2))
	fee.TxHash = tx
	if err := saveLog(context.Background(), store, 31337, swap); err != nil {
		t.Fatal(err)
	}
	if err := saveLog(context.Background(), store, 31337, fee); err != nil {
		t.Fatal(err)
	}
	swap.Data = mustRepack(t, "SwapOnce", true, big.NewInt(1000), uint8(18), quote, big.NewInt(250), big.NewInt(3), uint8(6), big.NewInt(50))
	if err := saveLog(context.Background(), store, 31337, swap); err != nil {
		t.Fatal(err)
	}
	if err := saveLog(context.Background(), store, 31337, fee); err != nil {
		t.Fatal(err)
	}

	swaps, err := store.SwapInfo.List(context.Background(), model.SwapInfoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(swaps) != 1 || !swaps[0].IsBuy || !swaps[0].QuoteAmount.Equal(250) || swaps[0].Trader != trader.Hex() {
		t.Fatalf("swaps = %+v", swaps)
	}
	fees, err := store.FeeInfo.List(context.Background(), model.FeeInfoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(fees) != 1 || fees[0].FeeType != model.FeeTypePlatform || !fees[0].FeeAmount.Equal(2) || fees[0].LogIndex != 3 {
		t.Fatalf("fees = %+v", fees)
	}
}

func TestRemovedAndUnknownLogsSkipBusinessTables(t *testing.T) {
	store := openStore(t)
	lg := packLog(t, "TokenCreated", 1, []common.Hash{
		common.HexToHash("0x11"),
		common.BytesToHash(common.HexToAddress("0x1000000000000000000000000000000000000001").Bytes()),
		common.BytesToHash(common.HexToAddress("0x2000000000000000000000000000000000000002").Bytes()),
	}, "Demo", "CCC", common.HexToAddress("0x3000000000000000000000000000000000000003"), "USDC", big.NewInt(10), big.NewInt(0))
	lg.Removed = true
	if err := saveLog(context.Background(), store, 31337, lg); err != nil {
		t.Fatal(err)
	}
	unknown := types.Log{Topics: []common.Hash{common.HexToHash("0xabc")}, TxHash: common.HexToHash("0x44"), Index: 1}
	if err := saveLog(context.Background(), store, 31337, unknown); err != nil {
		t.Fatal(err)
	}
	rows, err := store.TokenInfo.List(context.Background(), model.TokenInfoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 {
		t.Fatalf("tokens = %+v", rows)
	}
	events, err := store.SyncEvent.List(context.Background(), model.SyncEventFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %d", len(events))
	}
}

func packLog(t *testing.T, name string, index uint, indexed []common.Hash, args ...any) types.Log {
	t.Helper()
	ev := launchABI.Events[name]
	data := mustRepack(t, name, args...)
	topics := make([]common.Hash, 0, 1+len(indexed))
	topics = append(topics, ev.ID)
	topics = append(topics, indexed...)
	return types.Log{
		Address: common.HexToAddress("0x88D1aF96098a928eE278f162c1a84f339652f95b"),
		Topics:  topics,
		Data:    data,
		TxHash:  common.BigToHash(big.NewInt(int64(index) + 9)),
		Index:   index,
	}
}

func mustRepack(t *testing.T, name string, args ...any) []byte {
	t.Helper()
	data, err := launchABI.Events[name].Inputs.NonIndexed().Pack(args...)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
