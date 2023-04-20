package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitializePingRoutes(webserver *gin.Engine) {
	ping := webserver.Group("/ping")
	ping.GET(
		"",
		func(context *gin.Context) {
			context.JSON(http.StatusOK, gin.H{"status": "healthy"})
		},
	)
}
