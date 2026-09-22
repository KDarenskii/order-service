package pmonitor

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/KDarenskii/order-service/internal/app/config/section"
	"github.com/KDarenskii/order-service/internal/app/constant"
	"github.com/KDarenskii/order-service/internal/app/processor"
)

const (
	initTimeout     = 5 * time.Second
	shutdownTimeout = 5 * time.Second
)

type (
	openTelemetryProc struct {
		traceProvider *sdktrace.TracerProvider
		conn          *grpc.ClientConn
	}

	openTelemetryErrorHandler struct{}
)

func NewOpenTelemetryController(
	ctx context.Context, env string,
	cfg section.MonitorOpenTelemetry,
) (processor.Processor, error) {
	ctx, cancel := context.WithTimeout(ctx, initTimeout)
	defer cancel()

	var p openTelemetryProc

	attributes := []attribute.KeyValue{semconv.ServiceName(constant.AppName)}

	if env != "" {
		attributes = append(attributes, semconv.DeploymentEnvironment(strings.ToLower(env)))
	}

	res, err := resource.New(ctx, resource.WithAttributes(attributes...))
	if err != nil {
		return nil, fmt.Errorf("create resource: %w", err)
	}

	conn, err := grpc.NewClient(
		cfg.Address,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)
	if err != nil {
		closeJaegerConn(conn)
		return nil, fmt.Errorf("create gRPC client for Jaeger: %w", err)
	}

	err = waitForReady(ctx, conn)
	if err != nil {
		closeJaegerConn(conn)
		return nil, fmt.Errorf("connect to Jaeger at %s: %w", cfg.Address, err)
	}

	p.conn = conn

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		closeJaegerConn(conn)
		return nil, fmt.Errorf("create OTLP exporter: %w", err)
	}

	cfg.SampleRatio = min(1, max(0, cfg.SampleRatio))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))),
		sdktrace.WithBatcher(
			//nolint:gosec // G115
			exporter, sdktrace.WithMaxQueueSize(int(cfg.MaxQueueSize)),
			//nolint:gosec // G115
			sdktrace.WithMaxExportBatchSize(int(cfg.MaxBatchSize)),
			sdktrace.WithBatchTimeout(cfg.SendBatchTimeout),
			sdktrace.WithExportTimeout(cfg.ExportTimeout),
		),
	)

	p.traceProvider = tp

	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}),
	)

	otel.SetErrorHandler(openTelemetryErrorHandler{})

	log.Info().Str("service", constant.AppName).Str(
		"environment", strings.ToLower(env),
	).Msg("OpenTelemetry has been initialized")

	return &p, nil
}

func waitForReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()

	for {
		state := conn.GetState()

		if state == connectivity.Ready {
			return nil
		}

		if !conn.WaitForStateChange(ctx, state) {
			return fmt.Errorf("wait for Jaeger ready: %w", ctx.Err())
		}

	}
}

func (p *openTelemetryProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	processor.WatchForShutdown(ctx, wg, processor.CloserFunc(p.shutdown))
}

func (p *openTelemetryProc) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := p.traceProvider.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to shutdown trace provider")
	}
	if err := p.conn.Close(); err != nil {
		log.Error().Err(err).Msg("Failed to close Jaeger gRPC connection")
	}
	return nil
}

func closeJaegerConn(conn *grpc.ClientConn) {
	if conn == nil {
		return
	}

	err := conn.Close()
	if err != nil {
		log.Error().Err(err).Msg("Failed to close Jaeger gRPC connection")
	}
}

func (openTelemetryErrorHandler) Handle(err error) {
	log.Error().Err(err).Msg("OpenTelemetry error")
}
