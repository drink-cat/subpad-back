package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/config"
	"github.com/drink-cat/subpad-back/internal/model"
	"github.com/drink-cat/subpad-back/internal/svc"
)

func TestLoginAndDomain(t *testing.T) {
	if got, ok := brandFromHost("foods.launch.o1.local", "launch.o1.local"); !ok || got != "foods" {
		t.Fatalf("brand = %q ok = %v", got, ok)
	}
	if got, ok := brandFromHost("Foods.Launch.O1.Local:8080", "launch.o1.local"); !ok || got != "foods" {
		t.Fatalf("brand with port = %q ok = %v", got, ok)
	}
	if _, ok := brandFromHost("launch.o1.local", "launch.o1.local"); ok {
		t.Fatal("apex host should not be a brand")
	}
	if _, ok := brandFromHost("a.b.launch.o1.local", "launch.o1.local"); ok {
		t.Fatal("nested host should not be a brand")
	}

	db := testDB(t)
	store := model.NewStore(db)
	r := testRouter(t, store)
	if err := store.SubpadInfo.Create(t.Context(), &model.SubpadInfo{
		Brand:    "foods",
		NameFull: "Foods",
		SwapType: model.SwapTypeMock,
	}); err != nil {
		t.Fatal(err)
	}
	if err := store.UserInfo.Create(t.Context(), &model.UserInfo{
		Username: "alice",
		Password: "secret",
	}); err != nil {
		t.Fatal(err)
	}

	rec := perform(r, http.MethodGet, "/api/user_info/list", nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token status = %d body = %s", rec.Code, rec.Body.String())
	}

	rec = perform(r, http.MethodPost, "/api/user_info/login", map[string]any{
		"username": "alice",
		"password": "wrong",
	})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad password status = %d body = %s", rec.Code, rec.Body.String())
	}

	session := decodeData[loginData](t, perform(r, http.MethodPost, "/api/user_info/login", map[string]any{
		"username": "alice",
		"password": "secret",
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/user_info/list", nil)
	req.Host = "foods.launch.o1.local"
	req.Header.Set("Authorization", "Bearer "+session.JwtToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	users := decodeData[[]model.UserInfo](t, rec)
	if len(users) != 1 || users[0].Username != "alice" {
		t.Fatalf("users = %+v", users)
	}

	h := &Handler{sc: &svc.ServiceContext{
		Config: &config.Config{Domain: config.DomainConfig{Suffix: "launch.o1.local"}},
		Store:  store,
	}}
	probe := gin.New()
	probe.Use(h.DomainFilter())
	probe.GET("/probe", func(c *gin.Context) {
		row, ok := CurrentSubpad(c)
		if !ok {
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, gin.H{"brand": row.Brand, "id": row.ID})
	})

	req = httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Host = "foods.launch.o1.local:8080"
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"brand":"foods"`) {
		t.Fatalf("probe = %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Header().Get(headerSubpad), `"brand":"foods"`) || !strings.Contains(rec.Header().Get(headerSubpad), `"nameFull":"Foods"`) {
		t.Fatalf("subpad header = %s", rec.Header().Get(headerSubpad))
	}

	req = httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Host = "missing.launch.o1.local"
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing brand status = %d body = %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Host = "127.0.0.1:8080"
	req.Header.Set("X-Forwarded-Host", "foods.launch.o1.local")
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Header().Get(headerSubpad), `"brand":"foods"`) {
		t.Fatalf("forwarded host = %d header = %s", rec.Code, rec.Header().Get(headerSubpad))
	}

	req = httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Host = "foods.launch.o1.local"
	req.Header.Set("X-Forwarded-Host", "localhost:80")
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || rec.Header().Get(headerSubpad) != "" {
		t.Fatalf("forwarded plain host = %d header = %s", rec.Code, rec.Header().Get(headerSubpad))
	}

	req = httptest.NewRequest(http.MethodGet, "/probe", nil)
	req.Host = "localhost:8080"
	rec = httptest.NewRecorder()
	probe.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("plain host status = %d body = %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get(headerSubpad) != "" {
		t.Fatalf("plain host header = %s", rec.Header().Get(headerSubpad))
	}
}
