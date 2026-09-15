package sorder

import (
	"context"
	"time"

	"github.com/gofrs/uuid"

	"github.com/KDarenskii/order-service/internal/app/entity"
	"github.com/KDarenskii/order-service/internal/app/repository"
	"github.com/KDarenskii/order-service/internal/app/service"
)

type srv struct {
	repoOrder repository.Order
}

func NewService(repoOrder repository.Order) service.Order {
	return &srv{repoOrder: repoOrder}
}

func (s *srv) Create(ctx context.Context, req entity.RequestOrderCreate) (entity.Order, error) {
	now := time.Now()

	orderGUID := uuid.Must(uuid.NewV4())

	var totalPrice int64

	orderItems := make([]entity.OrderItem, 0, len(req.Items))

	for _, reqOrderItem := range req.Items {
		totalPrice += reqOrderItem.UnitPrice * int64(reqOrderItem.Quantity)

		newOrderItem := entity.OrderItem{
			GUID:        uuid.Must(uuid.NewV4()),
			UnitPrice:   reqOrderItem.UnitPrice,
			Quantity:    reqOrderItem.Quantity,
			ProductGUID: reqOrderItem.ProductGUID,
			OrderGUID:   orderGUID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		orderItems = append(orderItems, newOrderItem)
	}

	order := entity.Order{
		GUID:       orderGUID,
		Currency:   req.Currency,
		Status:     entity.OrderStatusPending,
		TotalPrice: totalPrice,
		UserGUID:   req.UserGUID,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      orderItems,
	}

	if err := s.repoOrder.Create(ctx, order); err != nil {
		return entity.Order{}, err
	}

	return order, nil
}

func (s *srv) Update(ctx context.Context, guid uuid.UUID, req entity.RequestOrderUpdate) (entity.Order, error) {
	var updatedOrder entity.Order

	err := s.repoOrder.InsideTx(ctx, func(ctx context.Context) error {
		order, err := s.repoOrder.GetByGUID(ctx, guid)
		if err != nil {
			return err
		}

		order.UpdatedAt = time.Now()
		order.Status = req.Status

		if err := s.repoOrder.Update(ctx, order); err != nil {
			return err
		}

		updatedOrder = order

		return nil
	})
	if err != nil {
		return entity.Order{}, err
	}

	return updatedOrder, nil
}

func (s *srv) List(ctx context.Context, req entity.RequestOrderList) ([]entity.Order, error) {
	return s.repoOrder.List(ctx, req.Status, req.UserGUID)
}

func (s *srv) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Order, error) {
	return s.repoOrder.GetByGUID(ctx, guid)
}

func (s *srv) Delete(ctx context.Context, guid uuid.UUID) error {
	return s.repoOrder.InsideTx(ctx, func(ctx context.Context) error {
		_, err := s.repoOrder.GetByGUID(ctx, guid)
		if err != nil {
			return err
		}

		return s.repoOrder.Delete(ctx, guid)
	})
}
