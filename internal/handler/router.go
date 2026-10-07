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
	r.Use(gin.Recovery(), h.DomainFilter())
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"mysql":  sc.DB != nil,
			"eth":    sc.Eth != nil,
			"cron":   sc.Scheduler != nil,
		})
	})

	api := r.Group("/api")
	api.GET("/config", h.getConfig)
	api.POST("/user_info/login", h.login)
	api.POST("/user_info/create", h.createUser)

	authn := api.Group("")
	authn.Use(h.JwtFilter())
	authn.POST("/user_info/update", h.updateUser)
	authn.POST("/user_info/delete", h.deleteUser)
	authn.GET("/user_info/get", h.getUser)
	authn.GET("/user_info/list", h.listUser)

	authn.POST("/token_info/create", h.createToken)
	authn.POST("/token_info/update", h.updateToken)
	authn.POST("/token_info/delete", h.deleteToken)
	authn.GET("/token_info/get", h.getToken)
	authn.GET("/token_info/list", h.listToken)

	authn.POST("/subpad_info/create", h.createSubpad)
	authn.POST("/subpad_info/update", h.updateSubpad)
	authn.POST("/subpad_info/delete", h.deleteSubpad)
	authn.GET("/subpad_info/get", h.getSubpad)
	authn.GET("/subpad_info/list", h.listSubpad)

	authn.POST("/fee_info/create", h.createFee)
	authn.POST("/fee_info/update", h.updateFee)
	authn.POST("/fee_info/delete", h.deleteFee)
	authn.GET("/fee_info/get", h.getFee)
	authn.GET("/fee_info/list", h.listFee)

	authn.POST("/swap_info/create", h.createSwap)
	authn.POST("/swap_info/update", h.updateSwap)
	authn.POST("/swap_info/delete", h.deleteSwap)
	authn.GET("/swap_info/get", h.getSwap)
	authn.GET("/swap_info/list", h.listSwap)

	authn.POST("/sync_event/create", h.createSyncEvent)
	authn.POST("/sync_event/update", h.updateSyncEvent)
	authn.POST("/sync_event/delete", h.deleteSyncEvent)
	authn.GET("/sync_event/get", h.getSyncEvent)
	authn.GET("/sync_event/list", h.listSyncEvent)

	authn.POST("/sync_cursor/create", h.createSyncCursor)
	authn.POST("/sync_cursor/update", h.updateSyncCursor)
	authn.POST("/sync_cursor/delete", h.deleteSyncCursor)
	authn.GET("/sync_cursor/get", h.getSyncCursor)
	authn.GET("/sync_cursor/list", h.listSyncCursor)
	return r
}
