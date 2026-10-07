package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createFee(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.FeeInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	if err := store.FeeInfo.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getFee(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	row, err := store.FeeInfo.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateFee(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	old, err := store.FeeInfo.Get(ctx, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	var row model.FeeInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = id
	row.CreatedAt = old.CreatedAt
	if err = store.FeeInfo.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteFee(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := store.FeeInfo.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listFee(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.FeeInfoFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.FeeInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
