package app

import (
	"github.com/correctinho/correct-jwt-service-go/api"
	"github.com/correctinho/correct-types-sdk-go/chain"
	"github.com/gin-gonic/gin"
)

// Router - rotas do serviço
func Router(g *gin.Engine) {

	srv := api.NewService()

	v1 := g.Group("/api").Group("/v1")

	v1.GET("/jwt/health", srv.GetHealth)

	v1.POST("/jwt/encode", srv.PostEncode, func(ctx *gin.Context) { chain.WriteResponse(ctx) })

	v1.POST("/jwt/decode", srv.PostDecode, func(ctx *gin.Context) { chain.WriteResponse(ctx) })
}
