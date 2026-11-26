package dto

import "avito/internal/model"

func RequestAssignToModel(request AssignRequest) *model.Delivery {
	return &model.Delivery{
		OrderID: *request.OrderID,
		Courier: model.Courier{},
	}
}

func ModelToResponseAssign(delivery *model.Delivery) AssignResponse {
	return AssignResponse{
		CourierID:        delivery.CourierID,
		OrderID:          delivery.OrderID,
		TransportType:    *delivery.Courier.TransportType,
		DeliveryDeadline: delivery.Deadline,
	}
}

func ModelToResponseUnassign(delivery *model.Delivery) UnassignResponse {
	return UnassignResponse{
		CourierId: delivery.CourierID,
		Status:    delivery.Status,
		OrderId:   delivery.OrderID,
	}
}
