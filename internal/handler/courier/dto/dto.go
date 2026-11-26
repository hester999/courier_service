package dto

type CourierDTO struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Status        string `json:"status"`
	TransportType string `json:"transport_type"`
}

type CreateCourierRequestDTO struct {
	Name          *string `json:"name"`
	Phone         *string `json:"phone"`
	Status        *string `json:"status"`
	TransportType *string `json:"transport_type"`
}

type CourierFullUpdateDTO struct {
	ID            *int64  `json:"id"`
	Name          *string `json:"name"`
	Phone         *string `json:"phone"`
	Status        *string `json:"status"`
	TransportType *string `json:"transport_type"`
}
