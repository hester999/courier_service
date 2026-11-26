package factory

import (
	"context"

	"avito/internal/model"
	"avito/internal/usecases/delivery"
	"avito/pkg"
)

type createdHandler struct {
	delivery *delivery.DeliveryUsecase
	log      pkg.Logger
}

func newCreatedHandler(deliv *delivery.DeliveryUsecase, log pkg.Logger) *createdHandler {
	return &createdHandler{
		delivery: deliv,
		log:      log,
	}
}

func (h *createdHandler) Handle(ctx context.Context, orderID string) error {
	h.log.Info("create order %s", orderID)
	deliv := &model.Delivery{
		OrderID: orderID,
	}
	_, err := h.delivery.AssignCourier(ctx, deliv)
	return err
}
