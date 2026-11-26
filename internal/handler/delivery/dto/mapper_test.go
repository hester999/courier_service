package dto_test

import (
	"testing"
	"time"

	"avito/internal/handler/delivery/dto"
	"avito/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestRequestAssignToModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   dto.AssignRequest
	}{
		{
			name: "basic",
			in: dto.AssignRequest{
				OrderID: strPtr("ORD-1"),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.RequestAssignToModel(tt.in)

			assert.Equal(t, "ORD-1", out.OrderID)
			assert.Equal(t, model.Courier{}, out.Courier)
		})
	}
}

func TestModelToResponseAssign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *model.Delivery
	}{
		{
			name: "assign response",
			in: func() *model.Delivery {
				transport := "scooter"
				deadline := time.Now()

				return &model.Delivery{
					OrderID:   "ORD-9",
					CourierID: 123,
					Deadline:  deadline,
					Courier: model.Courier{
						TransportType: &transport,
					},
				}
			}(),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.ModelToResponseAssign(tt.in)

			assert.Equal(t, int64(123), out.CourierID)
			assert.Equal(t, "ORD-9", out.OrderID)
			assert.Equal(t, "scooter", out.TransportType)
			assert.Equal(t, tt.in.Deadline, out.DeliveryDeadline)
		})
	}
}

func TestModelToResponseUnassign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *model.Delivery
	}{
		{
			name: "unassign response",
			in: &model.Delivery{
				CourierID: 777,
				Status:    "paused",
				OrderID:   "ORD-55",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			out := dto.ModelToResponseUnassign(tt.in)

			assert.Equal(t, int64(777), out.CourierId)
			assert.Equal(t, "paused", out.Status)
			assert.Equal(t, "ORD-55", out.OrderId)
		})
	}
}

func strPtr(s string) *string { return &s }
