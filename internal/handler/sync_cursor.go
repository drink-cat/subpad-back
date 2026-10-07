package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createSyncCursor(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SyncCursor
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	if err := store.SyncCursor.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getSyncCursor(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	row, err := store.SyncCursor.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateSyncCursor(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	old, err := store.SyncCursor.Get(ctx, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	var row model.SyncCursor
	if !bindJSON(c, &row) {
		return
	}
	row.ID = id
	row.CreatedAt = old.CreatedAt
	if err = store.SyncCursor.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteSyncCursor(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := store.SyncCursor.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listSyncCursor(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.SyncCursorFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.SyncCursor.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
