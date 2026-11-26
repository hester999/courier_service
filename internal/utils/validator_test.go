package utils_test

import (
	"testing"

	"avito/internal/utils"

	"github.com/stretchr/testify/assert"
)

func TestValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		limit   *int64
		offset  *int64
		wantErr bool
		wantMsg string
	}{
		{
			name:    "invalid limit",
			limit:   ptrInt(-1),
			offset:  ptrInt(0),
			wantErr: true,
			wantMsg: "invalid pagination parameter limit must be greater than 1",
		},
		{
			name:    "invalid offset",
			limit:   ptrInt(1),
			offset:  ptrInt(-1),
			wantErr: true,
			wantMsg: "invalid pagination parameter offset must be greater or equal than zero",
		},
		{
			name:    "valid limit and offset",
			limit:   ptrInt(1),
			offset:  ptrInt(1),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt // fix variable capture
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := utils.ValidatePagination(tt.limit, tt.offset)

			if tt.wantErr {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.wantMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ptrInt(v int64) *int64 {
	return &v
}
