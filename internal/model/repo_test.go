package model

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUserInfoBeforeSaveHashesPassword(t *testing.T) {
	u := &UserInfo{Password: "secret"}
	if err := u.BeforeSave(nil); err != nil {
		t.Fatal(err)
	}
	if u.Password == "secret" {
		t.Fatal("password stored in plaintext")
	}
	if err := u.CheckPassword("secret"); err != nil {
		t.Fatal(err)
	}
	hashed := u.Password
	if err := u.BeforeSave(nil); err != nil {
		t.Fatal(err)
	}
	if u.Password != hashed {
		t.Fatal("bcrypt hash was hashed again")
	}
}

func TestStoreCRUD(t *testing.T) {
	db := openTestDB(t)
	store := NewStore(db)
	ctx := context.Background()

	user := &UserInfo{Username: "alice", Password: "secret", FeeAddr: "0xfee"}
	if err := store.UserInfo.Create(ctx, user); err != nil {
		t.Fatal(err)
	}
	if err := user.CheckPassword("secret"); err != nil {
		t.Fatal(err)
	}
	gotUser, err := store.UserInfo.GetByUsername(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	gotUser.FeeAddr = "0xnew"
	if err = store.UserInfo.Update(ctx, gotUser); err != nil {
		t.Fatal(err)
	}
	gotUser, err = store.UserInfo.Get(ctx, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if gotUser.FeeAddr != "0xnew" {
		t.Fatalf("fee_addr = %s", gotUser.FeeAddr)
	}
	if err = gotUser.CheckPassword("secret"); err != nil {
		t.Fatal(err)
	}
	users, err := store.UserInfo.List(ctx, UserInfoFilter{Username: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 {
		t.Fatalf("users = %d", len(users))
	}
	if err = store.UserInfo.Delete(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	_, err = store.UserInfo.Get(ctx, user.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("get deleted user: %v", err)
	}

	subpadID := int64(7)
	token := &TokenInfo{
		UserID:      3,
		SubpadID:    &subpadID,
		PoolID:      "pool-1",
		Creator:     "0xcreator",
		ChainID:     1,
		TokenSymbol: "AAA",
	}
	if err = store.TokenInfo.Create(ctx, token); err != nil {
		t.Fatal(err)
	}
	chainID := 1
	userID := int64(3)
	tokens, err := store.TokenInfo.List(ctx, TokenInfoFilter{ChainID: &chainID, PoolID: "pool-1", UserID: &userID, TokenSymbol: "AAA"})
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 || tokens[0].ChainID != 1 || tokens[0].UserID != 3 || tokens[0].TokenAddr != "" || tokens[0].SubpadID == nil || *tokens[0].SubpadID != 7 {
		t.Fatalf("tokens = %+v", tokens)
	}
	tokens[0].TokenAddr = "0xtoken"
	tokens[0].PoolID = "pool-1"
	if err = store.TokenInfo.Update(ctx, &tokens[0]); err != nil {
		t.Fatal(err)
	}
	filled, err := store.TokenInfo.Get(ctx, token.ID)
	if err != nil {
		t.Fatal(err)
	}
	if filled.TokenAddr != "0xtoken" || filled.UserID != 3 {
		t.Fatalf("token = %+v", filled)
	}
	empty := &TokenInfo{PoolID: "pool-empty"}
	if err = store.TokenInfo.Create(ctx, empty); err != nil {
		t.Fatal(err)
	}
	empty, err = store.TokenInfo.Get(ctx, empty.ID)
	if err != nil {
		t.Fatal(err)
	}
	if empty.SubpadID != nil {
		t.Fatalf("subpad_id = %v", *empty.SubpadID)
	}

	subpad := &SubpadInfo{UserID: 3, Brand: "demo", NameFull: "Demo Pad", Status: 1, SwapType: SwapTypeMock}
	if err = store.SubpadInfo.Create(ctx, subpad); err != nil {
		t.Fatal(err)
	}
	if subpad.CreatedAt.IsZero() || subpad.UpdatedAt.IsZero() {
		t.Fatal("timestamps were not set")
	}
	byBrand, err := store.SubpadInfo.GetByBrand(ctx, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if byBrand.NameFull != "Demo Pad" || byBrand.SwapType != SwapTypeMock {
		t.Fatalf("subpad = %+v", byBrand)
	}

	fee := &FeeInfo{
		ChainID:    1,
		PoolID:     "pool-1",
		TxHash:     "0xtx",
		FeeType:    FeeTypePlatform,
		FeeToken:   "0xusdc",
		FeeDecimal: 6,
		FeeAmount:  NewAmount(100),
		FeeTo:      "0xfee",
	}
	if err = store.FeeInfo.Create(ctx, fee); err != nil {
		t.Fatal(err)
	}
	fees, err := store.FeeInfo.List(ctx, FeeInfoFilter{FeeTo: "0xfee", FeeType: FeeTypePlatform})
	if err != nil {
		t.Fatal(err)
	}
	if len(fees) != 1 || !fees[0].FeeAmount.Equal(100) {
		t.Fatalf("fees = %+v", fees)
	}

	buy := true
	swap := &SwapInfo{
		ChainID:        1,
		PoolID:         "pool-1",
		TxHash:         "0xswap",
		LogIndex:       3,
		Trader:         "0xtrader",
		IsBuy:          true,
		TokenAddr:      "0xtoken",
		TokenAmount:    NewAmount(1000),
		TokenDecimal:   18,
		QuoteTokenAddr: "0xusdc",
		QuoteAmount:    NewAmount(200),
		Fee:            NewAmount(2),
		QuoteDecimal:   6,
		Price:          NewAmount(50),
	}
	if err = store.SwapInfo.Create(ctx, swap); err != nil {
		t.Fatal(err)
	}
	dupSwap := &SwapInfo{ChainID: 1, TxHash: "0xswap", LogIndex: 3}
	if err = store.SwapInfo.Create(ctx, dupSwap); err == nil {
		t.Fatal("duplicate swap_info was inserted")
	}
	swaps, err := store.SwapInfo.List(ctx, SwapInfoFilter{Trader: "0xtrader", IsBuy: &buy, TokenAddr: "0xtoken"})
	if err != nil {
		t.Fatal(err)
	}
	if len(swaps) != 1 || !swaps[0].QuoteAmount.Equal(200) || !swaps[0].Fee.Equal(2) || !swaps[0].IsBuy {
		t.Fatalf("swaps = %+v", swaps)
	}

	event := &SyncEvent{
		ChainID:      1,
		BlockNumber:  10,
		TxHash:       "0xhash",
		LogIndex:     2,
		ContractAddr: "0xcontract",
		Topics:       "0xtopic",
		Data:         "0xdata",
	}
	if err = store.SyncEvent.Create(ctx, event); err != nil {
		t.Fatal(err)
	}
	dup := &SyncEvent{ChainID: 1, TxHash: "0xhash", LogIndex: 2}
	if err = store.SyncEvent.Create(ctx, dup); err == nil {
		t.Fatal("duplicate sync_event was inserted")
	}
	byLog, err := store.SyncEvent.GetByLog(ctx, 1, "0xhash", 2)
	if err != nil {
		t.Fatal(err)
	}
	if byLog.Removed != SyncEventActive {
		t.Fatalf("removed = %d", byLog.Removed)
	}
	byLog.Removed = SyncEventRemoved
	if err = store.SyncEvent.Update(ctx, byLog); err != nil {
		t.Fatal(err)
	}
	removed := SyncEventRemoved
	events, err := store.SyncEvent.List(ctx, SyncEventFilter{ContractAddr: "0xcontract", Removed: &removed})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %d", len(events))
	}

	if err = store.SyncCursor.Upsert(ctx, 1, 100); err != nil {
		t.Fatal(err)
	}
	if err = store.SyncCursor.Upsert(ctx, 1, 250); err != nil {
		t.Fatal(err)
	}
	cursor, err := store.SyncCursor.GetByChainID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if cursor.BlockNumber != 250 {
		t.Fatalf("block_number = %d", cursor.BlockNumber)
	}
	cursors, err := store.SyncCursor.List(ctx, SyncCursorFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(cursors) != 1 {
		t.Fatalf("cursors = %d", len(cursors))
	}
	if err = store.SyncCursor.Delete(ctx, cursor.ID); err != nil {
		t.Fatal(err)
	}
	if err = store.TokenInfo.Update(ctx, &TokenInfo{ID: 99999, PoolID: "missing"}); err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("update missing token: %v", err)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
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
	if err = AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}
