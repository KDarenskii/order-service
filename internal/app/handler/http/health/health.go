package rhealth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	rhandler "github.com/KDarenskii/order-service/internal/app/handler/http"
)

type handler struct{}

func NewHandler() rhandler.Health {
	return &handler{}
}

func (h handler) LastCheck(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
