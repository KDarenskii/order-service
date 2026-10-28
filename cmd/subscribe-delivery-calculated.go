package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/KDarenskii/order-service/internal/app/builder"
)

func SubscribeDeliveryCalculated() *cli.Command {
	return &cli.Command{
		Name:            "subscribe-delivery-calculated",
		Aliases:         []string{"consume-delivery-calculated"},
		Usage:           "Запускает consumer топика order.delivery.calculated",
		Action:          cmdSubscribeDeliveryCalculated,
		HideHelpCommand: true,
	}
}

func cmdSubscribeDeliveryCalculated(cCtx *cli.Context) error {
	b := builder.NewBuilder(cCtx)
	b.BuildConfig()
	b.BuildMonitorOpenTelemetry()
	b.BuildRepoConnPostgres()
	b.BuildRepoOrder()
	b.BuildServiceOrderDelivery()
	b.BuildBrokerKafka()
	b.BuildConsumerDeliveryCalculated()
	b.BuildProcHttp()
	b.Run()

	return nil
}
