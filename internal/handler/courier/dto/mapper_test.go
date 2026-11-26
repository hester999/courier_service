package dto_test

import (
	"testing"

	"avito/internal/handler/courier/dto"
	"avito/internal/model"

	"github.com/stretchr/testify/assert"
)

func strPtr(s string) *string { return &s }

func TestCouriersResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []*model.Courier
	}{
		{
			name: "two couriers",
			in: []*model.Courier{
				{
					ID:            1,
					Name:          "Ivan",
					Phone:         "123",
					Status:        "available",
					TransportType: strPtr("car"),
				},
				{
					ID:            2,
					Name:          "Petr",
					Phone:         "456",
					Status:        "busy",
					TransportType: strPtr("scooter"),
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt // copy
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.CouriersResponse(tt.in)

			assert.Len(t, out, 2)

			assert.Equal(t, int64(1), out[0].ID)
			assert.Equal(t, "Ivan", out[0].Name)
			assert.Equal(t, "123", out[0].Phone)
			assert.Equal(t, "available", out[0].Status)
			assert.Equal(t, "car", out[0].TransportType)

			assert.Equal(t, int64(2), out[1].ID)
			assert.Equal(t, "Petr", out[1].Name)
			assert.Equal(t, "456", out[1].Phone)
			assert.Equal(t, "busy", out[1].Status)
			assert.Equal(t, "scooter", out[1].TransportType)
		})
	}
}

func TestCourierResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *model.Courier
	}{
		{
			name: "basic",
			in: &model.Courier{
				ID:            10,
				Name:          "Test",
				Phone:         "999",
				Status:        "paused",
				TransportType: strPtr("car"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.CourierResponse(tt.in)

			assert.Equal(t, int64(10), out.ID)
			assert.Equal(t, "Test", out.Name)
			assert.Equal(t, "999", out.Phone)
			assert.Equal(t, "paused", out.Status)
			assert.Equal(t, "car", out.TransportType)
		})
	}
}

func TestCourierCreateRequestToModelCreateCourier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *dto.CreateCourierRequestDTO
	}{
		{
			name: "map create request",
			in: &dto.CreateCourierRequestDTO{
				Name:          strPtr("Ivan"),
				Phone:         strPtr("123"),
				Status:        strPtr("available"),
				TransportType: strPtr("car"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.CourierCreateRequestToModelCreateCourier(tt.in)

			assert.Equal(t, "Ivan", *out.Name)
			assert.Equal(t, "123", *out.Phone)
			assert.Equal(t, "available", *out.Status)
			assert.Equal(t, "car", *out.TransportType)
		})
	}
}

func TestCourierFullUpdateRequestToModelFullUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *dto.CourierFullUpdateDTO
	}{
		{
			name: "full update",
			in: &dto.CourierFullUpdateDTO{
				ID:            nil,
				Name:          strPtr("Ivan"),
				Phone:         strPtr("123"),
				Status:        strPtr("busy"),
				TransportType: strPtr("scooter"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.CourierFullUpdateRequestToModelFullUpdate(tt.in)

			assert.NotNil(t, out.ID)
			assert.Equal(t, "Ivan", *out.Name)
			assert.Equal(t, "123", *out.Phone)
			assert.Equal(t, "busy", *out.Status)
			assert.Equal(t, "scooter", *out.TransportType)
		})
	}
}
