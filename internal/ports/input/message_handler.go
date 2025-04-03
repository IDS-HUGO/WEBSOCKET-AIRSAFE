package input

import "multi/receive/internal/domain"

type MessageHandler interface {
	HandleMessage(payload string) error
	HandleWebSocketRequest(id int) (*domain.SensorData, error)
}
