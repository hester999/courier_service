package dto

import "time"

type AssignRequest struct {
	OrderID *string `json:"order_id"`
}

type AssignResponse struct {
	CourierID        int64     `json:"courier_id"`
	OrderID          string    `json:"order_id"`
	TransportType    string    `json:"transport_type"`
	DeliveryDeadline time.Time `json:"delivery_deadline"`
}
type UnassignResponse struct {
	OrderId   string `json:"order_id"`
	Status    string `json:"status"`
	CourierId int64  `json:"courier_id"`
}
