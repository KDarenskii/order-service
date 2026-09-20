package section

type Monitor struct {
	LogLevel    string `split_words:"true" default:"debug"`
	Environment string `default:"development"`
	Sentry      MonitorSentry
}

type MonitorSentry struct {
	Enabled bool   `default:"false"`
	DSN     string `default:""`
}
