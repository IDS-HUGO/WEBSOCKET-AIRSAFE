package domain

type WebSocketRequest struct {
	ID          int    `json:"id"`
	PhoneNumber string `json:"phone_number"`
}

type WebSocketResponse struct {
	Error string      `json:"error,omitempty"`
	Data  *SensorData `json:"data,omitempty"`
}
