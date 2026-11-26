package courier_test // используется пакет _test. Для инкапсуляции тестов.

import (
	"context"
	"errors"
	"testing"

	"avito/internal/apperr"
	"avito/internal/model"
	"avito/internal/usecases/courier"
	mock_repo "avito/internal/usecases/courier/mocks"
	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func stringPtr(s string) *string { return &s }
func int64Ptr(i int64) *int64    { return &i }

func TestUsecaseCourier_CreateCourier(t *testing.T) {
	t.Parallel() // теперь используется t.Parallel(). Тесты проверены на гонки. Запускал с флагом go test test_name.go -race. В мейке обновил цель и флаги

	tests := []struct {
		name           string
		req            *model.CreateCourierRequest
		mockSetup      func(*mock_repo.MockcourierRepo)
		wantCourier    *model.Courier
		wantErr        error
		wantErrMessage string
	}{
		{
			name: "success",
			req: &model.CreateCourierRequest{
				Name:          stringPtr("Ivan"),
				Phone:         stringPtr("+79991234567"),
				Status:        stringPtr("available"),
				TransportType: stringPtr("car"),
			},
			mockSetup: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().
					CreateCourier(gomock.Any(), gomock.Any()).
					Return(&model.Courier{ID: 1, Name: "Ivan"}, nil)
			},
			wantCourier: &model.Courier{ID: 1, Name: "Ivan"},
		},
		{
			name: "invalid phone (not RU/AM)",
			req: &model.CreateCourierRequest{
				Name:   stringPtr("Ivan"),
				Phone:  stringPtr("+12025550123"),
				Status: stringPtr("available"),
			},
			mockSetup:      func(m *mock_repo.MockcourierRepo) {},
			wantErr:        apperr.ErrInvalidPhone,
			wantErrMessage: "invalid phone",
		},
		{
			name: "status is nil",
			req: &model.CreateCourierRequest{
				Name:  stringPtr("Ivan"),
				Phone: stringPtr("+79991234567"),
			},
			mockSetup:      func(m *mock_repo.MockcourierRepo) {},
			wantErr:        apperr.ErrInvalidStatus,
			wantErrMessage: "invalid status",
		},
		{
			name: "invalid status",
			req: &model.CreateCourierRequest{
				Name:   stringPtr("Ivan"),
				Phone:  stringPtr("+79991234567"),
				Status: stringPtr("fired"),
			},
			mockSetup:      func(m *mock_repo.MockcourierRepo) {},
			wantErr:        apperr.ErrInvalidStatus,
			wantErrMessage: "invalid status",
		},
		{
			name: "invalid transport type",
			req: &model.CreateCourierRequest{
				Name:          stringPtr("Ivan"),
				Phone:         stringPtr("+79991234567"),
				Status:        stringPtr("available"),
				TransportType: stringPtr("rocket"),
			},
			mockSetup:      func(m *mock_repo.MockcourierRepo) {},
			wantErr:        apperr.ErrInvalidTransportType,
			wantErrMessage: "invalid transport type",
		},
		{
			name: "duplicate phone",
			req: &model.CreateCourierRequest{
				Name:          stringPtr("Ivan"),
				Phone:         stringPtr("+79991234567"),
				Status:        stringPtr("available"),
				TransportType: stringPtr("car"),
			},
			mockSetup: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().
					CreateCourier(gomock.Any(), gomock.Any()).
					Return(nil, &apperr.AppErr{
						Message:      "courier with this phone already exists",
						ServiceError: apperr.ErrAlreadyExists,
					})
			},
			wantErr:        apperr.ErrAlreadyExists,
			wantErrMessage: "exists",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repo.NewMockcourierRepo(ctrl)
			uc := courier.NewUsecasesCourier(mockRepo, pkg.NewLogger(0))

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			got, err := uc.CreateCourier(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
				var appErr *apperr.AppErr
				assert.True(t, errors.As(err, &appErr))
				assert.True(t, appErr.Is(tt.wantErr))
				if tt.wantErrMessage != "" {
					assert.Contains(t, appErr.Message, tt.wantErrMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.wantCourier.ID, got.ID)
			}
		})
	}
}

func TestUsecaseCourier_UpdateCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		req       *model.UpdateCourierRequest
		mockSetup func(*mock_repo.MockcourierRepo)
		wantErr   error
	}{
		{
			name: "success - update name",
			req:  &model.UpdateCourierRequest{ID: int64Ptr(1), Name: stringPtr("Petr")},
			mockSetup: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(&model.Courier{ID: 1}, nil)
			},
		},
		{
			name:      "invalid phone",
			req:       &model.UpdateCourierRequest{ID: int64Ptr(1), Phone: stringPtr("+12025550123")},
			mockSetup: func(m *mock_repo.MockcourierRepo) {},
			wantErr:   apperr.ErrInvalidPhone,
		},
		{
			name:      "invalid status",
			req:       &model.UpdateCourierRequest{ID: int64Ptr(1), Status: stringPtr("dead")},
			mockSetup: func(m *mock_repo.MockcourierRepo) {},
			wantErr:   apperr.ErrInvalidStatus,
		},
		{
			name: "not found",
			req:  &model.UpdateCourierRequest{ID: int64Ptr(999)},
			mockSetup: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).
					Return(nil, &apperr.AppErr{ServiceError: apperr.ErrNotFound})
			},
			wantErr: apperr.ErrNotFound,
		},
		{
			name: "duplicate phone",
			req:  &model.UpdateCourierRequest{ID: int64Ptr(1), Phone: stringPtr("79031245678")},
			mockSetup: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(nil, &apperr.AppErr{ServiceError: apperr.ErrAlreadyExists})
			},
			wantErr: apperr.ErrAlreadyExists,
		},
		{
			name:    "invalid transport type",
			req:     &model.UpdateCourierRequest{ID: int64Ptr(1), TransportType: stringPtr("rocket")},
			wantErr: apperr.ErrInvalidTransportType,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repo.NewMockcourierRepo(ctrl)
			uc := courier.NewUsecasesCourier(mockRepo, pkg.NewLogger(0))

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			got, err := uc.UpdateCourier(context.Background(), tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
				var appErr *apperr.AppErr
				assert.True(t, errors.As(err, &appErr))
				assert.True(t, appErr.Is(tt.wantErr))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestUsecaseCourier_GetCourierById(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      int64
		mock    func(*mock_repo.MockcourierRepo)
		wantErr error
	}{
		{
			name: "success",
			id:   1,
			mock: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().GetCourierById(gomock.Any(), int64(1)).
					Return(&model.Courier{ID: 1}, nil)
			},
		},
		{
			name: "not found",
			id:   999,
			mock: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().GetCourierById(gomock.Any(), int64(999)).
					Return(nil, &apperr.AppErr{ServiceError: apperr.ErrNotFound})
			},
			wantErr: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repo.NewMockcourierRepo(ctrl)
			uc := courier.NewUsecasesCourier(mockRepo, pkg.NewLogger(0))

			tt.mock(mockRepo)

			got, err := uc.GetCourierById(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
			}
		})
	}
}

func TestUsecaseCourier_GetCouriers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limit   *int64
		offset  *int64
		mock    func(*mock_repo.MockcourierRepo)
		wantLen int
		wantErr error
	}{
		{
			name:   "success",
			limit:  int64Ptr(10),
			offset: int64Ptr(0),
			mock: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().
					GetCouriers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return([]*model.Courier{{ID: 1}}, nil)
			},
			wantLen: 1,
		},
		{
			name:   "no couriers",
			limit:  nil,
			offset: nil,
			mock: func(m *mock_repo.MockcourierRepo) {
				m.EXPECT().
					GetCouriers(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, &apperr.AppErr{ServiceError: apperr.ErrNotFound})
			},
			wantErr: apperr.ErrNotFound,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_repo.NewMockcourierRepo(ctrl)
			uc := courier.NewUsecasesCourier(mockRepo, pkg.NewLogger(0))

			tt.mock(mockRepo)

			got, err := uc.GetCouriers(context.Background(), tt.limit, tt.offset)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}
