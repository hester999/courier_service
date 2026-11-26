package delivery_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"avito/internal/apperr"
	"avito/internal/handler/delivery"
	mock_delivery "avito/internal/handler/delivery/mocks"
	"avito/internal/model"
	"avito/pkg"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDeliveryHandler_AssignCourierOnOrder(t *testing.T) {
	t.Parallel()

	fixedNow := time.Date(2025, 12, 25, 15, 0, 0, 0, time.UTC)
	fixedDeadline := fixedNow.Add(3 * time.Hour)

	testCases := []struct {
		name               string
		requestBody        string
		setupMock          func(*mock_delivery.MockdeliveryUsecase)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:        "Success - courier assigned",
			requestBody: `{"order_id": "ORD-777"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				input := &model.Delivery{OrderID: "ORD-777"}
				output := &model.Delivery{
					ID:         10,
					CourierID:  42,
					OrderID:    "ORD-777",
					AssignedAt: fixedNow,
					Deadline:   fixedDeadline,
					Status:     "assigned",
					Courier:    model.Courier{ID: 42, TransportType: stringPtr("car")},
				}
				m.EXPECT().AssignCourier(gomock.Any(), input).Return(output, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"courier_id":42,"order_id":"ORD-777","transport_type":"car","delivery_deadline":"2025-12-25T18:00:00Z"}`,
		},
		{
			name:        "Conflict - already assigned",
			requestBody: `{"order_id": "ORD-001"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().AssignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-001"}).
					Return(nil, &apperr.AppErr{
						Message:      "order already has courier",
						ServiceError: apperr.ErrAlreadyExists,
					})
			},
			expectedStatusCode: http.StatusConflict,
			expectedBody:       `{"message":"order already has courier","code":409}`,
		},
		{
			name:        "Conflict - no available couriers",
			requestBody: `{"order_id": "ORD-002"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().AssignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-002"}).
					Return(nil, &apperr.AppErr{
						Message:      "no available couriers",
						ServiceError: apperr.ErrAvailableCouriers,
					})
			},
			expectedStatusCode: http.StatusConflict,
			expectedBody:       `{"message":"no available couriers","code":409}`,
		},
		{
			name:        "Not Found - order not found",
			requestBody: `{"order_id": "ORD-404"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().AssignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-404"}).
					Return(nil, &apperr.AppErr{
						Message:      "order not found",
						ServiceError: apperr.ErrNotFound,
					})
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"message":"not found order","code":404}`,
		},
		{
			name:               "Bad Request - missing order_id",
			requestBody:        `{}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"empty orderID","code":400}`,
		},
		{
			name:               "Bad Request - null order_id",
			requestBody:        `{"order_id": null}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"empty orderID","code":400}`,
		},
		{
			name:        "Internal Server Error - raw error",
			requestBody: `{"order_id": "ORD-500"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().AssignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-500"}).
					Return(nil, errors.New("db connection lost"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal server error","code":500}`,
		},
		{
			name:               "Bad Request - malformed JSON (unclosed brace)",
			requestBody:        `{"order_id": "ORD-123"`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
		{
			name:               "Bad Request - invalid JSON (wrong syntax)",
			requestBody:        `{"order_id": ORD-123}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
	}

	for _, tt := range testCases {
		tt := tt // COPY for parallel
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_delivery.NewMockdeliveryUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUC)
			}

			handler := delivery.NewDeliveryHandler(mockUC, pkg.NewLogger(3))

			req := httptest.NewRequest(http.MethodPost, "/assign", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.AssignCourierOnOrder(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			body := strings.TrimSpace(w.Body.String())
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func TestDeliveryHandler_UnassignCourierOnOrder(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		requestBody        string
		setupMock          func(*mock_delivery.MockdeliveryUsecase)
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name:        "Success - courier unassigned",
			requestBody: `{"order_id": "ORD-100"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				input := &model.Delivery{OrderID: "ORD-100"}
				output := &model.Delivery{
					OrderID:   "ORD-100",
					CourierID: 42,
					Status:    "unassigned",
				}
				m.EXPECT().UnassignCourier(gomock.Any(), input).Return(output, nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"order_id":"ORD-100","status":"unassigned","courier_id":42}`,
		},
		{
			name:        "Not Found - nothing to unassign",
			requestBody: `{"order_id": "ORD-404"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().UnassignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-404"}).
					Return(nil, &apperr.AppErr{
						Message:      "no assigned courier",
						ServiceError: apperr.ErrNotFound,
					})
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       `{"message":"no assigned courier","code":404}`,
		},
		{
			name:               "Bad Request - missing order_id",
			requestBody:        `{}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"empty orderID","code":400}`,
		},
		{
			name:        "Internal Server Error - raw error",
			requestBody: `{"order_id": "ORD-500"}`,
			setupMock: func(m *mock_delivery.MockdeliveryUsecase) {
				m.EXPECT().UnassignCourier(gomock.Any(), &model.Delivery{OrderID: "ORD-500"}).
					Return(nil, errors.New("boom"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"message":"internal server error","code":500}`,
		},
		{
			name:               "Bad Request - malformed JSON (missing quote)",
			requestBody:        `{"order_id": ORD-456}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
		{
			name:               "Bad Request - invalid JSON (trailing comma)",
			requestBody:        `{"order_id": "ORD-456",}`,
			setupMock:          nil,
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"message":"invalid request body","code":400}`,
		},
	}

	for _, tt := range testCases {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockUC := mock_delivery.NewMockdeliveryUsecase(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(mockUC)
			}

			handler := delivery.NewDeliveryHandler(mockUC, pkg.NewLogger(0))

			req := httptest.NewRequest(http.MethodPost, "/unassign", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			handler.UnassignCourierOnOrder(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			body := strings.TrimSpace(w.Body.String())
			assert.JSONEq(t, tt.expectedBody, body)
		})
	}
}

func stringPtr(str string) *string {
	return &str
}
