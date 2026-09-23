package mtracelog

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

const fieldTraceID = "trace_id"

type Hook struct{}

func (Hook) Run(e *zerolog.Event, _ zerolog.Level, _ string) {
	ctx := e.GetCtx()

	spanCtx := trace.SpanContextFromContext(ctx)

	if !spanCtx.IsValid() {
		return
	}

	e.Str(fieldTraceID, spanCtx.TraceID().String())
}
