package apperr

import "errors"

type AppErr struct {
	Message      string
	ServiceError error
}

func (a *AppErr) Error() string {
	if a.Message != "" {
		return a.Message
	}
	if a.ServiceError != nil {
		return a.ServiceError.Error()
	}
	return "unknown error"
}

func (a *AppErr) Is(target error) bool {
	return errors.Is(a.ServiceError, target)
}

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidStatus = errors.New("invalid status")
	ErrInvalidPhone  = errors.New("invalid phone. Allow only RU or AM phone number")
)

var (
	ErrAvailableCouriers    = errors.New("no available couriers")
	ErrInvalidTransportType = errors.New("invalid transport type")
)
