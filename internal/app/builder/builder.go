package builder

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"

	ccatalog "github.com/KDarenskii/order-service/internal/app/client/catalog"
	cgrpc "github.com/KDarenskii/order-service/internal/app/client/catalog/grpc"
	"github.com/KDarenskii/order-service/internal/app/config"
	"github.com/KDarenskii/order-service/internal/app/constant"
	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
	rhealth "github.com/KDarenskii/order-service/internal/app/handler/http/health"
	horder "github.com/KDarenskii/order-service/internal/app/handler/http/order"
	"github.com/KDarenskii/order-service/internal/app/processor"
	rprocessor "github.com/KDarenskii/order-service/internal/app/processor/http"
	pmonitor "github.com/KDarenskii/order-service/internal/app/processor/monitor"
	"github.com/KDarenskii/order-service/internal/app/repository"
	rcpostgres "github.com/KDarenskii/order-service/internal/app/repository/conn/postgres"
	porder "github.com/KDarenskii/order-service/internal/app/repository/order"
	"github.com/KDarenskii/order-service/internal/app/service"
	sorder "github.com/KDarenskii/order-service/internal/app/service/order"
)

type Builder struct {
	cCtx *cli.Context
	ctx  context.Context
	wg   sync.WaitGroup
	err  error
	cfg  config.Config

	chErrors chan error

	connPostgres    *rcpostgres.Client
	catalogGrpcConn *grpc.ClientConn

	catalogV1Client ccatalog.Client

	orderRepo repository.Order

	orderService service.Order

	healthHandler rhandler.Health
	orderHandler  rhandler.Order

	processors []processor.Processor

	otelServiceName string
}

func NewBuilder(cCtx *cli.Context) *Builder {
	b := Builder{
		cCtx:     cCtx,
		chErrors: make(chan error, 4096),
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	b.ctx = ctx

	signalChan := make(chan os.Signal, 1)

	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	go b.waitForSignal(signalChan, cancelFunc)

	go b.printErrors()

	b.healthHandler = rhealth.NewHandler()

	return &b
}

func (b *Builder) BuildConfig() {
	b.exec(
		func(b *Builder) {
			b.buildConfig()
		},
	)
}

func (b *Builder) Run() {
	if b.ctx.Err() != nil {
		log.Info().Msg("Shutdown during initialization")
		return
	}

	if b.err != nil {
		log.Fatal().Err(b.err).Msg("Failed to initialize application")
	}

	log.Info().Msg("Application initialized")

	defer log.Info().Msg("Application completed")

	for _, proc := range b.processors {
		proc.StartAsync(b.ctx, &b.wg)
	}

	b.wg.Wait()
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORY CONNECTIONS ///////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoConnPostgres() {
	b.exec(
		func(b *Builder) {
			pgClient, err := rcpostgres.NewClient(b.ctx, b.cfg.Repository.Postgres)
			if err != nil {
				b.err = err
				return
			}

			b.connPostgres = pgClient
		},
	)
}

////////////////////////////////////////////////////////////////////////////////
///// REPOSITORIES /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildRepoOrder() {
	b.exec(
		func(b *Builder) {
			b.orderRepo = porder.NewRepo(b.connPostgres)
		}, b.connPostgres,
	)
}

////////////////////////////////////////////////////////////////////////////////
///// CLIENTS /////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildClientGrpcCatalogV1() {
	b.exec(
		func(b *Builder) {
			client, conn, err := cgrpc.NewClient(b.cfg.Client.Catalog.GrpcAddress)
			if err != nil {
				b.err = err
				return
			}

			if err := client.Ping(b.ctx); err != nil {
				_ = conn.Close()
				b.err = err
				return
			}

			processor.WatchForShutdown(b.ctx, &b.wg, processor.CloserFunc(conn.Close))

			b.catalogGrpcConn = conn
			b.catalogV1Client = client
		},
	)
}

////////////////////////////////////////////////////////////////////////////////
///// SERVICES /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildServiceOrder() {
	b.exec(
		func(b *Builder) {
			b.orderService = sorder.NewService(b.orderRepo, b.catalogV1Client)
		}, b.orderRepo, b.catalogV1Client,
	)
}

////////////////////////////////////////////////////////////////////////////////
///// HANDLERS /////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildHandlerOrder() {
	b.exec(
		func(b *Builder) {
			b.orderHandler = horder.NewHandler(b.orderService)
		}, b.orderService,
	)
}

////////////////////////////////////////////////////////////////////////////////
///// PROCESSORS ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildProcHttp() {
	b.exec(
		func(b *Builder) {
			proc := rprocessor.NewHTTP(b.otelServiceName, b.healthHandler, b.orderHandler, b.cfg.Processor.WebServer)
			b.processors = append(b.processors, proc)
		}, b.healthHandler, b.orderHandler,
	)
}

////////////////////////////////////////////////////////////////////////////////
///// MONITOR ///////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) BuildMonitorOpenTelemetry() {
	cfg := b.cfg.Monitor.OpenTelemetry

	if !cfg.Enabled {
		log.Warn().Msg("OpenTelemetry is disabled by config")
		return
	}

	b.exec(
		func(b *Builder) {
			proc, err := pmonitor.NewOpenTelemetryController(b.ctx, b.cfg.Monitor.Environment, cfg)
			if err != nil {
				b.err = fmt.Errorf("init OpenTelemetry: %w", err)
				return
			}

			b.processors = append(b.processors, proc)

			b.otelServiceName = constant.AppName
		},
	)
}

////////////////////////////////////////////////////////////////////////////////
///// PRIVATE //////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (b *Builder) buildConfig() {
	config.Load(
		config.LoadArgs{
			Output:          b.cCtx.App.Writer,
			EnableSimpleLog: b.cCtx.Bool("no-json"),
		},
	)

	b.cfg = config.Root
}

func (b *Builder) exec(cb func(b *Builder), requiredArgs ...any) {
	if b.err != nil || b.ctx.Err() != nil {
		return
	}

	for i, requiredArg := range requiredArgs {
		rv := reflect.ValueOf(requiredArg)
		if !rv.IsValid() {
			b.err = fmt.Errorf("BUG: required argument #%d is nil (check dependencies)", i)
			return
		}
		if rv.Type().Kind() == reflect.Struct || !rv.IsZero() {
			continue
		}
		b.err = fmt.Errorf("BUG: required %s, but empty", rv.Type().String())
		return
	}

	cb(b)
}

func (b *Builder) waitForSignal(sig chan os.Signal, cancelFunc func()) {
	s := <-sig
	log.Info().Str("signal", s.String()).Msg("Shutdown is requested")
	cancelFunc()
}

func (b *Builder) printErrors() {
	for chErr := range b.chErrors {
		log.Error().Err(chErr).Msg("Got new error")
	}
}
