package eorder

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/app/entity"
	ehandler "github.com/KDarenskii/order-service/internal/app/handler/event"
	"github.com/KDarenskii/order-service/internal/app/service"
	"github.com/KDarenskii/order-service/internal/pkg/broker"
)

type handler struct {
	srv service.OrderDelivery
}

func NewHandler(srv service.OrderDelivery) ehandler.OrderDeliverCalculated {
	return &handler{srv: srv}
}

func (h *handler) CallbackOrderDeliveryCalculated(
	ctx context.Context, ev *entity.EventOrderDeliveryCalculated, _ []broker.Header,
) error {
	log.Info().Ctx(ctx).EmbedObject(ev).Msg("Получено сообщение")

	orderGUID, err := uuid.FromString(ev.OrderGUID)
	if err != nil {
		return broker.NotCriticalError(fmt.Errorf("broker: invalid order_guid: %s", ev.OrderGUID))
	}

	if err := h.srv.UpdateDelivery(ctx, orderGUID, ev.DeliveryPrice); err != nil {
		return fmt.Errorf("update delivery: %w", err)
	}

	log.Info().EmbedObject(ev).Msg("Сообщение обработано")

	return nil
}
