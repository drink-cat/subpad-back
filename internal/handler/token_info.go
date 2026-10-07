package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createToken(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.TokenInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	user, ok := CurrentUser(c)
	if !ok {
		fail(c, http.StatusUnauthorized, codeUnauthorized, "unauthorized")
		return
	}
	row.UserID = user.ID
	if err := store.TokenInfo.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getToken(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := queryID(c)
	if !ok {
		return
	}
	row, err := store.TokenInfo.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateToken(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.TokenInfo
	if !bindJSON(c, &row) {
		return
	}
	if !requireID(c, row.ID) {
		return
	}
	ctx := c.Request.Context()
	old, err := store.TokenInfo.Get(ctx, row.ID)
	if err != nil {
		writeErr(c, err)
		return
	}
	row.UserID = old.UserID
	if err = store.TokenInfo.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteToken(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var body idBody
	if !bindJSON(c, &body) {
		return
	}
	if !requireID(c, body.ID) {
		return
	}
	if err := store.TokenInfo.Delete(c.Request.Context(), body.ID); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listToken(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.TokenInfoFilter
	if !bindQuery(c, &f) {
		return
	}
	// 只返回当前域名对应的 subpad。没有子域名前缀时是默认 pad，subpad_id 为 0。
	subpadID := int64(0)
	if subpad, ok := CurrentSubpad(c); ok {
		subpadID = subpad.ID
	}
	f.SubpadID = &subpadID
	rows, err := store.TokenInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
