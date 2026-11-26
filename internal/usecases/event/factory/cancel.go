package factory

import (
	"context"

	"avito/internal/model"
	"avito/internal/usecases/delivery"
	"avito/pkg"
)

type canceledHandle struct {
	delivery *delivery.DeliveryUsecase
	log      pkg.Logger
}

func newCanceledHandler(delivery *delivery.DeliveryUsecase, log pkg.Logger) *canceledHandle {
	return &canceledHandle{delivery: delivery, log: log}

}

func (h *canceledHandle) Handle(ctx context.Context, orderID string) error {
	h.log.Info("cancel order %s", orderID)
	deliv := &model.Delivery{OrderID: orderID}
	_, err := h.delivery.UnassignCourier(ctx, deliv)
	if err != nil {
		return err
	}
	return nil
}
