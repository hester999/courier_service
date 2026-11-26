package dto

import "avito/internal/model"

func CouriersResponse(courier []*model.Courier) []*CourierDTO {
	res := make([]*CourierDTO, len(courier))

	for i, c := range courier {
		res[i] = &CourierDTO{
			ID:            c.ID,
			Name:          c.Name,
			Phone:         c.Phone,
			Status:        c.Status,
			TransportType: *c.TransportType,
		}
	}
	return res
}

func CourierResponse(courier *model.Courier) *CourierDTO {
	return &CourierDTO{
		ID:            courier.ID,
		Name:          courier.Name,
		Phone:         courier.Phone,
		Status:        courier.Status,
		TransportType: *courier.TransportType,
	}
}

func CourierCreateRequestToModelCreateCourier(courier *CreateCourierRequestDTO) *model.CreateCourierRequest {
	return &model.CreateCourierRequest{
		Name:          courier.Name,
		Phone:         courier.Phone,
		Status:        courier.Status,
		TransportType: courier.TransportType,
	}
}

func CourierFullUpdateRequestToModelFullUpdate(courier *CourierFullUpdateDTO) *model.UpdateCourierRequest {
	if courier.ID == nil {
		courier.ID = new(int64)
	}
	return &model.UpdateCourierRequest{
		ID:            courier.ID,
		Name:          courier.Name,
		Phone:         courier.Phone,
		Status:        courier.Status,
		TransportType: courier.TransportType,
	}
}
