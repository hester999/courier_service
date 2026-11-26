package delivery

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/pkg"
)

type DeliveryUsecase struct {
	deliveryStorage delivery
	courierStorage  courier
	deadlineFactory DeadlineFactory

	gate gatewayOrder
	tx   TxRunner

	cursor time.Time
	log    pkg.Logger
}

const (
	available = "available"
	unassign  = "unassign"
	busy      = "busy"
)

func NewDeliveryUsecase(tx TxRunner, dr delivery, cr courier, deadline DeadlineFactory, gate gatewayOrder, log pkg.Logger) *DeliveryUsecase {
	return &DeliveryUsecase{tx: tx, deliveryStorage: dr, courierStorage: cr, deadlineFactory: deadline, gate: gate, log: log, cursor: time.Now().Add(-5 * time.Second)}
}

func (d *DeliveryUsecase) AssignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error) {
	var result *model.Delivery

	err := d.tx.Run(ctx, func(ctx context.Context) error {

		availableCouriers, err := d.courierStorage.GetCouriersWithMinimalOrders(ctx)
		if err != nil {
			var appErr *apperr.AppErr
			if errors.As(err, &appErr) {
				if appErr.Is(apperr.ErrNotFound) {
					return &apperr.AppErr{
						Message:      "no available couriers",
						ServiceError: apperr.ErrAvailableCouriers,
					}
				}
			}
			return err
		}
		randomCourier := d.getRandomCourier(availableCouriers)

		delivery.CourierID = randomCourier.ID
		delivery.Courier.TransportType = randomCourier.TransportType

		delivery.Deadline, err = d.deadlineFactory.GetDeadlineByTransport(*delivery.Courier.TransportType)
		if err != nil {
			return &apperr.AppErr{
				Message:      err.Error(),
				ServiceError: err,
			}
		}
		result, err = d.deliveryStorage.AssignCourier(ctx, delivery)
		if err != nil {
			return err
		}

		courierUpdate := &model.Courier{
			Status: busy,
			ID:     delivery.CourierID,
		}

		courierInfo, err := d.courierStorage.UpdateCourier(ctx, courierUpdate)
		if err != nil {
			d.log.Error("failed to update courier %v", err)
			return err
		}
		result.Courier = *courierInfo
		return nil
	})
	if err != nil {
		d.log.Error("failed to assign courier %v", err)
		return nil, err
	}

	return result, nil
}

func (d *DeliveryUsecase) getRandomCourier(couriers []*model.Courier) *model.Courier { // выбирает рандомно курьера из списка доступных

	randomInt := rand.Intn(len(couriers))

	return couriers[randomInt]
}

func (d *DeliveryUsecase) UnassignCourier(ctx context.Context, delivery *model.Delivery) (*model.Delivery, error) {
	var result *model.Delivery
	err := d.tx.Run(ctx, func(ctx context.Context) error {

		tmp, err := d.deliveryStorage.UnassignCourier(ctx, delivery)
		if err != nil {
			d.log.Error("failed to unassign courier:%v", err)
			var appErr *apperr.AppErr
			if errors.As(err, &appErr) {
				if appErr.Is(apperr.ErrNotFound) {
					return &apperr.AppErr{
						Message:      "delivery not found",
						ServiceError: appErr.ServiceError,
					}
				}
			}
			return err
		}

		updateCourier := &model.Courier{
			Status: unassign,
			ID:     tmp.CourierID,
		}

		courierInfo, err := d.courierStorage.UpdateCourier(ctx, updateCourier)
		if err != nil {
			d.log.Error("failed to update courier:%v", err)
			return err
		}

		tmp.Courier = *courierInfo
		tmp.Status = courierInfo.Status

		err = d.deliveryStorage.DeleteDeliveryByOrderID(ctx, delivery.OrderID)
		if err != nil {
			d.log.Error("failed to delete delivery:%v", err)
			return err
		}

		err = d.courierStorage.IncrementTotalOrder(ctx, tmp.Courier.ID)
		if err != nil {
			d.log.Error("failed to increment total order:%v", err)
			return err
		}
		result = tmp
		return err
	})
	if err != nil {
		d.log.Error("failed to unassign courier:%v", err)
		return nil, err
	}
	return result, nil
}

func (d *DeliveryUsecase) AssignCouriersOnOrdersFromGateWay(ctx context.Context) error {
	from := d.cursor
	d.log.Warn("FROM: %v", from)

	orders, err := d.gate.GetNewOrders(ctx, from)
	if err != nil {
		return &apperr.AppErr{
			Message:      "failed to get new orders",
			ServiceError: err,
		}
	}

	if len(orders) == 0 {
		d.log.Info("No new orders found")
		return nil
	}

	nextCursor := from
	for _, o := range orders {
		if o.CreatedAt.After(nextCursor) {
			nextCursor = o.CreatedAt
		}
	}

	d.cursor = nextCursor.Add(time.Nanosecond)
	d.log.Warn("CURSOR UPDATED TO: %v", d.cursor)

	err = d.tx.Run(ctx, func(ctx context.Context) error {

		availableCouriers, err := d.courierStorage.GetCouriersWithMinimalOrders(ctx)
		if err != nil {
			var appErr *apperr.AppErr
			if errors.As(err, &appErr) && appErr.Is(apperr.ErrNotFound) {
				return &apperr.AppErr{
					Message:      "no available couriers",
					ServiceError: apperr.ErrAvailableCouriers,
				}
			}
			return err
		}

		randomDeliveries := d.getRandomCouriersWithOrders(availableCouriers, orders)
		if len(randomDeliveries) == 0 {
			d.log.Info("no deliveries to assign")
			return nil
		}

		for _, v := range randomDeliveries {
			deadline, err := d.deadlineFactory.GetDeadlineByTransport(*v.Courier.TransportType)
			if err != nil {
				return &apperr.AppErr{
					Message:      err.Error(),
					ServiceError: err,
				}
			}
			v.Deadline = deadline
		}

		assigned, err := d.deliveryStorage.AssignCouriers(ctx, randomDeliveries)
		if err != nil {
			return err
		}

		for _, v := range assigned {
			_, err = d.courierStorage.UpdateCourier(ctx, &model.Courier{
				ID:     v.CourierID,
				Status: busy,
			})
			if err != nil {
				d.log.Error("failed to update courier: %v", err)
				return err
			}
		}

		return nil
	})

	if err != nil {
		var appErr *apperr.AppErr
		if errors.As(err, &appErr) && appErr.Is(apperr.ErrAvailableCouriers) {
			d.log.Info(appErr.Message)
			return nil
		}

		d.log.Error("failed to assign couriers: %v", err)
		return err
	}

	d.log.Info("assign couriers success")
	return nil
}

func (d *DeliveryUsecase) getRandomCouriersWithOrders( // получаем готовую пачку заказов. Если курьер уже назначен на заказ, тл будет выбран свободный. Заказы назначаются по порядку
	couriers []*model.Courier,
	orders []model.Order,
) []*model.Delivery {

	deliveries := make([]*model.Delivery, 0)

	if len(couriers) == 0 || len(orders) == 0 {
		return deliveries
	}

	used := make(map[int64]bool)
	orderIndex := 0

	for orderIndex < len(orders) && len(used) < len(couriers) {
		cour := d.getRandomCourier(couriers)

		if used[cour.ID] {
			continue
		}

		deliveries = append(deliveries, &model.Delivery{
			OrderID:   orders[orderIndex].ID,
			CourierID: cour.ID,
			Courier:   *cour,
		})

		used[cour.ID] = true
		orderIndex++
	}

	return deliveries
}

func (d *DeliveryUsecase) CloseGateWay() {
	_ = d.gate.Close()
}

func (d *DeliveryUsecase) CompletedDelivery(ctx context.Context, orderID string) error {
	err := d.tx.Run(ctx, func(ctx context.Context) error {
		var err error

		deliveryInfo, err := d.deliveryStorage.GetDeliveryByOrderID(ctx, orderID)
		if err != nil {
			d.log.Error("failed to get delivery info: %v", err)
			return err
		}

		if err = d.courierStorage.UpdateCouriers(ctx, []int64{deliveryInfo.CourierID}, available); err != nil {
			d.log.Error("failed to update courier: %v", err)
			return err
		}
		return nil
	})

	return err
}
