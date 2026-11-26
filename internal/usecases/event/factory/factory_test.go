// internal/usecases/event/factory/factory_handlers_test.go
package factory_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"avito/internal/model"
	"avito/internal/usecases/delivery"
	dmocks "avito/internal/usecases/delivery/mocks"
	"avito/internal/usecases/event/factory"
	"avito/pkg"
)

func TestHandlerFactory_Get_Routing(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDel := dmocks.NewMockdelivery(ctrl)
	mockCou := dmocks.NewMockcourier(ctrl)
	mockTx := dmocks.NewMockTxRunner(ctrl)
	mockDf := dmocks.NewMockDeadlineFactory(ctrl)

	logger := pkg.NewLogger(0)
	uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, mockDf, nil, logger)
	f := factory.NewHandlerFactory(uc, logger)

	h1, err := f.Get("cancelled")
	assert.NoError(t, err)
	assert.NotNil(t, h1)

	h2, err := f.Get("canceled")
	assert.NoError(t, err)
	assert.NotNil(t, h2)

	h3, err := f.Get("completed")
	assert.NoError(t, err)
	assert.NotNil(t, h3)

	h4, err := f.Get("created")
	assert.NoError(t, err)
	assert.NotNil(t, h4)

	h5, err := f.Get("unknown-status")
	assert.NoError(t, err)
	assert.NotNil(t, h5)
}

func TestCreatedHandler_Handle_CallsAssignCourierFlow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDel := dmocks.NewMockdelivery(ctrl)
	mockCou := dmocks.NewMockcourier(ctrl)
	mockTx := dmocks.NewMockTxRunner(ctrl)
	mockDf := dmocks.NewMockDeadlineFactory(ctrl)
	logger := pkg.NewLogger(0)

	mockTx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {

			mockCou.EXPECT().
				GetCouriersWithMinimalOrders(gomock.Any()).
				Return([]*model.Courier{{ID: 1, TransportType: stringPtr("car")}}, nil)

			mockDf.EXPECT().
				GetDeadlineByTransport("car").
				Return(time.Now().Add(time.Hour), nil)

			mockDel.EXPECT().
				AssignCourier(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, d *model.Delivery) (*model.Delivery, error) {
					assert.Equal(t, "ORD-1", d.OrderID)
					assert.Equal(t, int64(1), d.CourierID)
					return &model.Delivery{OrderID: d.OrderID, CourierID: d.CourierID}, nil
				})

			mockCou.EXPECT().
				UpdateCourier(gomock.Any(), &model.Courier{Status: "busy", ID: 1}).
				Return(&model.Courier{ID: 1, Status: "busy"}, nil)

			return fn(ctx)
		},
	)

	uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, mockDf, nil, logger)
	f := factory.NewHandlerFactory(uc, logger)

	h, err := f.Get("created")
	assert.NoError(t, err)

	err = h.Handle(context.Background(), "ORD-1")
	assert.NoError(t, err)
}

func TestCanceledHandler_Handle_CallsUnassignCourierFlow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDel := dmocks.NewMockdelivery(ctrl)
	mockCou := dmocks.NewMockcourier(ctrl)
	mockTx := dmocks.NewMockTxRunner(ctrl)

	mockTx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {

			mockDel.EXPECT().
				UnassignCourier(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, d *model.Delivery) (*model.Delivery, error) {
					assert.Equal(t, "ORD-2", d.OrderID)
					return &model.Delivery{OrderID: d.OrderID, CourierID: 10}, nil
				})

			mockCou.EXPECT().
				UpdateCourier(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, c *model.Courier) (*model.Courier, error) {
					assert.Equal(t, int64(10), c.ID)
					assert.Equal(t, "unassign", c.Status)
					return &model.Courier{ID: c.ID, Status: c.Status}, nil
				})

			mockDel.EXPECT().
				DeleteDeliveryByOrderID(gomock.Any(), "ORD-2").
				Return(nil)

			mockCou.EXPECT().
				IncrementTotalOrder(gomock.Any(), int64(10)).
				Return(nil)

			return fn(ctx)
		},
	)
	logger := pkg.NewLogger(0)
	uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, nil, nil, logger)
	f := factory.NewHandlerFactory(uc, logger)

	h, err := f.Get("cancelled")
	assert.NoError(t, err)

	err = h.Handle(context.Background(), "ORD-2")
	assert.NoError(t, err)
}

func TestCompletedHandler_Handle_CallsCompletedDeliveryFlow(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDel := dmocks.NewMockdelivery(ctrl)
	mockCou := dmocks.NewMockcourier(ctrl)
	mockTx := dmocks.NewMockTxRunner(ctrl)

	mockTx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {

			mockDel.EXPECT().
				GetDeliveryByOrderID(gomock.Any(), "ORD-3").
				Return(&model.Delivery{OrderID: "ORD-3", CourierID: 777}, nil)

			mockCou.EXPECT().
				UpdateCouriers(gomock.Any(), []int64{777}, "available").
				Return(nil)

			return fn(ctx)
		},
	)
	logger := pkg.NewLogger(0)
	uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, nil, nil, logger)
	f := factory.NewHandlerFactory(uc, logger)

	h, err := f.Get("completed")
	assert.NoError(t, err)

	err = h.Handle(context.Background(), "ORD-3")
	assert.NoError(t, err)
}

func stringPtr(s string) *string {
	return &s
}
