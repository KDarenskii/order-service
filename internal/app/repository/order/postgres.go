package porder

import (
	"context"

	"github.com/gofrs/uuid"

	"github.com/KDarenskii/order-service/internal/app/entity"
	"github.com/KDarenskii/order-service/internal/app/repository"
	rcpostgres "github.com/KDarenskii/order-service/internal/app/repository/conn/postgres"
)

type repoPg struct {
	conn *rcpostgres.Client
}

func NewRepo(client *rcpostgres.Client) repository.Order {
	return &repoPg{conn: client}
}

func (r *repoPg) InsideTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.conn.InsideTx(ctx, fn)
}

func (r *repoPg) Create(ctx context.Context, order entity.Order) error {
	res := r.conn.GetDB(ctx).WithContext(ctx).Create(&order)
	return res.Error
}

func (r *repoPg) Update(ctx context.Context, order entity.Order) error {
	res := r.conn.GetDB(ctx).WithContext(ctx).
		Model(&entity.Order{}).
		Where("guid = ?", order.GUID).
		Updates(map[string]any{
			"status":     order.Status,
			"updated_at": order.UpdatedAt,
		})

	return rcpostgres.NotFoundIfNoRows(res)
}

func (r *repoPg) GetByGUID(ctx context.Context, guid uuid.UUID) (entity.Order, error) {
	var order entity.Order

	res := r.conn.GetDB(ctx).WithContext(ctx).Preload("Items").Where("guid = ?", guid).First(&order)

	if rcpostgres.IsNotFoundError(res) {
		return entity.Order{}, entity.ErrNotFound
	}

	if res.Error != nil {
		return entity.Order{}, res.Error
	}

	return order, nil
}

func (r *repoPg) Delete(ctx context.Context, guid uuid.UUID) error {
	res := r.conn.GetDB(ctx).WithContext(ctx).
		Where("guid = ?", guid).
		Delete(&entity.Order{})

	return rcpostgres.NotFoundIfNoRows(res)
}

func (r *repoPg) List(ctx context.Context, status *entity.OrderStatus, userGUID *uuid.UUID) ([]entity.Order, error) {
	var orders []entity.Order

	query := r.conn.GetDB(ctx).WithContext(ctx)

	if status != nil {
		query = query.Where("status = ?", status)
	}

	if userGUID != nil {
		query = query.Where("user_guid = ?", userGUID)
	}

	res := query.Find(&orders)

	if res.Error != nil {
		return []entity.Order{}, res.Error
	}

	return orders, nil
}
