package entity

import (
	"time"

	"github.com/gofrs/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
)

const (
	tableNameOrder     = "orders"
	tableNameOrderItem = "order_items"
)

type Order struct {
	ID         int64       `gorm:"unique;autoIncrement"`
	GUID       uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserGUID   *uuid.UUID  `gorm:"type:uuid"`
	TotalPrice int64       `gorm:"not null"`
	Currency   string      `gorm:"not null"`
	Status     OrderStatus `gorm:"not null"`
	CreatedAt  time.Time   `gorm:"not null"`
	UpdatedAt  time.Time   `gorm:"not null"`

	Items []OrderItem `gorm:"foreignKey:OrderGUID;references:GUID"`
}

func (Order) TableName() string { return tableNameOrder }

type OrderItem struct {
	ID          int64     `gorm:"unique;autoIncrement"`
	GUID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	OrderGUID   uuid.UUID `gorm:"type:uuid;not null"`
	ProductGUID uuid.UUID `gorm:"type:uuid;not null"`
	Quantity    int       `gorm:"not null"`
	UnitPrice   int64     `gorm:"not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}

func (OrderItem) TableName() string { return tableNameOrderItem }

// //////////////////////////////////////////////////////////////////////////////
// /// HTTP REQUEST & RESPONSE //////////////////////////////////////////////////
// //////////////////////////////////////////////////////////////////////////////

type RequestOrderCreate struct {
	UserGUID *uuid.UUID               `json:"user_guid"`
	Currency string                   `json:"currency" binding:"required,len=3"`
	Items    []RequestOrderItemCreate `json:"items"    binding:"required,min=1,max=100,dive"`
}

type RequestOrderItemCreate struct {
	ProductGUID uuid.UUID `json:"product_guid" binding:"required"`
	Quantity    int       `json:"quantity"     binding:"required,gt=0,lte=1000"`
	UnitPrice   int64     `json:"unit_price"   binding:"required,gt=0,lte=100000000"`
}

type RequestOrderUpdate struct {
	Status OrderStatus `json:"status" binding:"required,oneof=pending paid shipped delivered cancelled"`
}

type RequestOrderList struct {
	Status   *OrderStatus `json:"status" binding:"omitempty,oneof=pending paid shipped delivered cancelled"`
	UserGUID *uuid.UUID   `json:"user_guid" binding:"omitempty"`
}

type ResponseOrderItem struct {
	GUID        uuid.UUID `json:"guid"`
	ProductGUID uuid.UUID `json:"product_guid"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`
}

type ResponseOrderCreate struct {
	GUID       uuid.UUID   `json:"guid"`
	TotalPrice int64       `json:"total_price"`
	Currency   string      `json:"currency"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`

	Items []ResponseOrderItem `json:"items"`
}

type ResponseOrderGet struct {
	GUID       uuid.UUID   `json:"guid"`
	UserGUID   *uuid.UUID  `json:"user_guid"`
	TotalPrice int64       `json:"total_price"`
	Currency   string      `json:"currency"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`

	Items []ResponseOrderItem `json:"items"`
}

type ResponseOrderUpdate struct {
	GUID      uuid.UUID   `json:"guid"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type ResponseOrderListItem struct {
	GUID       uuid.UUID   `json:"guid"`
	UserGUID   *uuid.UUID  `json:"user_guid"`
	TotalPrice int64       `json:"total_price"`
	Currency   string      `json:"currency"`
	Status     OrderStatus `json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type ResponseOrderList struct {
	Data []ResponseOrderListItem `json:"data"`
}
