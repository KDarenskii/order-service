package rprocessor

import (
	"net/http"
	"net/http/pprof"

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

func vGenericRegPprof(r *gin.Engine) {
	r.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	r.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	r.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	r.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	r.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	r.POST("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))

	r.GET("/debug/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
	r.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	r.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	r.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	r.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	r.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
}
