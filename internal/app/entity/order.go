package entity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/KDarenskii/order-service/internal/pkg/broker"
)

type OrderStatus string

const (
	OrderStatusPending            OrderStatus = "pending"
	OrderStatusPaid               OrderStatus = "paid"
	OrderStatusConfirmed          OrderStatus = "confirmed"
	OrderStatusDeliveryCalculated OrderStatus = "delivery_calculated"
)

const (
	tableNameOrder     = "orders"
	tableNameOrderItem = "order_items"
)

type Order struct {
	ID            int64       `gorm:"unique;autoIncrement"`
	GUID          uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserGUID      *uuid.UUID  `gorm:"type:uuid"`
	TotalPrice    int64       `gorm:"not null"`
	DeliveryPrice int64       `gorm:"column:delivery_price"`
	Currency      string      `gorm:"not null"`
	Status        OrderStatus `gorm:"not null"`
	CreatedAt     time.Time   `gorm:"not null"`
	UpdatedAt     time.Time   `gorm:"not null"`

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

////////////////////////////////////////////////////////////////////////////////
///// EVENT MODEL //////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

type EventOrderCreated struct {
	OrderGUID  string                  `json:"order_guid"`
	UserGUID   *string                 `json:"user_guid,omitempty"`
	Currency   string                  `json:"currency"`
	TotalPrice int64                   `json:"total_price"`
	Items      []EventOrderCreatedItem `json:"items"`
	CreatedAt  string                  `json:"created_at"`
}

func (e EventOrderCreated) EventId() string {
	return e.OrderGUID
}

func (e EventOrderCreated) MarshalZerologObject(ev *zerolog.Event) {
	ev.Str("order_guid", e.OrderGUID).
		Str("currency", e.Currency).
		Int64("total_price", e.TotalPrice).
		Int("items_count", len(e.Items))
}

type EventOrderCreatedItem struct {
	ProductGUID string `json:"product_guid"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
}

type EventOrderDeliveryCalculated struct {
	OrderGUID     string `json:"order_guid"`
	DeliveryPrice int64  `json:"delivery_price"`
	Currency      string `json:"currency"`
	CalculatedAt  string `json:"calculated_at"`
}

func (e EventOrderDeliveryCalculated) MarshalZerologObject(ev *zerolog.Event) {
	ev.Str("order_guid", e.OrderGUID).
		Int64("delivery_price", e.DeliveryPrice).
		Str("currency", e.Currency)
}

////////////////////////////////////////////////////////////////////////////////
///// EVENT AUXILIARIES ////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

const (
	BrokerHeaderKeyOrderEventType = "type"
	BrokerHeaderKeyOrderEventID   = "event-id"

	BrokerHeaderValueOrderCreated = "order.created"
)

func BrokerHeaderOrderCreatedType() broker.Header {
	return broker.Header{
		Key:   BrokerHeaderKeyOrderEventType,
		Value: BrokerHeaderValueOrderCreated,
	}
}

func BrokerHeaderOrderCreatedEventID() broker.Header {
	return broker.Header{
		Key:   BrokerHeaderKeyOrderEventID,
		Value: uuid.Must(uuid.NewV4()).String(),
	}
}

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
}

type RequestOrderUpdate struct {
	Status OrderStatus `json:"status" binding:"required,oneof=pending paid confirmed delivery_calculated"`
}

type RequestOrderList struct {
	Status   *OrderStatus `json:"status" binding:"omitempty,oneof=pending paid confirmed delivery_calculated"`
	UserGUID *uuid.UUID   `json:"user_guid" binding:"omitempty"`
}

type ResponseOrderItem struct {
	GUID        uuid.UUID `json:"guid"`
	ProductGUID uuid.UUID `json:"product_guid"`
	Quantity    int       `json:"quantity"`
	UnitPrice   int64     `json:"unit_price"`
}

type ResponseOrderCreate struct {
	GUID          uuid.UUID   `json:"guid"`
	UserGUID      *uuid.UUID  `json:"user_guid"`
	TotalPrice    int64       `json:"total_price"`
	DeliveryPrice int64       `json:"delivery_price"`
	Currency      string      `json:"currency"`
	Status        OrderStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`

	Items []ResponseOrderItem `json:"items"`
}

type ResponseOrderGet struct {
	GUID          uuid.UUID   `json:"guid"`
	UserGUID      *uuid.UUID  `json:"user_guid"`
	TotalPrice    int64       `json:"total_price"`
	DeliveryPrice int64       `json:"delivery_price"`
	Currency      string      `json:"currency"`
	Status        OrderStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`

	Items []ResponseOrderItem `json:"items"`
}

type ResponseOrderUpdate struct {
	GUID          uuid.UUID   `json:"guid"`
	Status        OrderStatus `json:"status"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	DeliveryPrice int64       `json:"delivery_price"`

	Items []ResponseOrderItem `json:"items"`
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
