package courier

import (
	"avito/internal/model"
	"context"
)

//go:generate mockgen -source=contract.go -destination=mocks/courier_mocks.go -package=mocks Usecases
type usecasesCourier interface {
	GetCouriers(ctx context.Context, limit, offset *int64) ([]*model.Courier, error)
	GetCourierById(ctx context.Context, id int64) (*model.Courier, error)
	CreateCourier(ctx context.Context, courier *model.CreateCourierRequest) (*model.Courier, error)
	UpdateCourier(ctx context.Context, courier *model.UpdateCourierRequest) (*model.Courier, error)
}
