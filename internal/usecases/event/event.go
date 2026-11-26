package event

import (
	"context"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/pkg"
)

type Event struct {
	factory handlerFactory
	gateWay gateWay
	log     pkg.Logger
}

const (
	created   = "created"
	cancelled = "cancelled"
	completed = "completed"
	canceled  = "canceled"
)

func NewEvent(factory handlerFactory, gate gateWay, log pkg.Logger) *Event {
	return &Event{
		factory: factory,
		gateWay: gate,
		log:     log,
	}
}

func (e *Event) Handle(ctx context.Context, order *model.Order) error {
	if !e.checkStatus(order.Status) {
		e.log.Warn("order status is invalid %v", order.Status)
		return nil
	}

	actualOrder, err := e.gateWay.GetOrderByID(ctx, order.ID)
	if err != nil {
		e.log.Error("get order id %v error %v", order.ID, err)
		return err
	}
	e.log.Info("get order id %v actual status %v", order.ID, actualOrder.Status)
	handler, err := e.factory.Get(actualOrder.Status)
	if err != nil {
		e.log.Error("Error getting handler for order %v", err)
		return &apperr.AppErr{
			Message:      "Error getting handler",
			ServiceError: err,
		}
	}
	e.log.Info("Handling order %v", order.ID)
	err = handler.Handle(ctx, order.ID)
	if err != nil {
		e.log.Error("Error handling order %v", err)
		return &apperr.AppErr{
			Message:      "Error handling order",
			ServiceError: err,
		}
	}
	return nil
}

func (e *Event) checkStatus(status string) bool {
	var statusMap = map[string]struct{}{
		created:   {},
		cancelled: {},
		completed: {},
		canceled:  {},
	}
	if _, ok := statusMap[status]; !ok {
		return false
	}
	return true
}
