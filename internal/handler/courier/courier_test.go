package courier_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"avito/internal/apperr"
	"avito/internal/handler/courier"
	mock_usecases "avito/internal/handler/courier/mocks"
	"avito/internal/model"
	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestHandlerCourier_GetById(t *testing.T) {
	t.Parallel()

	type mockBehavior func(s *mock_usecases.MockusecasesCourier, id int64, result *model.Courier, err error)

	testCases := []struct {
		name               string
		idParam            string
		mockBehavior       mockBehavior
		resultCourier      *model.Courier
		resultErr          error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:    "Success",
			idParam: "1",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, id int64, result *model.Courier, err error) {
				s.EXPECT().GetCourierById(gomock.Any(), id).Return(result, err)
			},
			resultCourier: &model.Courier{
				ID:            1,
				Name:          "test",
				Phone:         "+79067893210",
				Status:        "busy",
				TransportType: stringPtr("on_foot"),
			},
			resultErr:          nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"id":1,"name":"test","phone":"+79067893210","status":"busy","transport_type":"on_foot"}`,
		},
		{
			name:               "Invalid ID param",
			idParam:            "abc",
			mockBehavior:       func(_ *mock_usecases.MockusecasesCourier, _ int64, _ *model.Courier, _ error) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"id is must be integer","code":400}`,
		},
		{
			name:          "Courier not found",
			idParam:       "999",
			resultCourier: nil,
			resultErr:     &apperr.AppErr{Message: "not found", ServiceError: apperr.ErrNotFound},
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, id int64, result *model.Courier, err error) {
				s.EXPECT().GetCourierById(gomock.Any(), id).Return(result, err)
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"message":"not found","code":404}`,
		},
		{
			name:          "Internal error",
			idParam:       "2",
			resultCourier: nil,
			resultErr: &apperr.AppErr{
				Message:      "internal server error",
				ServiceError: errors.New("internal server error"),
			},
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, id int64, result *model.Courier, err error) {
				s.EXPECT().GetCourierById(gomock.Any(), id).Return(result, err)
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal server error","code":500}`,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_usecases.NewMockusecasesCourier(ctrl)
			handler := courier.NewHandlerCourier(mockUC, pkg.NewLogger(0))

			if tt.mockBehavior != nil {
				id, _ := strconv.ParseInt(tt.idParam, 10, 64)
				tt.mockBehavior(mockUC, id, tt.resultCourier, tt.resultErr)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/v1/courier/"+tt.idParam, nil)
			req = mux.SetURLVars(req, map[string]string{"id": tt.idParam})
			w := httptest.NewRecorder()

			handler.GetCourier(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestHandlerCourier_GetCouriers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		query              string
		mockBehavior       func(s *mock_usecases.MockusecasesCourier)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:  "Success with pagination params",
			query: "limit=5&offset=10",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier) {
				s.EXPECT().GetCouriers(gomock.Any(), int64Ptr(5), int64Ptr(10)).
					Return([]*model.Courier{
						{
							ID:            1,
							Name:          "Alice",
							Phone:         "+79991234567",
							Status:        "available",
							TransportType: stringPtr("car"),
						},
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":1,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"car"}]`,
		},
		{
			name:  "No couriers found - returns 204 No Content",
			query: "limit=10&offset=100",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier) {
				s.EXPECT().GetCouriers(gomock.Any(), int64Ptr(10), int64Ptr(100)).
					Return(nil, &apperr.AppErr{
						Message:      "no couriers",
						ServiceError: apperr.ErrNotFound,
					})
			},
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
		},
		{
			name:  "Default limit and offset when not provided",
			query: "",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier) {
				defaultLimit := int64(10)
				defaultOffset := int64(0)

				s.EXPECT().
					GetCouriers(gomock.Any(), &defaultLimit, &defaultOffset).
					Return([]*model.Courier{
						{
							ID:            2,
							Name:          "Bob",
							Phone:         "+79991234567",
							Status:        "available",
							TransportType: stringPtr("bike"),
						},
					}, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":2,"name":"Bob","phone":"+79991234567","status":"available","transport_type":"bike"}]`,
		},
		{
			name:               "Invalid limit - not a number",
			query:              "limit=lalala&offset=10",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid limit parameter","code":400}`,
		},
		{
			name:               "Invalid offset - not a number",
			query:              "limit=10&offset=hahaha",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid offset parameter","code":400}`,
		},
		{
			name:               "Invalid limit - negative",
			query:              "limit=-5&offset=0",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid pagination parameter limit must be greater than 1","code":400}`,
		},
		{
			name:               "Invalid offset - negative",
			query:              "limit=10&offset=-1",
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid pagination parameter offset must be greater or equal than zero","code":400}`,
		},
		{
			name:  "Internal server error",
			query: "limit=10&offset=0",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier) {
				s.EXPECT().GetCouriers(gomock.Any(), int64Ptr(10), int64Ptr(0)).
					Return(nil, &apperr.AppErr{Message: "db down", ServiceError: errors.New("connection refused")})
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"db down","code":500}`,
		},
		{
			name:  "Internal server error unknown",
			query: "limit=10&offset=0",
			mockBehavior: func(s *mock_usecases.MockusecasesCourier) {
				s.EXPECT().GetCouriers(gomock.Any(), int64Ptr(10), int64Ptr(0)).
					Return(nil, errors.New("unknown err"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal server error","code":500}`,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_usecases.NewMockusecasesCourier(ctrl)
			handler := courier.NewHandlerCourier(mockUC, pkg.NewLogger(0))

			if tt.mockBehavior != nil {
				tt.mockBehavior(mockUC)
			}

			path := "/api/v1/couriers"
			if tt.query != "" {
				path += "?" + tt.query
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			w := httptest.NewRecorder()

			handler.GetCouriers(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)

			body := strings.TrimSpace(w.Body.String())
			if tt.expectedBody == "" {
				assert.Empty(t, body)
			} else {
				assert.JSONEq(t, tt.expectedBody, body)
			}
		})
	}
}

func TestHandlerCourier_CreateCourier(t *testing.T) {
	t.Parallel()

	type mockBehavior func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error)

	testCases := []struct {
		name               string
		body               string
		mockBehavior       mockBehavior
		resultCourier      *model.Courier
		resultErr          error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Invalid JSON body",
			body:               `{`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
		{
			name:               "Validation error - name required",
			body:               `{"phone":"+79991234567","status":"available","transport_type":"bike"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"name is required","code":400}`,
		},
		{
			name:               "Validation error - phone required",
			body:               `{"name":"Alice","status":"available","transport_type":"bike"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"phone is required","code":400}`,
		},
		{
			name:               "Validation error - status required",
			body:               `{"name":"Alice","phone":"+79991234567","transport_type":"bike"}`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"status is required","code":400}`,
		},
		{
			name: "Success",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier: &model.Courier{
				ID:            10,
				Name:          "Alice",
				Phone:         "+79991234567",
				Status:        "available",
				TransportType: stringPtr("bike"),
			},
			resultErr:          nil,
			expectedStatusCode: http.StatusCreated,
			expectedBody:       `{"id":10,"name":"Alice","phone":"+79991234567","status":"available","transport_type":"bike"}`,
		},
		{
			name: "Already exists -> 409",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "already exists", ServiceError: apperr.ErrAlreadyExists},
			expectedStatusCode: http.StatusConflict,
			expectedBody:       `{"message":"already exists","code":409}`,
		},
		{
			name: "Invalid phone from usecase -> 400",
			body: `{"name":"Alice","phone":"123","status":"available","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid phone", ServiceError: apperr.ErrInvalidPhone},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid phone","code":400}`,
		},
		{
			name: "Invalid status from usecase -> 400",
			body: `{"name":"Alice","phone":"+79991234567","status":"nope","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid status", ServiceError: apperr.ErrInvalidStatus},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid status","code":400}`,
		},
		{
			name: "Invalid transport type from usecase -> 400",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"hoverboard"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid transport type", ServiceError: apperr.ErrInvalidTransportType},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid transport type","code":400}`,
		},
		{
			name: "Unknown error (not AppErr) -> 500",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          errors.New("boom"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal error","code":500}`,
		},
		{
			name: "AppErr but unknown ServiceError -> 500 internal error",
			body: `{"name":"Alice","phone":"+79991234567","status":"available","transport_type":"bike"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().CreateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "db down", ServiceError: errors.New("connection refused")},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal error","code":500}`,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_usecases.NewMockusecasesCourier(ctrl)
			handler := courier.NewHandlerCourier(mockUC, pkg.NewLogger(0))

			if tt.mockBehavior != nil {
				tt.mockBehavior(mockUC, tt.resultCourier, tt.resultErr)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/couriers", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.CreateCourier(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func TestHandlerCourier_UpdateCourier(t *testing.T) {
	t.Parallel()

	type mockBehavior func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error)

	testCases := []struct {
		name               string
		body               string
		mockBehavior       mockBehavior
		resultCourier      *model.Courier
		resultErr          error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:               "Invalid JSON body",
			body:               `{`,
			mockBehavior:       nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
		{
			name: "Success",
			body: `{"id":1,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier: &model.Courier{
				ID:            1,
				Name:          "Bob",
				Phone:         "+79990000000",
				Status:        "busy",
				TransportType: stringPtr("car"),
			},
			resultErr:          nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"id":1,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"car"}`,
		},
		{
			name: "Not found -> 404",
			body: `{"id":999,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "not found", ServiceError: apperr.ErrNotFound},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"message":"not found","code":404}`,
		},
		{
			name: "Invalid status -> 400",
			body: `{"id":1,"name":"Bob","phone":"+79990000000","status":"nope","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid status", ServiceError: apperr.ErrInvalidStatus},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid status","code":400}`,
		},
		{
			name: "Invalid phone -> 400",
			body: `{"id":1,"name":"Bob","phone":"123","status":"busy","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid phone", ServiceError: apperr.ErrInvalidPhone},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid phone","code":400}`,
		},
		{
			name: "Invalid transport type -> 400",
			body: `{"id":1,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"hoverboard"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "invalid transport type", ServiceError: apperr.ErrInvalidTransportType},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid transport type","code":400}`,
		},
		{
			name: "Unknown error (not AppErr) -> 500",
			body: `{"id":1,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          errors.New("boom"),
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal error","code":500}`,
		},
		{
			name: "AppErr but unknown ServiceError -> 500 internal error",
			body: `{"id":1,"name":"Bob","phone":"+79990000000","status":"busy","transport_type":"car"}`,
			mockBehavior: func(s *mock_usecases.MockusecasesCourier, result *model.Courier, err error) {
				s.EXPECT().UpdateCourier(gomock.Any(), gomock.Any()).Return(result, err)
			},
			resultCourier:      nil,
			resultErr:          &apperr.AppErr{Message: "db down", ServiceError: errors.New("connection refused")},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal error","code":500}`,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_usecases.NewMockusecasesCourier(ctrl)
			handler := courier.NewHandlerCourier(mockUC, pkg.NewLogger(0))

			if tt.mockBehavior != nil {
				tt.mockBehavior(mockUC, tt.resultCourier, tt.resultErr)
			}

			req := httptest.NewRequest(http.MethodPut, "/api/v1/couriers", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			handler.UpdateCourier(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			assert.JSONEq(t, tt.expectedBody, strings.TrimSpace(w.Body.String()))
		})
	}
}

func int64Ptr(i int64) *int64 { return &i }

func stringPtr(s string) *string { return &s }
