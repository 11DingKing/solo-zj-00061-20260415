package handlers

import (
	errors2 "errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rameshsunkara/go-rest-api-example/internal/db"
	errors3 "github.com/rameshsunkara/go-rest-api-example/internal/errors"
	"github.com/rameshsunkara/go-rest-api-example/internal/middleware"
	"github.com/rameshsunkara/go-rest-api-example/internal/models/external"
	"github.com/rameshsunkara/go-rest-api-example/internal/services"
	"github.com/rameshsunkara/go-rest-api-example/pkg/logger"
)

const (
	OrderIDPath = "id"
	MaxPageSize = 100
)

type OrdersHandler struct {
	svc    services.OrdersService
	logger logger.Logger
}

func NewOrdersHandler(lgr logger.Logger, svc services.OrdersService) (*OrdersHandler, error) {
	if lgr == nil || svc == nil {
		return nil, errors2.New("missing required parameters to create orders handler")
	}
	return &OrdersHandler{svc: svc, logger: lgr}, nil
}

func (o *OrdersHandler) Create(c *gin.Context) {
	lgr, requestID := o.logger.WithReqID(c)
	var orderInput external.OrderInput
	if err := c.ShouldBindJSON(&orderInput); err != nil {
		apiErr := errors3.NewValidationError(
			errors3.OrderCreateInvalidInput,
			"Invalid order request body",
			requestID,
			err,
		)
		middleware.HandleError(c, apiErr)
		return
	}

	order, err := o.svc.Create(c, &orderInput, requestID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	middleware.HandleSuccess(c, http.StatusCreated, order)
}

func (o *OrdersHandler) GetAll(c *gin.Context) {
	lgr, requestID := o.logger.WithReqID(c)
	limit, err := o.parseLimitQueryParam(c, requestID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	orders, err := o.svc.GetAll(c, limit, requestID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	middleware.HandleSuccess(c, http.StatusOK, orders)
}

func (o *OrdersHandler) GetByID(c *gin.Context) {
	lgr, requestID := o.logger.WithReqID(c)
	id := c.Param(OrderIDPath)

	order, err := o.svc.GetByID(c, id, requestID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	middleware.HandleSuccess(c, http.StatusOK, order)
}

func (o *OrdersHandler) DeleteByID(c *gin.Context) {
	lgr, requestID := o.logger.WithReqID(c)
	id := c.Param(OrderIDPath)

	err := o.svc.DeleteByID(c, id, requestID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	middleware.HandleSuccess(c, http.StatusNoContent, nil)
}

func (o *OrdersHandler) parseLimitQueryParam(c *gin.Context, requestID string) (int64, error) {
	lgr, _ := o.logger.WithReqID(c)
	l := db.DefaultPageSize
	if input, exists := c.GetQuery("limit"); exists && input != "" {
		val, err := strconv.Atoi(input)
		if err != nil || val < 1 || val > MaxPageSize {
			apiErr := errors3.NewValidationError(
				"",
				fmt.Sprintf("Integer value within 1 and %d is expected for limit query param", MaxPageSize),
				requestID,
				err,
			)
			lgr.Error().
				Int("HttpStatusCode", apiErr.StatusCode).
				Str("ErrorCode", apiErr.ErrorCode).
				Msg(apiErr.Message)
			return 0, apiErr
		}
		l = val
	}
	return int64(l), nil
}
