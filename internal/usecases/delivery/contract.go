package delivery

import (
	"context"
	"time"

	"avito/internal/model"
)

//go:generate mockgen -source=contract.go -destination=mocks/delivery_mocks.go -package=mocks REPO
type delivery interface {
	AssignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error)
	UnassignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error)
	DeleteDeliveryByOrderID(ctx context.Context, orderID string) error
	AssignCouriers(ctx context.Context, deliveries []*model.Delivery) ([]*model.Delivery, error)
	GetDeliveryByOrderID(ctx context.Context, orderID string) (*model.Delivery, error)
}

type courier interface {
	GetCouriersByStatus(ctx context.Context, status string) ([]*model.Courier, error)
	IncrementTotalOrder(ctx context.Context, courierID int64) error
	GetCouriersWithMinimalOrders(ctx context.Context) ([]*model.Courier, error)
	UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error)
	UpdateCouriers(ctx context.Context, courierIDs []int64, status string) error
}

type TxRunner interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type DeadlineFactory interface {
	GetDeadlineByTransport(transport string) (time.Time, error)
}

type gatewayOrder interface {
	Close() error
	GetNewOrders(ctx context.Context, from time.Time) ([]model.Order, error)
}
