package delivery_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/usecases/delivery"
	"avito/internal/usecases/delivery/mocks"
	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeliveryUsecase_AssignCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         *model.Delivery
		mockSetup     func(*mocks.Mockdelivery, *mocks.Mockcourier, *mocks.MockTxRunner, *mocks.MockDeadlineFactory)
		wantCourierID int64
		wantErr       bool
		wantErrType   error
	}{
		{
			name:  "success",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, cou *mocks.Mockcourier, tx *mocks.MockTxRunner, df *mocks.MockDeadlineFactory) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return([]*model.Courier{{ID: 1, TransportType: stringPtr("car")}}, nil)

						df.EXPECT().
							GetDeadlineByTransport("car").
							Return(time.Now().Add(time.Hour), nil)

						del.EXPECT().
							AssignCourier(gomock.Any(), gomock.Any()).
							Return(&model.Delivery{CourierID: 1}, nil)

						cou.EXPECT().
							UpdateCourier(gomock.Any(), &model.Courier{Status: "busy", ID: 1}).
							Return(&model.Courier{ID: 1, Status: "busy"}, nil)

						return fn(ctx)
					})
			},
			wantCourierID: 1,
			wantErr:       false,
		},
		{
			name:  "no available couriers",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, cou *mocks.Mockcourier, tx *mocks.MockTxRunner, df *mocks.MockDeadlineFactory) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return(nil, &apperr.AppErr{ServiceError: apperr.ErrNotFound})
						return fn(ctx)
					})
			},
			wantErr:     true,
			wantErrType: apperr.ErrAvailableCouriers,
		},
		{
			name:  "unknown DB error",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, cou *mocks.Mockcourier, tx *mocks.MockTxRunner, df *mocks.MockDeadlineFactory) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return(nil, errors.New("unknown"))
						return fn(ctx)
					})
			},
			wantErr:     true,
			wantErrType: errors.New("unknown"),
		},
		{
			name:  "invalid transport type",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, cou *mocks.Mockcourier, tx *mocks.MockTxRunner, df *mocks.MockDeadlineFactory) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return([]*model.Courier{{ID: 1, TransportType: stringPtr("rocket")}}, nil)

						df.EXPECT().
							GetDeadlineByTransport("rocket").
							Return(time.Now(), &apperr.AppErr{
								ServiceError: apperr.ErrInvalidTransportType,
							})

						return fn(ctx)
					})
			},
			wantErr:     true,
			wantErrType: apperr.ErrInvalidTransportType,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDel := mocks.NewMockdelivery(ctrl)
			mockCou := mocks.NewMockcourier(ctrl)
			mockTx := mocks.NewMockTxRunner(ctrl)
			mockDf := mocks.NewMockDeadlineFactory(ctrl)

			tt.mockSetup(mockDel, mockCou, mockTx, mockDf)

			uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, mockDf, nil, pkg.NewLogger(0))

			result, err := uc.AssignCourier(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantCourierID, result.CourierID)
			}
		})
	}
}

func TestDeliveryUsecase_UnassignCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       *model.Delivery
		mockSetup   func(*mocks.Mockdelivery, *mocks.Mockcourier, *mocks.MockTxRunner)
		wantErr     bool
		wantErrType error
	}{
		{
			name:  "success",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, cou *mocks.Mockcourier, tx *mocks.MockTxRunner) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						del.EXPECT().UnassignCourier(gomock.Any(), gomock.Any()).
							Return(&model.Delivery{CourierID: 1}, nil)

						cou.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).
							Return(&model.Courier{ID: 1}, nil)

						del.EXPECT().
							DeleteDeliveryByOrderID(gomock.Any(), "ORD-123").
							Return(nil)

						cou.EXPECT().
							IncrementTotalOrder(gomock.Any(), int64(1)).
							Return(nil)

						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name:  "unassign fails",
			input: &model.Delivery{OrderID: "ORD-123"},
			mockSetup: func(del *mocks.Mockdelivery, _ *mocks.Mockcourier, tx *mocks.MockTxRunner) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						del.EXPECT().
							UnassignCourier(gomock.Any(), gomock.Any()).
							Return(nil, errors.New("unknown"))
						return fn(ctx)
					})
			},

			wantErr:     true,
			wantErrType: errors.New("unknown"),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDel := mocks.NewMockdelivery(ctrl)
			mockCou := mocks.NewMockcourier(ctrl)
			mockTx := mocks.NewMockTxRunner(ctrl)

			tt.mockSetup(mockDel, mockCou, mockTx)

			uc := delivery.NewDeliveryUsecase(mockTx, mockDel, mockCou, nil, nil, pkg.NewLogger(0))

			result, err := uc.UnassignCourier(context.Background(), tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestDeliveryUsecase_AssignCouriersOnOrdersFromGateWay(t *testing.T) {
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name      string
		mockSetup func(
			*mocks.MockgatewayOrder,
			*mocks.Mockcourier,
			*mocks.Mockdelivery,
			*mocks.MockTxRunner,
			*mocks.MockDeadlineFactory,
		)
		wantErr bool
	}{
		{
			name: "success assign batch",
			mockSetup: func(gate *mocks.MockgatewayOrder, cou *mocks.Mockcourier, del *mocks.Mockdelivery, tx *mocks.MockTxRunner, df *mocks.MockDeadlineFactory) {

				gate.EXPECT().
					GetNewOrders(gomock.Any(), gomock.Any()).
					Return([]model.Order{
						{ID: "O1", CreatedAt: now},
					}, nil)

				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {

						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return([]*model.Courier{
								{ID: 1, TransportType: stringPtr("car")},
							}, nil)

						df.EXPECT().
							GetDeadlineByTransport("car").
							Return(time.Now().Add(time.Hour), nil)

						del.EXPECT().
							AssignCouriers(gomock.Any(), gomock.Any()).
							Return([]*model.Delivery{
								{CourierID: 1},
							}, nil)

						cou.EXPECT().
							UpdateCourier(gomock.Any(), &model.Courier{ID: 1, Status: "busy"}).
							Return(&model.Courier{ID: 1, Status: "busy"}, nil)

						return fn(ctx)
					})
			},
			wantErr: false,
		},
		{
			name: "no new orders",
			mockSetup: func(gate *mocks.MockgatewayOrder, _ *mocks.Mockcourier, _ *mocks.Mockdelivery, _ *mocks.MockTxRunner, _ *mocks.MockDeadlineFactory) {
				gate.EXPECT().
					GetNewOrders(gomock.Any(), gomock.Any()).
					Return([]model.Order{}, nil)
			},
			wantErr: false,
		},
		{
			name: "no available couriers",
			mockSetup: func(gate *mocks.MockgatewayOrder, cou *mocks.Mockcourier, _ *mocks.Mockdelivery, tx *mocks.MockTxRunner, _ *mocks.MockDeadlineFactory) {

				gate.EXPECT().
					GetNewOrders(gomock.Any(), gomock.Any()).
					Return([]model.Order{{ID: "O1", CreatedAt: now}}, nil)

				tx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
					func(ctx context.Context, fn func(context.Context) error) error {
						cou.EXPECT().
							GetCouriersWithMinimalOrders(gomock.Any()).
							Return(nil, &apperr.AppErr{ServiceError: apperr.ErrNotFound})
						return fn(ctx)
					})
			},
			wantErr: false, // по логике usecase — не ошибка
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockGate := mocks.NewMockgatewayOrder(ctrl)
			mockDel := mocks.NewMockdelivery(ctrl)
			mockCou := mocks.NewMockcourier(ctrl)
			mockTx := mocks.NewMockTxRunner(ctrl)
			mockDf := mocks.NewMockDeadlineFactory(ctrl)

			tt.mockSetup(mockGate, mockCou, mockDel, mockTx, mockDf)

			uc := delivery.NewDeliveryUsecase(
				mockTx,
				mockDel,
				mockCou,
				mockDf,
				mockGate,
				pkg.NewLogger(0),
			)

			err := uc.AssignCouriersOnOrdersFromGateWay(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeliveryUsecase_CompletedDelivery(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDel := mocks.NewMockdelivery(ctrl)
	mockCou := mocks.NewMockcourier(ctrl)
	mockTx := mocks.NewMockTxRunner(ctrl)

	mockTx.EXPECT().Run(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, fn func(context.Context) error) error {

			mockDel.EXPECT().
				GetDeliveryByOrderID(gomock.Any(), "ORD-1").
				Return(&model.Delivery{CourierID: 10}, nil)

			mockCou.EXPECT().
				UpdateCouriers(gomock.Any(), []int64{10}, "available").
				Return(nil)

			return fn(ctx)
		})

	uc := delivery.NewDeliveryUsecase(
		mockTx,
		mockDel,
		mockCou,
		nil,
		nil,
		pkg.NewLogger(0),
	)

	err := uc.CompletedDelivery(context.Background(), "ORD-1")
	assert.NoError(t, err)
}

func stringPtr(s string) *string { return &s }
