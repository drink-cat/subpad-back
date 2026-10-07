package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const apiLogBodyLimit = 8192

// AccessLog 记录每次 /api 请求的路径、状态、请求体和响应体。密码和 jwtToken 不写入日志。
func (h *Handler) AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Next()
			return
		}
		start := time.Now()
		reqBody := readRequestBody(c)
		capture := &bodyLogWriter{ResponseWriter: c.Writer}
		c.Writer = capture
		c.Next()

		status := c.Writer.Status()
		args := []any{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"latencyMs", time.Since(start).Milliseconds(),
		}
		if q := c.Request.URL.RawQuery; q != "" {
			args = append(args, "query", q)
		}
		if code, message, ok := respMeta(capture.body.Bytes()); ok {
			args = append(args, "code", code, "message", message)
		}
		if user, ok := CurrentUser(c); ok {
			args = append(args, "userId", user.ID, "username", user.Username)
		}
		if subpad, ok := CurrentSubpad(c); ok {
			args = append(args, "brand", subpad.Brand, "subpadId", subpad.ID)
		}
		if req := redactBody(reqBody); req != "" {
			args = append(args, "req", req)
		}
		if resp := redactBody(capture.body.Bytes()); resp != "" {
			args = append(args, "resp", resp)
		}
		if last := c.Errors.Last(); last != nil {
			args = append(args, "err", last.Err)
		}
		logAPI(status, args...)
	}
}

func logAPI(status int, args ...any) {
	switch {
	case status >= 500:
		slog.Error("api", args...)
	case status >= 400:
		slog.Warn("api", args...)
	default:
		slog.Info("api", args...)
	}
}

func readRequestBody(c *gin.Context) []byte {
	if c.Request.Body == nil {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyLogWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func respMeta(body []byte) (int64, string, bool) {
	var meta struct {
		Code    int64  `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(body, &meta); err != nil {
		return 0, "", false
	}
	return meta.Code, meta.Message, true
}

func redactBody(body []byte) string {
	body = bytes.TrimSpace(body)
	if len(body) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return clip(string(body))
	}
	redactValue(v)
	out, err := json.Marshal(v)
	if err != nil {
		return clip(string(body))
	}
	return clip(string(out))
}

func redactValue(v any) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if secretKey(k) {
				t[k] = "***"
				continue
			}
			redactValue(val)
		}
	case []any:
		for _, item := range t {
			redactValue(item)
		}
	}
}

func secretKey(k string) bool {
	switch strings.ToLower(k) {
	case "password", "jwttoken", "authorization":
		return true
	default:
		return false
	}
}

func clip(s string) string {
	if len(s) <= apiLogBodyLimit {
		return s
	}
	return s[:apiLogBodyLimit] + "...(truncated)"
}
