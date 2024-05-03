package api

import (
	"net/http"

	"github.com/correctinho/correct-types-sdk-go/chain"
	"github.com/gin-gonic/gin"
)

// GetHealth - health-check
func (srv *Service) GetHealth(ctx *gin.Context) {
	chain.Response(ctx, http.StatusOK, nil)
}
