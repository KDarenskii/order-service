package section

import "time"

type (
	Processor struct {
		WebServer ProcessorWebServer `split_words:"true"`
	}

	ProcessorWebServer struct {
		ListenPort        uint32        `split_words:"true" default:"8080"`
		ReadHeaderTimeout time.Duration `split_words:"true" default:"5s"`
	}
)
