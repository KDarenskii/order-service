package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/KDarenskii/order-service/cmd"
	"github.com/KDarenskii/order-service/internal/app/constant"
	msentry "github.com/KDarenskii/order-service/internal/app/monitor/sentry"
)

func main() {
	app := &cli.App{
		Name:    constant.AppName,
		Version: constant.Version,
		Usage:   "Order management service",
		Commands: []*cli.Command{
			cmd.WebServer(),
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "no-json",
				Usage: "Human-readable log format",
			},
		},
	}

	defer msentry.Flush()

	if err := app.Run(os.Args); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
