package ehandler

import (
	"context"

	"github.com/KDarenskii/order-service/internal/app/entity"
	"github.com/KDarenskii/order-service/internal/pkg/broker"
)

type (
	OrderDeliverCalculated interface {
		CallbackOrderDeliveryCalculated(
			ctx context.Context, ev *entity.EventOrderDeliveryCalculated, _ []broker.Header,
		) error
	}
)
