package dto

import "time"

type Order struct {
	ID        string    `json:"order_id"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}
