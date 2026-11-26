package courier

import (
	"avito/internal/model"
	"context"
)

//go:generate mockgen -source=contract.go -destination=mocks/courier_mocks.go -package=mocks REPO
type courierRepo interface {
	GetCourierById(ctx context.Context, id int64) (*model.Courier, error)
	GetCouriers(ctx context.Context, limit, offset *int64) ([]*model.Courier, error)
	CreateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error)
	UpdateCourier(ctx context.Context, courier *model.Courier) (*model.Courier, error)
}
