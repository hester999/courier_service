package status_check

import (
	"context"
)

//go:generate mockgen -source=contract.go -destination=mocks/checker_mocks.go -package=mocks
type deliveryRepo interface {
	GetOverdueDeliveriesWithBusyCouriers(ctx context.Context) ([]int64, error)
}

type courierRepo interface {
	UpdateCouriers(ctx context.Context, couriersID []int64, status string) error
}

type TxRunner interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

type cache interface {
	Add(key string, val map[int64]struct{})
	Get(key string) (map[int64]struct{}, bool)
	Remove(key string)
	GetEvictChan() <-chan map[int64]struct{}
}
