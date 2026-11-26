package factory

import (
	"avito/internal/apperr"
	"time"
)

type DeadLineFactory struct{}

func NewDeadLineFabric() *DeadLineFactory {
	return &DeadLineFactory{}
}

func (d *DeadLineFactory) GetDeadlineByTransport(t string) (time.Time, error) {

	switch t {
	case "on_foot":
		return time.Now().Add(time.Minute * 30).UTC(), nil
	case "scooter":
		return time.Now().Add(time.Minute * 15).UTC(), nil
	case "car":
		return time.Now().Add(time.Minute * 5).UTC(), nil
	default:
		return time.Now(), apperr.ErrInvalidTransportType
	}
}
