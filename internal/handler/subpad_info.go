package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createSubpad(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SubpadInfo
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	if err := store.SubpadInfo.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getSubpad(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := queryID(c)
	if !ok {
		return
	}
	row, err := store.SubpadInfo.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateSubpad(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SubpadInfo
	if !bindJSON(c, &row) {
		return
	}
	if !requireID(c, row.ID) {
		return
	}
	ctx := c.Request.Context()
	old, err := store.SubpadInfo.Get(ctx, row.ID)
	if err != nil {
		writeErr(c, err)
		return
	}
	row.CreatedAt = old.CreatedAt
	if err = store.SubpadInfo.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteSubpad(c *gin.Context) {
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
	if err := store.SubpadInfo.Delete(c.Request.Context(), body.ID); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listSubpad(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.SubpadInfoFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.SubpadInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
