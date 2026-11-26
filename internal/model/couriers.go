package model

import "time"

type Courier struct {
	ID            int64
	Name          string
	Phone         string
	Status        string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	TransportType *string
	TotalOrders   *int64
}

type CreateCourierRequest struct {
	Name          *string
	Phone         *string
	Status        *string
	TransportType *string
}

type UpdateCourierRequest struct {
	ID            *int64
	Name          *string
	Phone         *string
	Status        *string
	TransportType *string
}
