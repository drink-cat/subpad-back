package handler

import (
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
	id, ok := pathID(c)
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
	id, ok := pathID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	if _, err := store.TokenInfo.Get(ctx, id); err != nil {
		writeErr(c, err)
		return
	}
	var row model.TokenInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = id
	if err := store.TokenInfo.Update(ctx, &row); err != nil {
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
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := store.TokenInfo.Delete(c.Request.Context(), id); err != nil {
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
	rows, err := store.TokenInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
