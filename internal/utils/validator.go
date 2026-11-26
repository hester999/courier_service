package utils

import "avito/internal/apperr"

func ValidatePagination(limit, offset *int64) error {

	if limit != nil && *limit < 1 {
		return &apperr.AppErr{
			Message: "invalid pagination parameter limit must be greater than 1",
		}
	}
	if offset != nil && *offset < 0 {
		return &apperr.AppErr{
			Message: "invalid pagination parameter offset must be greater or equal than zero",
		}
	}
	return nil
}
