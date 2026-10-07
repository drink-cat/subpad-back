package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/model"
	"github.com/drink-cat/subpad-back/internal/svc"
)

const (
	codeOK           int64 = 0
	codeBadRequest   int64 = 400
	codeNotFound     int64 = 404
	codeConflict     int64 = 409
	codeUnauthorized int64 = 401
	codeInternal     int64 = 500
	codeUnavailable  int64 = 503
)

type BaseResp struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type Handler struct {
	sc *svc.ServiceContext
}

func okJSON(c *gin.Context, data any) {
	c.JSON(http.StatusOK, BaseResp{Code: codeOK, Message: "ok", Data: data})
}

func fail(c *gin.Context, status int, code int64, message string) {
	c.JSON(status, BaseResp{Code: code, Message: message})
}

func (h *Handler) store(c *gin.Context) (*model.Store, bool) {
	if h.sc.Store == nil {
		fail(c, http.StatusServiceUnavailable, codeUnavailable, "mysql is not configured")
		return nil, false
	}
	return h.sc.Store, true
}

type idBody struct {
	ID int64 `json:"id"`
}

func queryID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Query("id"), 10, 64)
	if err != nil || id <= 0 {
		fail(c, http.StatusBadRequest, codeBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func requireID(c *gin.Context, id int64) bool {
	if id <= 0 {
		fail(c, http.StatusBadRequest, codeBadRequest, "invalid id")
		return false
	}
	return true
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		fail(c, http.StatusBadRequest, codeBadRequest, err.Error())
		return false
	}
	return true
}

func bindQuery(c *gin.Context, dst any) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		fail(c, http.StatusBadRequest, codeBadRequest, err.Error())
		return false
	}
	return true
}

func writeErr(c *gin.Context, err error) {
	switch {
	case err == nil:
		return
	case errors.Is(err, gorm.ErrRecordNotFound):
		fail(c, http.StatusNotFound, codeNotFound, "not found")
	case isConflict(err):
		fail(c, http.StatusConflict, codeConflict, "conflict")
	case strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "nil row"):
		fail(c, http.StatusBadRequest, codeBadRequest, err.Error())
	default:
		slog.Error("api", "method", c.Request.Method, "path", c.Request.URL.Path, "err", err)
		fail(c, http.StatusInternalServerError, codeInternal, "internal error")
	}
}

func isConflict(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unique") || strings.Contains(msg, "duplicate")
}
