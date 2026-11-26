package factory

import (
	"context"
)

//go:generate mockgen -source=contract.go -destination=mocks/factory_mock.go -package=mocks
type StatusHandler interface {
	Handle(ctx context.Context, orderID string) error
}
