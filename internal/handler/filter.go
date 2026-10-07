package handler

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/drink-cat/subpad-back/internal/auth"
	"github.com/drink-cat/subpad-back/internal/model"
)

const (
	ctxUser      = "currentUser"
	ctxSubpad    = "currentSubpad"
	headerSubpad = "X-Subpad-Info"
)

func CurrentUser(c *gin.Context) (*model.UserInfo, bool) {
	v, ok := c.Get(ctxUser)
	if !ok {
		return nil, false
	}
	user, ok := v.(*model.UserInfo)
	return user, ok && user != nil
}

func CurrentSubpad(c *gin.Context) (*model.SubpadInfo, bool) {
	v, ok := c.Get(ctxSubpad)
	if !ok {
		return nil, false
	}
	row, ok := v.(*model.SubpadInfo)
	return row, ok && row != nil
}

func (h *Handler) JwtFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.sc.Store == nil {
			fail(c, http.StatusServiceUnavailable, codeUnavailable, "mysql is not configured")
			c.Abort()
			return
		}
		raw, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			fail(c, http.StatusUnauthorized, codeUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		userID, _, err := auth.Parse(h.sc.Config.JWT.Secret, raw)
		if err != nil {
			fail(c, http.StatusUnauthorized, codeUnauthorized, "unauthorized")
			c.Abort()
			return
		}
		user, err := h.sc.Store.UserInfo.Get(c.Request.Context(), userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fail(c, http.StatusUnauthorized, codeUnauthorized, "unauthorized")
			} else {
				writeErr(c, err)
			}
			c.Abort()
			return
		}
		user.Password = ""
		c.Set(ctxUser, user)
		c.Next()
	}
}

func (h *Handler) DomainFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		brand, ok := brandFromHost(c.Request.Host, h.sc.Config.Domain.Suffix)
		if !ok {
			c.Next()
			return
		}
		if h.sc.Store == nil {
			fail(c, http.StatusServiceUnavailable, codeUnavailable, "mysql is not configured")
			c.Abort()
			return
		}
		row, err := h.sc.Store.SubpadInfo.GetByBrand(c.Request.Context(), brand)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				fail(c, http.StatusNotFound, codeNotFound, "subpad not found")
			} else {
				writeErr(c, err)
			}
			c.Abort()
			return
		}
		raw, err := json.Marshal(row)
		if err != nil {
			writeErr(c, err)
			c.Abort()
			return
		}
		c.Header(headerSubpad, string(raw))
		c.Set(ctxSubpad, row)
		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

func brandFromHost(host, suffix string) (string, bool) {
	host = strings.ToLower(strings.TrimSpace(host))
	suffix = strings.ToLower(strings.Trim(strings.TrimSpace(suffix), "."))
	if host == "" || suffix == "" {
		return "", false
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil {
		host = parsed
	}
	host = strings.TrimSuffix(host, ".")
	prefix, ok := strings.CutSuffix(host, "."+suffix)
	if !ok || prefix == "" || strings.Contains(prefix, ".") {
		return "", false
	}
	return prefix, true
}

func (h *Handler) jwtTTL() time.Duration {
	hours := h.sc.Config.JWT.ExpireHours
	if hours <= 0 {
		hours = 72
	}
	return time.Duration(hours) * time.Hour
}
