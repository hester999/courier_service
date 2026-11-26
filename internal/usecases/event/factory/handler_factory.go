package factory

import (
	"avito/internal/usecases/delivery"
	"avito/pkg"
)

type HandlerFactory struct {
	deliveryUsecases *delivery.DeliveryUsecase
	log              pkg.Logger
}

func NewHandlerFactory(delivery *delivery.DeliveryUsecase, log pkg.Logger) *HandlerFactory {
	return &HandlerFactory{
		deliveryUsecases: delivery,
		log:              log,
	}
}

const (
	created   = "created"
	cancelled = "cancelled"
	completed = "completed"
	canceled  = "canceled"
)

func (h *HandlerFactory) Get(status string) (StatusHandler, error) {
	switch status {
	case cancelled, canceled:
		return newCanceledHandler(h.deliveryUsecases, h.log), nil
	case completed:
		return newCompletedHandler(h.deliveryUsecases, h.log), nil
	default:
		return newCreatedHandler(h.deliveryUsecases, h.log), nil
	}
}
