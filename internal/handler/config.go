package handler

import (
	"github.com/gin-gonic/gin"
)

type publicConfig struct {
	QuoteToken publicQuoteToken `json:"quoteToken"`
	SyncLog    []publicSyncLog  `json:"syncLog"`
}

type publicQuoteToken struct {
	LocalUsdc   string `json:"localUsdc"`
	SepoliaUsdc string `json:"sepoliaUsdc"`
}

type publicSyncLog struct {
	Name           string `json:"name"`
	ChainID        int64  `json:"chainId"`
	RPCURL         string `json:"rpcUrl"`
	LaunchContract string `json:"launchContract"`
}

func (h *Handler) getConfig(c *gin.Context) {
	cfg := h.sc.Config
	chains := make([]publicSyncLog, 0, len(cfg.SyncLog))
	for _, item := range cfg.SyncLog {
		chains = append(chains, publicSyncLog{
			Name:           item.Name,
			ChainID:        item.ChainID,
			RPCURL:         item.RPCURL,
			LaunchContract: item.LaunchContract,
		})
	}
	okJSON(c, publicConfig{
		QuoteToken: publicQuoteToken{
			LocalUsdc:   cfg.QuoteToken.LocalUsdc,
			SepoliaUsdc: cfg.QuoteToken.SepoliaUsdc,
		},
		SyncLog: chains,
	})
}
