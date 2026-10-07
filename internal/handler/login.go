package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/auth"
)

type loginBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginData struct {
	JwtToken string `json:"jwtToken"`
}

func (h *Handler) login(c *gin.Context) {
	store, ok := h.store(c)
	if !ok {
		return
	}
	var body loginBody
	if !bindJSON(c, &body) {
		return
	}
	if body.Username == "" || body.Password == "" {
		fail(c, http.StatusBadRequest, codeBadRequest, "username and password are required")
		return
	}
	user, err := store.UserInfo.GetByUsername(c.Request.Context(), body.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		writeErr(c, err)
		return
	}
	if err != nil || user.CheckPassword(body.Password) != nil {
		fail(c, http.StatusUnauthorized, codeUnauthorized, "invalid username or password")
		return
	}
	token, err := auth.Sign(h.sc.Config.JWT.Secret, user.ID, user.Username, h.jwtTTL())
	if err != nil {
		writeErr(c, err)
		return
	}
	okJSON(c, loginData{JwtToken: token})
}
