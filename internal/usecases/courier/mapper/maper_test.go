package mapper

import (
	"avito/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCourierCreateToCourier(t *testing.T) {
	t.Parallel() // Enable parallel execution for the test
	excepted := &model.Courier{
		Name:          "t",
		Phone:         "t",
		Status:        "t",
		TransportType: strPtr("t"),
	}
	input := &model.CreateCourierRequest{
		Name:          strPtr("t"),
		Phone:         strPtr("t"),
		Status:        strPtr("t"),
		TransportType: strPtr("t"),
	}
	res := CourierCreateToCourier(input)

	if !assert.Equal(t, excepted, res) {
		t.Fail()
	}
}

func TestCourierUpdateToCourier(t *testing.T) {
	t.Parallel() // Enable parallel execution for the test
	tests := []struct {
		name   string
		input  *model.UpdateCourierRequest
		assert func(t *testing.T, out *model.Courier)
	}{
		{
			name: "All fields provided",
			input: &model.UpdateCourierRequest{
				ID:            intPtr(10),
				Name:          strPtr("Ivan"),
				Phone:         strPtr("123"),
				Status:        strPtr("available"),
				TransportType: strPtr("car"),
			},
			assert: func(t *testing.T, out *model.Courier) {
				assert.Equal(t, int64(10), out.ID)
				assert.Equal(t, "Ivan", out.Name)
				assert.Equal(t, "123", out.Phone)
				assert.Equal(t, "available", out.Status)
				assert.NotNil(t, out.TransportType)
				assert.Equal(t, "car", *out.TransportType)
			},
		},
		{
			name: "Nil fields get default zero values",
			input: &model.UpdateCourierRequest{
				ID:            nil,
				Name:          nil,
				Phone:         nil,
				Status:        nil,
				TransportType: nil,
			},
			assert: func(t *testing.T, out *model.Courier) {
				assert.Equal(t, int64(0), out.ID)
				assert.Equal(t, "", out.Name)
				assert.Equal(t, "", out.Phone)
				assert.Equal(t, "", out.Status)
				assert.NotNil(t, out.TransportType)
				assert.Equal(t, "", *out.TransportType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Enable parallel execution for subtests
			out := CourierUpdateToCourier(tt.input)
			tt.assert(t, out)
		})
	}
}

func intPtr(v int64) *int64   { return &v }
func strPtr(v string) *string { return &v }
