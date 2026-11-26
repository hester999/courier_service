package courier

import (
	"context"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/usecases/courier/mapper"
	"avito/pkg"

	"github.com/nyaruka/phonenumbers"
)

type UsecasesCourier struct {
	repo   courierRepo
	logger pkg.Logger
}

func NewUsecasesCourier(repo courierRepo, logger pkg.Logger) *UsecasesCourier {
	return &UsecasesCourier{repo: repo, logger: logger}
}

func (u *UsecasesCourier) GetCouriers(ctx context.Context, limit, offset *int64) ([]*model.Courier, error) {
	return u.repo.GetCouriers(ctx, limit, offset)
}

func (u *UsecasesCourier) GetCourierById(ctx context.Context, id int64) (*model.Courier, error) {
	return u.repo.GetCourierById(ctx, id)
}

func (u *UsecasesCourier) CreateCourier(ctx context.Context, courier *model.CreateCourierRequest) (*model.Courier, error) {
	phone, err := phonenumbers.Parse(*courier.Phone, "RU")
	if err != nil ||
		(!phonenumbers.IsValidNumberForRegion(phone, "RU") &&
			!phonenumbers.IsValidNumberForRegion(phone, "AM")) {

		return nil, &apperr.AppErr{
			Message:      apperr.ErrInvalidPhone.Error(),
			ServiceError: apperr.ErrInvalidPhone,
		}
	}
	if courier.Status != nil {
		if !u.checkStaus(*courier.Status) {
			return nil, &apperr.AppErr{
				Message:      "invalid status",
				ServiceError: apperr.ErrInvalidStatus,
			}
		}
	} else {
		return nil, &apperr.AppErr{
			Message:      "invalid status",
			ServiceError: apperr.ErrInvalidStatus,
		}
	}

	if courier.TransportType != nil {
		if !u.checkTransportType(*courier.TransportType) {
			return nil, &apperr.AppErr{
				Message:      "invalid transport type",
				ServiceError: apperr.ErrInvalidTransportType,
			}
		}
	}

	return u.repo.CreateCourier(ctx, mapper.CourierCreateToCourier(courier))
}

func (u *UsecasesCourier) UpdateCourier(ctx context.Context, courier *model.UpdateCourierRequest) (*model.Courier, error) {
	if courier.Phone != nil {
		phone, err := phonenumbers.Parse(*courier.Phone, "RU")
		if err != nil ||
			(!phonenumbers.IsValidNumberForRegion(phone, "RU") &&
				!phonenumbers.IsValidNumberForRegion(phone, "AM")) {

			return nil, &apperr.AppErr{
				Message:      apperr.ErrInvalidPhone.Error(),
				ServiceError: apperr.ErrInvalidPhone,
			}
		}
	}
	if courier.Status != nil {
		if !u.checkStaus(*courier.Status) {
			return nil, &apperr.AppErr{
				Message:      "invalid status",
				ServiceError: apperr.ErrInvalidStatus,
			}
		}
	}
	if courier.TransportType != nil {
		if !u.checkTransportType(*courier.TransportType) {
			return nil, &apperr.AppErr{
				Message:      "invalid transport type",
				ServiceError: apperr.ErrInvalidTransportType,
			}
		}
	}

	return u.repo.UpdateCourier(ctx, mapper.CourierUpdateToCourier(courier))
}

const (
	availableStatus = "available"
	busyStatus      = "busy"
	pausedStatus    = "paused"
	carType         = "car"
	scooterType     = "scooter"
	footType        = "on_foot"
)

func (u *UsecasesCourier) checkStaus(status string) bool {
	switch status {
	case availableStatus, busyStatus, pausedStatus:
		return true
	default:
		return false
	}
}

func (u *UsecasesCourier) checkTransportType(transport string) bool {
	m := map[string]bool{
		scooterType: true,
		footType:    true,
		carType:     true,
	}
	return m[transport]
}
