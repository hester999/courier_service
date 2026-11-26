package kafka

import (
	"context"

	"avito/internal/model"
)

//go:generate mockgen -source=contract.go -destination=mocks/kafka_mocks.go -package=mocks Usecases
type handlerEvent interface {
	Handle(ctx context.Context, order *model.Order) error
}
