package sorder

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog/log"

	"github.com/KDarenskii/order-service/internal/app/entity"
	"github.com/KDarenskii/order-service/internal/app/repository"
	"github.com/KDarenskii/order-service/internal/app/service"
)

type deliverySrv struct {
	repoOrder repository.Order
}

func NewDeliveryService(repoOrder repository.Order) service.OrderDelivery {
	return &deliverySrv{repoOrder: repoOrder}
}

func (s *deliverySrv) UpdateDelivery(ctx context.Context, orderGUID uuid.UUID, deliveryPrice int64) error {
	now := time.Now()

	err := s.repoOrder.UpdateDelivery(ctx, orderGUID, deliveryPrice, now)

	if errors.Is(err, entity.ErrNotFound) {
		log.Warn().Ctx(ctx).Str("order_guid", orderGUID.String()).Msg("Заказ не найден или уже обработан")
		return nil
	}

	return err
}
