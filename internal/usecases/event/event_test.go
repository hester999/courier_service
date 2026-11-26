package event_test

import (
	"context"
	"errors"
	"testing"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/usecases/event"
	"avito/internal/usecases/event/factory"
	"avito/internal/usecases/event/mocks"
	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type statusHandlerStub struct {
	handleFn func(ctx context.Context, orderID string) error
}

func (s *statusHandlerStub) Handle(ctx context.Context, orderID string) error {
	if s.handleFn != nil {
		return s.handleFn(ctx, orderID)
	}
	return nil
}

var _ factory.StatusHandler = (*statusHandlerStub)(nil)

func TestEvent_Handle_InvalidStatus_ReturnsNilAndNoCalls(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := mocks.NewMockhandlerFactory(ctrl)
	g := mocks.NewMockgateWay(ctrl)
	logger := pkg.NewLogger(0)
	uc := event.NewEvent(f, g, logger)

	err := uc.Handle(context.Background(), &model.Order{ID: "1", Status: "unknown-status"})
	assert.NoError(t, err)

}

func TestEvent_Handle_GatewayError_ReturnsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := mocks.NewMockhandlerFactory(ctrl)
	g := mocks.NewMockgateWay(ctrl)

	gwErr := errors.New("gateway down")

	g.EXPECT().
		GetOrderByID(gomock.Any(), "1").
		Return(nil, gwErr)
	logger := pkg.NewLogger(0)
	uc := event.NewEvent(f, g, logger)

	err := uc.Handle(context.Background(), &model.Order{ID: "1", Status: "created"})
	assert.ErrorIs(t, err, gwErr)
}

func TestEvent_Handle_FactoryGetError_WrapsAsAppErr(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := mocks.NewMockhandlerFactory(ctrl)
	g := mocks.NewMockgateWay(ctrl)

	getErr := errors.New("no handler")

	g.EXPECT().
		GetOrderByID(gomock.Any(), "1").
		Return(&model.Order{ID: "1", Status: "completed"}, nil)

	f.EXPECT().
		Get("completed").
		Return(nil, getErr)
	logger := pkg.NewLogger(0)
	uc := event.NewEvent(f, g, logger)

	err := uc.Handle(context.Background(), &model.Order{ID: "1", Status: "created"})
	assert.Error(t, err)

	var appErr *apperr.AppErr
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "Error getting handler", appErr.Message)
	assert.ErrorIs(t, err, getErr)
}

func TestEvent_Handle_HandlerError_WrapsAsAppErr(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := mocks.NewMockhandlerFactory(ctrl)
	g := mocks.NewMockgateWay(ctrl)

	handleErr := errors.New("handler failed")

	g.EXPECT().
		GetOrderByID(gomock.Any(), "1").
		Return(&model.Order{ID: "1", Status: "cancelled"}, nil)

	h := &statusHandlerStub{
		handleFn: func(ctx context.Context, orderID string) error {
			if orderID != "1" {
				return errors.New("unexpected order id")
			}
			return handleErr
		},
	}

	f.EXPECT().
		Get("cancelled").
		Return(h, nil)
	logger := pkg.NewLogger(0)
	uc := event.NewEvent(f, g, logger)

	err := uc.Handle(context.Background(), &model.Order{ID: "1", Status: "created"})
	assert.Error(t, err)

	var appErr *apperr.AppErr
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, "Error handling order", appErr.Message)
	assert.ErrorIs(t, err, handleErr)
}

func TestEvent_Handle_Success_CallsHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	f := mocks.NewMockhandlerFactory(ctrl)
	g := mocks.NewMockgateWay(ctrl)

	g.EXPECT().
		GetOrderByID(gomock.Any(), "42").
		Return(&model.Order{ID: "42", Status: "created"}, nil)

	called := false
	h := &statusHandlerStub{
		handleFn: func(ctx context.Context, orderID string) error {
			called = true
			if orderID != "42" {
				return errors.New("unexpected order id")
			}
			return nil
		},
	}

	f.EXPECT().
		Get("created").
		Return(h, nil)
	logger := pkg.NewLogger(0)
	uc := event.NewEvent(f, g, logger)

	err := uc.Handle(context.Background(), &model.Order{ID: "42", Status: "created"})
	assert.NoError(t, err)
	assert.True(t, called, "expected handler to be called")
}
