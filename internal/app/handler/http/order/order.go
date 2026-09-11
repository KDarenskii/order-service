package horder

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"

	"github.com/KDarenskii/order-service/internal/app/entity"
	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
	"github.com/KDarenskii/order-service/internal/app/service"
	"github.com/KDarenskii/order-service/internal/pkg/http/httph"
)

type handler struct {
	srv service.Order
}

func NewHandler(srv service.Order) rhandler.Order {
	return &handler{srv: srv}
}

func (h *handler) Create(c *gin.Context) {
	var req entity.RequestOrderCreate

	if err := c.ShouldBindJSON(&req); err != nil {
		httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
		return
	}

	order, err := h.srv.Create(c.Request.Context(), req)
	if err != nil {
		httph.HandleError(c.Writer, c.Request, err)
		return
	}

	orderItemsResponse := mapOrderItemsToResponseOrderItems(order.Items)

	response := entity.ResponseOrderCreate{
		GUID:       order.GUID,
		TotalPrice: order.TotalPrice,
		Currency:   order.Currency,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
		Items:      orderItemsResponse,
	}

	httph.SendJSON(c.Writer, http.StatusCreated, response)
}

func (h *handler) Update(c *gin.Context) {
	guid, err := uuid.FromString(c.Param("guid"))
	if err != nil {
		httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
		return
	}

	var req entity.RequestOrderUpdate

	if err := c.ShouldBindJSON(&req); err != nil {
		httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
		return
	}

	order, err := h.srv.Update(c.Request.Context(), guid, req)
	if err != nil {
		httph.HandleError(c.Writer, c.Request, err)
		return
	}

	response := entity.ResponseOrderUpdate{
		GUID:      order.GUID,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}

	httph.SendJSON(c.Writer, http.StatusOK, response)
}

func (h *handler) Delete(c *gin.Context) {
	guid, err := uuid.FromString(c.Param("guid"))
	if err != nil {
		httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
		return
	}

	err = h.srv.Delete(c.Request.Context(), guid)
	if err != nil {
		httph.HandleError(c.Writer, c.Request, err)
		return
	}

	httph.SendEmpty(c.Writer, http.StatusOK)
}

func (h *handler) GetByGUID(c *gin.Context) {
	guid, err := uuid.FromString(c.Param("guid"))
	if err != nil {
		httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
		return
	}

	order, err := h.srv.GetByGUID(c.Request.Context(), guid)
	if err != nil {
		httph.HandleError(c.Writer, c.Request, err)
		return
	}

	orderItemsResponse := mapOrderItemsToResponseOrderItems(order.Items)

	response := entity.ResponseOrderGet{
		GUID:       order.GUID,
		TotalPrice: order.TotalPrice,
		Currency:   order.Currency,
		Status:     order.Status,
		CreatedAt:  order.CreatedAt,
		UserGUID:   order.UserGUID,
		UpdatedAt:  order.UpdatedAt,
		Items:      orderItemsResponse,
	}

	httph.SendJSON(c.Writer, http.StatusOK, response)
}

func (h *handler) List(c *gin.Context) {
	var req entity.RequestOrderList

	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			httph.HandleError(c.Writer, c.Request, entity.ErrIncorrectParameters)
			return
		}
	}

	orders, err := h.srv.List(c.Request.Context(), req)
	if err != nil {
		httph.HandleError(c.Writer, c.Request, err)
		return
	}

	responseOrders := make([]entity.ResponseOrderListItem, 0, len(orders))

	for _, order := range orders {
		responseOrder := entity.ResponseOrderListItem{
			GUID:       order.GUID,
			UserGUID:   order.UserGUID,
			TotalPrice: order.TotalPrice,
			Currency:   order.Currency,
			Status:     order.Status,
			CreatedAt:  order.CreatedAt,
			UpdatedAt:  order.UpdatedAt,
		}

		responseOrders = append(responseOrders, responseOrder)
	}

	response := entity.ResponseOrderList{
		Data: responseOrders,
	}

	httph.SendJSON(c.Writer, http.StatusOK, response)
}

func mapOrderItemsToResponseOrderItems(orderItems []entity.OrderItem) []entity.ResponseOrderItem {
	orderItemsResponse := make([]entity.ResponseOrderItem, 0, len(orderItems))

	for _, orderItem := range orderItems {
		orderItemResponse := entity.ResponseOrderItem{
			GUID:        orderItem.GUID,
			ProductGUID: orderItem.ProductGUID,
			Quantity:    orderItem.Quantity,
			UnitPrice:   orderItem.UnitPrice,
		}

		orderItemsResponse = append(orderItemsResponse, orderItemResponse)
	}

	return orderItemsResponse
}
