package rprocessor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/app/config/section"
	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
	"github.com/KDarenskii/order-service/internal/app/processor"
	"github.com/KDarenskii/order-service/internal/pkg/http/httph"
	"github.com/KDarenskii/order-service/internal/pkg/http/mzerolog"
	"github.com/KDarenskii/order-service/internal/util"
)

type httpProc struct {
	server http.Server
	addr   string
}

func NewHTTP(hHealth rhandler.Health,
	hOrder rhandler.Order,
	cfg section.ProcessorWebServer,
) processor.Processor {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(
		adaptRequestMiddleware(httph.NewErrorMiddleware()),
		mzerolog.NewMiddleware(mzerolog.WithSkipper(util.IsFilteredHttpRoute)),
		gin.Recovery(),
	)

	router.NoRoute(handleNotFound)

	vGenericRegHealthCheck(router, hHealth)

	v1 := router.Group("/v1")
	v1RegOrderHandler(v1, hOrder)

	for _, routeInfo := range router.Routes() {
		if routeInfo.Path == "" || routeInfo.Method == "" {
			continue
		}

		log.Info().Str("method", routeInfo.Method).Str("path", routeInfo.Path).Msg("Registered http route")
	}

	p := httpProc{addr: fmt.Sprintf(":%d", cfg.ListenPort)}
	p.server.Handler = router
	p.server.ReadHeaderTimeout = cfg.ReadHeaderTimeout

	return &p
}

func (p *httpProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	var lc net.ListenConfig

	l, err := lc.Listen(ctx, "tcp", p.addr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start http server")
	}

	log.Info().Str("address", p.addr).Str("network", "TCP").Msg("Started http server")

	go p.serve(l)

	processor.WatchForShutdown(ctx, wg, processor.NewCloserContextFunc(p.server.Shutdown, context.Background(), 5*time.Second))
}

func (p *httpProc) serve(l net.Listener) {
	_ = p.server.Serve(l)
}

func adaptRequestMiddleware(m httph.Middleware) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var called bool

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx.Request = r
			ctx.Next()
		})

		m(next).ServeHTTP(ctx.Writer, ctx.Request)

		if !called {
			ctx.Abort()
		}
	}
}
