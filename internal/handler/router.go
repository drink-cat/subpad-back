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

	h := &Handler{sc: sc}
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

	api := r.Group("/api")
	api.POST("/user_info/create", h.createUser)
	api.POST("/user_info/update", h.updateUser)
	api.POST("/user_info/delete", h.deleteUser)
	api.GET("/user_info/get", h.getUser)
	api.GET("/user_info/list", h.listUser)

	api.POST("/token_info/create", h.createToken)
	api.POST("/token_info/update", h.updateToken)
	api.POST("/token_info/delete", h.deleteToken)
	api.GET("/token_info/get", h.getToken)
	api.GET("/token_info/list", h.listToken)

	api.POST("/subpad_info/create", h.createSubpad)
	api.POST("/subpad_info/update", h.updateSubpad)
	api.POST("/subpad_info/delete", h.deleteSubpad)
	api.GET("/subpad_info/get", h.getSubpad)
	api.GET("/subpad_info/list", h.listSubpad)

	api.POST("/fee_info/create", h.createFee)
	api.POST("/fee_info/update", h.updateFee)
	api.POST("/fee_info/delete", h.deleteFee)
	api.GET("/fee_info/get", h.getFee)
	api.GET("/fee_info/list", h.listFee)

	api.POST("/sync_event/create", h.createSyncEvent)
	api.POST("/sync_event/update", h.updateSyncEvent)
	api.POST("/sync_event/delete", h.deleteSyncEvent)
	api.GET("/sync_event/get", h.getSyncEvent)
	api.GET("/sync_event/list", h.listSyncEvent)

	api.POST("/sync_cursor/create", h.createSyncCursor)
	api.POST("/sync_cursor/update", h.updateSyncCursor)
	api.POST("/sync_cursor/delete", h.deleteSyncCursor)
	api.GET("/sync_cursor/get", h.getSyncCursor)
	api.GET("/sync_cursor/list", h.listSyncCursor)
	return r
}
