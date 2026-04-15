package services

import (
	"context"
	"errors"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/rameshsunkara/go-rest-api-example/internal/db"
	errors2 "github.com/rameshsunkara/go-rest-api-example/internal/errors"
	"github.com/rameshsunkara/go-rest-api-example/internal/models/data"
	"github.com/rameshsunkara/go-rest-api-example/internal/models/external"
	"github.com/rameshsunkara/go-rest-api-example/internal/utilities"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrdersService interface {
	Create(ctx context.Context, orderInput *external.OrderInput, debugID string) (*external.Order, error)
	GetAll(ctx context.Context, limit int64, debugID string) (*[]external.Order, error)
	GetByID(ctx context.Context, id string, debugID string) (*external.Order, error)
	DeleteByID(ctx context.Context, id string, debugID string) error
}

type OrdersServiceImpl struct {
	repo db.OrdersDataService
}

func NewOrdersService(repo db.OrdersDataService) (OrdersService, error) {
	if repo == nil {
		return nil, errors.New("missing required parameters to create orders service")
	}
	return &OrdersServiceImpl{repo: repo}, nil
}

func (s *OrdersServiceImpl) Create(ctx context.Context, orderInput *external.OrderInput, debugID string) (*external.Order, error) {
	products := make([]data.Product, len(orderInput.Products))
	for i, p := range orderInput.Products {
		products[i] = data.Product{Name: p.Name, Price: p.Price, Quantity: p.Quantity}
	}

	order := data.Order{
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Products:    products,
		User:        faker.Email(),
		TotalAmount: utilities.CalculateTotalAmount(products),
		Status:      data.OrderPending,
	}

	id, err := s.repo.Create(ctx, &order)
	if err != nil {
		return nil, errors2.NewInternalError(
			errors2.OrderCreateServerError,
			errors2.UnexpectedErrorMessage,
			debugID,
			err,
		)
	}

	extOrder := external.Order{
		ID:          id,
		CreatedAt:   utilities.FormatTimeToISO(order.CreatedAt),
		UpdatedAt:   utilities.FormatTimeToISO(order.UpdatedAt),
		Products:    order.Products,
		User:        order.User,
		TotalAmount: order.TotalAmount,
		Status:      order.Status,
		Version:     order.Version,
	}

	return &extOrder, nil
}

func (s *OrdersServiceImpl) GetAll(ctx context.Context, limit int64, debugID string) (*[]external.Order, error) {
	orders, err := s.repo.GetAll(ctx, limit)
	if err != nil {
		return nil, errors2.NewInternalError(
			errors2.OrdersGetServerError,
			errors2.UnexpectedErrorMessage,
			debugID,
			err,
		)
	}

	extOrders := make([]external.Order, 0, len(*orders))
	for _, o := range *orders {
		extOrders = append(extOrders, external.Order{
			ID:          o.ID.Hex(),
			Version:     o.Version,
			Status:      o.Status,
			TotalAmount: o.TotalAmount,
			User:        o.User,
			CreatedAt:   utilities.FormatTimeToISO(o.CreatedAt),
			UpdatedAt:   utilities.FormatTimeToISO(o.UpdatedAt),
			Products:    o.Products,
		})
	}

	return &extOrders, nil
}

func (s *OrdersServiceImpl) GetByID(ctx context.Context, id string, debugID string) (*external.Order, error) {
	oID, err := primitive.ObjectIDFromHex(id)
	if err != nil || oID.IsZero() {
		return nil, errors2.NewValidationError(
			errors2.OrderGetInvalidParams,
			"invalid order ID",
			debugID,
			err,
		)
	}

	order, err := s.repo.GetByID(ctx, oID)
	if err != nil {
		if errors.Is(err, db.ErrPOIDNotFound) {
			return nil, errors2.NewNotFoundError(
				errors2.OrderGetNotFound,
				"order not found",
				debugID,
				err,
			)
		}
		return nil, errors2.NewInternalError(
			errors2.OrdersGetServerError,
			"failed to fetch order",
			debugID,
			err,
		)
	}

	extOrder := external.Order{
		ID:          order.ID.Hex(),
		Version:     order.Version,
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		User:        order.User,
		CreatedAt:   utilities.FormatTimeToISO(order.CreatedAt),
		UpdatedAt:   utilities.FormatTimeToISO(order.UpdatedAt),
		Products:    order.Products,
	}

	return &extOrder, nil
}

func (s *OrdersServiceImpl) DeleteByID(ctx context.Context, id string, debugID string) error {
	oID, err := primitive.ObjectIDFromHex(id)
	if err != nil || oID.IsZero() {
		return errors2.NewValidationError(
			errors2.OrderDeleteInvalidID,
			"invalid order ID",
			debugID,
			err,
		)
	}

	if dbErr := s.repo.DeleteByID(ctx, oID); dbErr != nil {
		if errors.Is(dbErr, db.ErrPOIDNotFound) {
			return errors2.NewNotFoundError(
				errors2.OrderDeleteNotFound,
				"could not delete order",
				debugID,
				dbErr,
			)
		}
		return errors2.NewInternalError(
			errors2.OrderDeleteServerError,
			"could not delete order",
			debugID,
			dbErr,
		)
	}

	return nil
}
