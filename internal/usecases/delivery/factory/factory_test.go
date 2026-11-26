package factory_test

import (
	"avito/internal/apperr"
	"avito/internal/usecases/delivery/factory"

	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeadLineFactory_GetDeadlineByTransport(t *testing.T) {
	t.Parallel()
	d := factory.NewDeadLineFabric()

	tests := []struct {
		name      string
		transport string
		wantDur   time.Duration
		wantErr   bool
	}{
		{"on_foot", "on_foot", 30 * time.Minute, false},
		{"scooter", "scooter", 15 * time.Minute, false},
		{"car", "car", 5 * time.Minute, false},
		{"invalid", "bike", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := d.GetDeadlineByTransport(tt.transport)
			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, apperr.ErrInvalidTransportType))
			} else {
				assert.NoError(t, err)
				assert.WithinDuration(t, time.Now().Add(tt.wantDur), got, time.Second)
			}
		})
	}
}
