package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/model"
	"github.com/drink-cat/subpad-back/internal/svc"
)

func TestAPIWithoutMySQL(t *testing.T) {
	r := testRouter(t, nil)
	rec := perform(r, http.MethodGet, "/api/user_info/1", nil)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	body := decodeResp(t, rec)
	if body.Code != codeUnavailable {
		t.Fatalf("code = %d", body.Code)
	}
}

func TestUserAndRelatedAPI(t *testing.T) {
	db := testDB(t)
	store := model.NewStore(db)
	r := testRouter(t, store)

	rec := perform(r, http.MethodPost, "/api/user_info", map[string]any{
		"username": "alice",
		"password": "secret",
		"fee_addr": "0xfee",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d body = %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret") || strings.Contains(rec.Body.String(), "$2a$") {
		t.Fatalf("password leaked: %s", rec.Body.String())
	}
	created := decodeResp(t, rec)
	if created.Code != codeOK {
		t.Fatalf("code = %d message = %s", created.Code, created.Message)
	}
	var user model.UserInfo
	if err := json.Unmarshal(created.Data, &user); err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 || user.FeeAddr != "0xfee" {
		t.Fatalf("user = %+v", user)
	}

	rec = perform(r, http.MethodGet, "/api/user_info/"+itoa(user.ID), nil)
	got := decodeData[model.UserInfo](t, rec)
	if got.Username != "alice" {
		t.Fatalf("username = %s", got.Username)
	}

	rec = perform(r, http.MethodPut, "/api/user_info/"+itoa(user.ID), map[string]any{
		"username": "alice",
		"fee_addr": "0xnew",
	})
	updated := decodeData[model.UserInfo](t, rec)
	if updated.FeeAddr != "0xnew" {
		t.Fatalf("fee_addr = %s", updated.FeeAddr)
	}
	stored, err := store.UserInfo.Get(t.Context(), user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = stored.CheckPassword("secret"); err != nil {
		t.Fatal(err)
	}

	rec = perform(r, http.MethodGet, "/api/user_info?username=alice", nil)
	users := decodeData[[]model.UserInfo](t, rec)
	if len(users) != 1 {
		t.Fatalf("users = %d", len(users))
	}

	rec = perform(r, http.MethodPost, "/api/user_info", map[string]any{"username": "bob"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing password status = %d body = %s", rec.Code, rec.Body.String())
	}

	subpadID := int64(7)
	rec = perform(r, http.MethodPost, "/api/token_info", model.TokenInfo{
		SubpadID:  &subpadID,
		PoolID:    "pool-1",
		Creator:   "0xcreator",
		ChainID:   1,
		TokenAddr: "0xtoken",
	})
	token := decodeData[model.TokenInfo](t, rec)
	rec = perform(r, http.MethodGet, "/api/token_info?pool_id=pool-1&chainid=1", nil)
	tokens := decodeData[[]model.TokenInfo](t, rec)
	if len(tokens) != 1 || tokens[0].ID != token.ID || tokens[0].SubpadID == nil || *tokens[0].SubpadID != 7 {
		t.Fatalf("tokens = %+v", tokens)
	}

	rec = perform(r, http.MethodPost, "/api/subpad_info", model.SubpadInfo{
		UserID:   3,
		Brand:    "demo",
		NameFull: "Demo Pad",
		Status:   1,
		SwapType: model.SwapTypeMock,
	})
	subpad := decodeData[model.SubpadInfo](t, rec)
	if subpad.CreatedAt.IsZero() || subpad.SwapType != model.SwapTypeMock {
		t.Fatalf("subpad = %+v", subpad)
	}
	rec = perform(r, http.MethodGet, "/api/subpad_info?swap_type="+model.SwapTypeMock, nil)
	subpads := decodeData[[]model.SubpadInfo](t, rec)
	if len(subpads) != 1 || subpads[0].ID != subpad.ID {
		t.Fatalf("subpads = %+v", subpads)
	}
	rec = perform(r, http.MethodPut, "/api/subpad_info/"+itoa(subpad.ID), model.SubpadInfo{
		UserID:      3,
		Brand:       "demo",
		NameFull:    "Demo Pad 2",
		Status:      1,
		SwapType:    model.SwapTypeUni,
		Description: "detail",
	})
	changed := decodeData[model.SubpadInfo](t, rec)
	if !changed.CreatedAt.Equal(subpad.CreatedAt) {
		t.Fatalf("created_at changed from %s to %s", subpad.CreatedAt, changed.CreatedAt)
	}
	if changed.NameFull != "Demo Pad 2" || changed.SwapType != model.SwapTypeUni {
		t.Fatalf("subpad = %+v", changed)
	}

	rec = perform(r, http.MethodPost, "/api/fee_info", model.FeeInfo{
		ChainID:    1,
		PoolID:     "pool-1",
		TxHash:     "0xtx",
		FeeType:    model.FeeTypePlatform,
		FeeDecimal: 6,
		FeeAmount:  100,
		FeeTo:      "0xfee",
	})
	fee := decodeData[model.FeeInfo](t, rec)
	rec = perform(r, http.MethodGet, "/api/fee_info?fee_to=0xfee&fee_type=platform", nil)
	fees := decodeData[[]model.FeeInfo](t, rec)
	if len(fees) != 1 || fees[0].ID != fee.ID || fees[0].FeeAmount != 100 {
		t.Fatalf("fees = %+v", fees)
	}

	rec = perform(r, http.MethodPost, "/api/sync_event", model.SyncEvent{
		ChainID:      1,
		BlockNumber:  10,
		TxHash:       "0xhash",
		LogIndex:     2,
		ContractAddr: "0xcontract",
	})
	event := decodeData[model.SyncEvent](t, rec)
	rec = perform(r, http.MethodPost, "/api/sync_event", model.SyncEvent{
		ChainID:  1,
		TxHash:   "0xhash",
		LogIndex: 2,
	})
	if rec.Code != http.StatusConflict {
		t.Fatalf("duplicate status = %d body = %s", rec.Code, rec.Body.String())
	}
	rec = perform(r, http.MethodPut, "/api/sync_event/"+itoa(event.ID), model.SyncEvent{
		ChainID:      1,
		BlockNumber:  10,
		TxHash:       "0xhash",
		LogIndex:     2,
		ContractAddr: "0xcontract",
		Removed:      model.SyncEventRemoved,
	})
	marked := decodeData[model.SyncEvent](t, rec)
	if marked.Removed != model.SyncEventRemoved || !marked.CreatedAt.Equal(event.CreatedAt) {
		t.Fatalf("event = %+v", marked)
	}

	rec = perform(r, http.MethodPost, "/api/sync_cursor", model.SyncCursor{ChainID: 1, BlockNumber: 100})
	cursor := decodeData[model.SyncCursor](t, rec)
	rec = perform(r, http.MethodPut, "/api/sync_cursor/"+itoa(cursor.ID), model.SyncCursor{ChainID: 1, BlockNumber: 250})
	moved := decodeData[model.SyncCursor](t, rec)
	if moved.BlockNumber != 250 || !moved.CreatedAt.Equal(cursor.CreatedAt) {
		t.Fatalf("cursor = %+v", moved)
	}
	rec = perform(r, http.MethodGet, "/api/sync_cursor?chainid=1", nil)
	cursors := decodeData[[]model.SyncCursor](t, rec)
	if len(cursors) != 1 || cursors[0].BlockNumber != 250 {
		t.Fatalf("cursors = %+v", cursors)
	}

	rec = perform(r, http.MethodDelete, "/api/user_info/"+itoa(user.ID), nil)
	deleted := decodeResp(t, rec)
	if deleted.Code != codeOK || string(deleted.Data) != "null" {
		t.Fatalf("delete body = %s", rec.Body.String())
	}
	rec = perform(r, http.MethodGet, "/api/user_info/"+itoa(user.ID), nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get deleted status = %d body = %s", rec.Code, rec.Body.String())
	}
	rec = perform(r, http.MethodGet, "/api/user_info/abc", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id status = %d", rec.Code)
	}

	rec = perform(r, http.MethodGet, "/health", nil)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("health = %d %s", rec.Code, rec.Body.String())
	}
}

func testRouter(t *testing.T, store *model.Store) *gin.Engine {
	t.Helper()
	return NewRouter(&svc.ServiceContext{
		Config: &config.Config{Server: config.ServerConfig{Mode: gin.TestMode}},
		Store:  store,
	})
}

func testDB(t *testing.T) *gorm.DB {
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
	return db
}

func perform(h http.Handler, method, target string, body any) *httptest.ResponseRecorder {
	var rdr io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			panic(err)
		}
		rdr = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, target, rdr)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

type apiResp struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func decodeResp(t *testing.T, rec *httptest.ResponseRecorder) apiResp {
	t.Helper()
	var body apiResp
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode %s: %v", rec.Body.String(), err)
	}
	if rec.Code >= 400 && body.Code == codeOK {
		t.Fatalf("http %d but code ok: %s", rec.Code, rec.Body.String())
	}
	return body
}

func decodeData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	body := decodeResp(t, rec)
	if rec.Code != http.StatusOK || body.Code != codeOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var data T
	if err := json.Unmarshal(body.Data, &data); err != nil {
		t.Fatalf("decode data %s: %v", rec.Body.String(), err)
	}
	return data
}

func itoa(id int64) string {
	return strconv.FormatInt(id, 10)
}
