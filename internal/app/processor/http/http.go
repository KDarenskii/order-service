package rprocessor

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KDarenskii/order-service/internal/app/config/section"
	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
)

type httpProc struct {
	server *http.Server
	addr   string
}

func NewHTTP(hHealth rhandler.Health, cfg section.ProcessorWebServer) *httpProc {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(gin.Recovery())

	router.NoRoute(handleNotFound)

	vGenericRegHealthCheck(router, hHealth)

	for _, routeInfo := range router.Routes() {
		if routeInfo.Path == "" || routeInfo.Method == "" {
			continue
		}

		log.Printf("Registered http route: %s %s", routeInfo.Method, routeInfo.Path)
	}

	addr := fmt.Sprintf(":%d", cfg.ListenPort)

	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}

	return &httpProc{server: server, addr: addr}
}

func (p *httpProc) Serve() error {
	log.Printf("Starting HTTP server on %s", p.addr)
	return p.server.ListenAndServe()
}
