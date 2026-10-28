package eprocessor

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/app/entity"
	ehandler "github.com/KDarenskii/order-service/internal/app/handler/event"
	"github.com/KDarenskii/order-service/internal/app/processor"
	"github.com/KDarenskii/order-service/internal/pkg/broker"
)

type orderDeliveryCalculatedProc struct {
	h   ehandler.OrderDeliverCalculated
	bus broker.Bus[entity.EventOrderDeliveryCalculated]
}

func NewOrderDeliveryCalculatedEventsCatcher(
	h ehandler.OrderDeliverCalculated,
	bus broker.Bus[entity.EventOrderDeliveryCalculated],
) processor.Processor {
	return &orderDeliveryCalculatedProc{h, bus}
}

func (p *orderDeliveryCalculatedProc) StartAsync(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)

	if err := p.bus.Subscribe(ctx, wg, p.h.CallbackOrderDeliveryCalculated); err != nil {
		wg.Done()
		log.Fatal().Str("topic", p.bus.QueueName()).Err(err).Msg("Subscribe error")
	}

	log.Info().Str("topic", p.bus.QueueName()).Msg("Subscribed consumer")

	wg.Done()
}
