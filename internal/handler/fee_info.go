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
	id, ok := queryID(c)
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
	var row model.FeeInfo
	if !bindJSON(c, &row) {
		return
	}
	if !requireID(c, row.ID) {
		return
	}
	ctx := c.Request.Context()
	old, err := store.FeeInfo.Get(ctx, row.ID)
	if err != nil {
		writeErr(c, err)
		return
	}
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
	var body idBody
	if !bindJSON(c, &body) {
		return
	}
	if !requireID(c, body.ID) {
		return
	}
	if err := store.FeeInfo.Delete(c.Request.Context(), body.ID); err != nil {
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
	// 只返回当前域名对应的 subpad。没有子域名前缀时是默认 pad，subpad_id 为 0。
	// fee_info 没有 pad 字段，按同一条链上的 pool_id 找到 token_info 再过滤。
	subpadID := int64(0)
	if subpad, ok := CurrentSubpad(c); ok {
		subpadID = subpad.ID
	}
	f.SubpadID = &subpadID
	rows, err := store.FeeInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
