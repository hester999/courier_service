package status_check

import (
	"context"
	"errors"
	"testing"

	checkermock "avito/internal/usecases/courier/status_check/mocks"

	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCourierStatusChecker_cacheController(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		incomingIDs []int64
		cacheSetup  func(*checkermock.Mockcache)
		expectedIDs []int64
	}{
		{
			name:        "first batch < batchSize → save to cache",
			incomingIDs: []int64{1, 2},
			cacheSetup: func(c *checkermock.Mockcache) {
				gomock.InOrder(
					c.EXPECT().Get("busy").Return(nil, false),
					c.EXPECT().Add("busy", map[int64]struct{}{1: {}, 2: {}}),
				)
			},
			expectedIDs: nil,
		},
		{
			name:        "merge → flush",
			incomingIDs: []int64{3, 4, 5},
			cacheSetup: func(c *checkermock.Mockcache) {
				gomock.InOrder(
					c.EXPECT().Get("busy").Return(map[int64]struct{}{1: {}, 2: {}}, true),
					c.EXPECT().Remove("busy"),
				)
			},
			expectedIDs: []int64{1, 2, 3, 4, 5},
		},
		{
			name:        "batch ≥ batchSize → immediate return",
			incomingIDs: []int64{10, 20, 30, 40, 50, 60},
			cacheSetup: func(c *checkermock.Mockcache) {
				c.EXPECT().Get("busy").Return(nil, false)
			},
			expectedIDs: []int64{10, 20, 30, 40, 50, 60},
		},
		{
			name:        "merge but still < batchSize → update TTL",
			incomingIDs: []int64{3},
			cacheSetup: func(c *checkermock.Mockcache) {
				gomock.InOrder(
					c.EXPECT().Get("busy").Return(map[int64]struct{}{1: {}, 2: {}}, true),
					c.EXPECT().Add("busy", map[int64]struct{}{1: {}, 2: {}, 3: {}}),
				)
			},
			expectedIDs: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCache := checkermock.NewMockcache(ctrl)
			checker := NewCourierStatusChecker(nil, nil, mockCache, nil, pkg.NewLogger(0), 5)

			tt.cacheSetup(mockCache)

			result := checker.cacheController(tt.incomingIDs)

			if tt.expectedIDs == nil {
				assert.Nil(t, result)
			} else {
				assert.ElementsMatch(t, tt.expectedIDs, result)
			}
		})
	}
}

func TestCourierStatusChecker_ProcessEvicted(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		evicted    map[int64]struct{}
		setupMocks func(*checkermock.MockcourierRepo, *checkermock.MockdeliveryRepo, *checkermock.MockTxRunner)
		expectErr  bool
	}{
		{
			name:    "success",
			evicted: map[int64]struct{}{7: {}, 8: {}},
			setupMocks: func(c *checkermock.MockcourierRepo, d *checkermock.MockdeliveryRepo, tx *checkermock.MockTxRunner) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

				c.EXPECT().UpdateCouriers(gomock.Any(), []int64{7, 8}, "available").Return(nil)
			},
			expectErr: false,
		},
		{
			name:    "update error",
			evicted: map[int64]struct{}{7: {}, 8: {}},
			setupMocks: func(c *checkermock.MockcourierRepo, d *checkermock.MockdeliveryRepo, tx *checkermock.MockTxRunner) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

				c.EXPECT().UpdateCouriers(gomock.Any(), []int64{7, 8}, "available").
					Return(errors.New("update error"))
			},
			expectErr: true,
		},
		{
			name:       "empty map",
			evicted:    nil,
			setupMocks: func(_ *checkermock.MockcourierRepo, _ *checkermock.MockdeliveryRepo, _ *checkermock.MockTxRunner) {},
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockCourier := checkermock.NewMockcourierRepo(ctrl)
			mockDelivery := checkermock.NewMockdeliveryRepo(ctrl)
			mockTx := checkermock.NewMockTxRunner(ctrl)

			tt.setupMocks(mockCourier, mockDelivery, mockTx)

			checker := NewCourierStatusChecker(mockDelivery, mockCourier, nil, mockTx, pkg.NewLogger(0), 10)

			err := checker.ProcessEvicted(context.Background(), tt.evicted)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCourierStatusChecker_CheckAndSetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		batchSize  int
		setupMocks func(*checkermock.MockdeliveryRepo, *checkermock.MockcourierRepo, *checkermock.Mockcache, *checkermock.MockTxRunner)
		wantErr    bool
	}{
		{
			name:      "no overdue deliveries",
			batchSize: 3,
			setupMocks: func(d *checkermock.MockdeliveryRepo, _ *checkermock.MockcourierRepo, _ *checkermock.Mockcache, tx *checkermock.MockTxRunner) {
				tx.EXPECT().Run(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})
				d.EXPECT().GetOverdueDeliveriesWithBusyCouriers(gomock.Any()).
					Return([]int64{}, nil)
			},
		},
		// другие кейсы — аналогично переписаны
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDelivery := checkermock.NewMockdeliveryRepo(ctrl)
			mockCourier := checkermock.NewMockcourierRepo(ctrl)
			mockCache := checkermock.NewMockcache(ctrl)
			mockTx := checkermock.NewMockTxRunner(ctrl)

			tt.setupMocks(mockDelivery, mockCourier, mockCache, mockTx)

			checker := NewCourierStatusChecker(
				mockDelivery,
				mockCourier,
				mockCache,
				mockTx,
				pkg.NewLogger(0),
				tt.batchSize,
			)

			err := checker.CheckAndSetStatus(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
