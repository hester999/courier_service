package mapper

import "avito/internal/model"

func CourierCreateToCourier(courier *model.CreateCourierRequest) *model.Courier {
	return &model.Courier{
		Name:          *courier.Name,
		Phone:         *courier.Phone,
		Status:        *courier.Status,
		TransportType: courier.TransportType,
	}
}

func CourierUpdateToCourier(courier *model.UpdateCourierRequest) *model.Courier {
	if courier.ID == nil {
		courier.ID = new(int64)
	}
	if courier.Name == nil {
		courier.Name = new(string)
	}
	if courier.Phone == nil {
		courier.Phone = new(string)
	}
	if courier.Status == nil {
		courier.Status = new(string)
	}
	if courier.TransportType == nil {
		courier.TransportType = new(string)
	}

	return &model.Courier{
		Name:          *courier.Name,
		Phone:         *courier.Phone,
		Status:        *courier.Status,
		ID:            *courier.ID,
		TransportType: courier.TransportType,
	}
}
