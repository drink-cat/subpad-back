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
	mountCRUD(api, "/user_info", h.createUser, h.listUser, h.getUser, h.updateUser, h.deleteUser)
	mountCRUD(api, "/token_info", h.createToken, h.listToken, h.getToken, h.updateToken, h.deleteToken)
	mountCRUD(api, "/subpad_info", h.createSubpad, h.listSubpad, h.getSubpad, h.updateSubpad, h.deleteSubpad)
	mountCRUD(api, "/fee_info", h.createFee, h.listFee, h.getFee, h.updateFee, h.deleteFee)
	mountCRUD(api, "/sync_event", h.createSyncEvent, h.listSyncEvent, h.getSyncEvent, h.updateSyncEvent, h.deleteSyncEvent)
	mountCRUD(api, "/sync_cursor", h.createSyncCursor, h.listSyncCursor, h.getSyncCursor, h.updateSyncCursor, h.deleteSyncCursor)
	return r
}

func mountCRUD(api *gin.RouterGroup, path string, create, list, get, update, del gin.HandlerFunc) {
	api.POST(path, create)
	api.GET(path, list)
	api.GET(path+"/:id", get)
	api.PUT(path+"/:id", update)
	api.DELETE(path+"/:id", del)
}
