package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/drink-cat/subpad-back/internal/svc"
)

func NewRouter(sc *svc.ServiceContext) *gin.Engine {
	mode := sc.Config.Server.Mode
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"mysql":  sc.DB != nil,
			"eth":    sc.Eth != nil,
			"cron":   sc.Scheduler != nil,
		})
	})
	return r
}
