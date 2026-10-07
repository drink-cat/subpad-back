package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createSwap(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SwapInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	if err := store.SwapInfo.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getSwap(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := queryID(c)
	if !ok {
		return
	}
	row, err := store.SwapInfo.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateSwap(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SwapInfo
	if !bindJSON(c, &row) {
		return
	}
	if !requireID(c, row.ID) {
		return
	}
	ctx := c.Request.Context()
	old, err := store.SwapInfo.Get(ctx, row.ID)
	if err != nil {
		writeErr(c, err)
		return
	}
	row.CreatedAt = old.CreatedAt
	if err = store.SwapInfo.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteSwap(c *gin.Context) {
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
	if err := store.SwapInfo.Delete(c.Request.Context(), body.ID); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listSwap(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.SwapInfoFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.SwapInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
