package rprocessor

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
)

func vGenericRegHealthCheck(r *gin.Engine, h rhandler.Health) {
	r.GET("/health", h.LastCheck)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func handleNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"error": "route not found"})
}
