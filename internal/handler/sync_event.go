package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

func (h *Handler) createSyncEvent(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SyncEvent
	if !bindJSON(c, &row) {
		return
	}
	row.ID = 0
	row.CreatedAt = time.Time{}
	row.UpdatedAt = time.Time{}
	if err := store.SyncEvent.Create(c.Request.Context(), &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getSyncEvent(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := queryID(c)
	if !ok {
		return
	}
	row, err := store.SyncEvent.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateSyncEvent(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var row model.SyncEvent
	if !bindJSON(c, &row) {
		return
	}
	if !requireID(c, row.ID) {
		return
	}
	ctx := c.Request.Context()
	old, err := store.SyncEvent.Get(ctx, row.ID)
	if err != nil {
		writeErr(c, err)
		return
	}
	row.CreatedAt = old.CreatedAt
	if err = store.SyncEvent.Update(ctx, &row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteSyncEvent(c *gin.Context) {
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
	if err := store.SyncEvent.Delete(c.Request.Context(), body.ID); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listSyncEvent(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.SyncEventFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.SyncEvent.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
