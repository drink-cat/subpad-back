package handler

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/drink-cat/subpad-back/internal/model"
)

func TestAccessLogRedactsSecrets(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	db := testDB(t)
	store := model.NewStore(db)
	r := testRouter(t, store)

	rec := perform(r, http.MethodPost, "/api/user_info/create", map[string]any{
		"username": "alice",
		"password": "secret",
		"feeAddr":  "0xfee",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d body = %s", rec.Code, rec.Body.String())
	}
	log := buf.String()
	if !strings.Contains(log, "path=/api/user_info/create") || !strings.Contains(log, "alice") || !strings.Contains(log, `password\":\"***\"`) {
		t.Fatalf("log = %s", log)
	}
	if strings.Contains(log, "secret") {
		t.Fatalf("password leaked: %s", log)
	}

	buf.Reset()
	session := decodeData[loginData](t, perform(r, http.MethodPost, "/api/user_info/login", map[string]any{
		"username": "alice",
		"password": "secret",
	}))
	log = buf.String()
	if session.JwtToken == "" || strings.Contains(log, session.JwtToken) || strings.Contains(log, "secret") {
		t.Fatalf("login log leaked: %s", log)
	}
	if !strings.Contains(log, "path=/api/user_info/login") || !strings.Contains(log, `jwtToken\":\"***\"`) {
		t.Fatalf("login log = %s", log)
	}

	buf.Reset()
	req := httptest.NewRequest(http.MethodGet, "/api/user_info/get?id=1", nil)
	req.Header.Set("Authorization", "Bearer "+session.JwtToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	log = buf.String()
	if rec.Code != http.StatusOK || !strings.Contains(log, "userId=1") || !strings.Contains(log, "username=alice") || !strings.Contains(log, "id=1") {
		t.Fatalf("status = %d log = %s", rec.Code, log)
	}
	if strings.Contains(log, session.JwtToken) {
		t.Fatalf("token leaked: %s", log)
	}

	buf.Reset()
	perform(r, http.MethodGet, "/health", nil)
	if buf.Len() != 0 {
		t.Fatalf("health logged: %s", buf.String())
	}
}
