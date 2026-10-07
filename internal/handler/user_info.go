package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/model"
)

type userBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
	FeeAddr  string `json:"fee_addr"`
}

func (h *Handler) createUser(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var body userBody
	if !bindJSON(c, &body) {
		return
	}
	row := &model.UserInfo{
		Username: body.Username,
		Password: body.Password,
		FeeAddr:  body.FeeAddr,
	}
	if err := store.UserInfo.Create(c.Request.Context(), row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) getUser(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	row, err := store.UserInfo.Get(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) updateUser(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	ctx := c.Request.Context()
	row, err := store.UserInfo.Get(ctx, id)
	if err != nil {
		writeErr(c, err)
		return
	}
	var body userBody
	if !bindJSON(c, &body) {
		return
	}
	row.Username = body.Username
	row.FeeAddr = body.FeeAddr
	if body.Password != "" {
		row.Password = body.Password
	}
	if err = store.UserInfo.Update(ctx, row); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, row)
}

func (h *Handler) deleteUser(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := store.UserInfo.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, nil)
}

func (h *Handler) listUser(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var f model.UserInfoFilter
	if !bindQuery(c, &f) {
		return
	}
	rows, err := store.UserInfo.List(c.Request.Context(), f)
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, rows)
}
