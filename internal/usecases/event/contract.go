package event

import (
	"context"

	"avito/internal/model"
	"avito/internal/usecases/event/factory"
)

//go:generate mockgen -source=contract.go -destination=mocks/event_mock.go -package=mocks
type handlerFactory interface {
	Get(status string) (factory.StatusHandler, error)
}

type gateWay interface {
	GetOrderByID(ctx context.Context, id string) (*model.Order, error)
}
