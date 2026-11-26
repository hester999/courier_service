package factory

import (
	"context"

	"avito/internal/usecases/delivery"
	"avito/pkg"
)

type completedHandler struct {
	delivery *delivery.DeliveryUsecase
	log      pkg.Logger
}

func newCompletedHandler(delivery *delivery.DeliveryUsecase, log pkg.Logger) *completedHandler {
	return &completedHandler{delivery: delivery, log: log}
}

func (h *completedHandler) Handle(ctx context.Context, orderID string) error {
	h.log.Info("complete order %s", orderID)
	err := h.delivery.CompletedDelivery(ctx, orderID)
	if err != nil {
		return err
	}
	return nil
}
