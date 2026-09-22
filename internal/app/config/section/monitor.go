package section

import "time"

type Monitor struct {
	LogLevel      string `split_words:"true" default:"debug"`
	Environment   string `default:"development"`
	Sentry        MonitorSentry
	OpenTelemetry MonitorOpenTelemetry `split_words:"true"`
}

type MonitorSentry struct {
	Enabled bool   `default:"false"`
	DSN     string `default:""`
}

type MonitorOpenTelemetry struct {
	Enabled          bool          `default:"false"`
	Address          string        `default:""`
	MaxQueueSize     uint64        `default:"2048" split_words:"true"`
	MaxBatchSize     uint64        `default:"512" split_words:"true"`
	SendBatchTimeout time.Duration `default:"5s" split_words:"true"`
	ExportTimeout    time.Duration `default:"30s" split_words:"true"`
	SampleRatio      float64       `default:"1" split_words:"true"`
}
