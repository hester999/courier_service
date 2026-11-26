package delivery

import (
	"avito/internal/model"
	"context"
)

//go:generate mockgen -source=contract.go -destination=mocks/delivery_mocks.go -package=mocks Usecases
type deliveryUsecase interface {
	AssignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error)
	UnassignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error)
}
