package msentry

import (
	"io"
	"time"

	sentryGo "github.com/getsentry/sentry-go"
	sentryzerolog "github.com/getsentry/sentry-go/zerolog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/app/config/section"
)

const flushTimeout = 5 * time.Second

var enabled bool

type Options struct {
	ServiceName string
	Environment string
	Release     string
}

func Init(cfg section.MonitorSentry, opts Options) (io.Writer, bool) {
	if !cfg.Enabled {
		log.Warn().Msg("Sentry is disabled by config")
		return nil, false
	}

	if cfg.DSN == "" {
		log.Warn().Msg("Sentry DSN is empty, integration is skipped")
		return nil, false
	}

	if err := sentryGo.Init(
		sentryGo.ClientOptions{
			Dsn:              cfg.DSN,
			Release:          opts.Release,
			Environment:      opts.Environment,
			AttachStacktrace: true,
		},
	); err != nil {
		log.Error().Err(err).Msg("Failed to initialize Sentry; running without it")
		return nil, false
	}

	sentryGo.ConfigureScope(
		func(scope *sentryGo.Scope) {
			scope.SetTag("service", opts.ServiceName)
		},
	)

	writer, err := sentryzerolog.NewWithHub(
		sentryGo.CurrentHub(), sentryzerolog.Options{
			Levels: []zerolog.Level{
				zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
			},
			WithBreadcrumbs: false,
			FlushTimeout:    flushTimeout,
		},
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Sentry writer; running without it")
		return nil, false
	}

	enabled = true

	log.Info().Msg("Sentry has been initialized")

	return writer, true
}

func Flush() {
	if !enabled {
		return
	}

	success := sentryGo.Flush(flushTimeout)

	if !success {
		log.Warn().Dur("timeout", flushTimeout).Msg("Sentry flush timeout exceeded")
		return
	}

	log.Info().Msg("Sentry events have been flushed")
}
