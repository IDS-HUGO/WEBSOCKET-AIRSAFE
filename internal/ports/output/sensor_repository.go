package output

import "multi/receive/internal/domain"

type SensorRepository interface {
	SaveSensorData(data *domain.SensorData) error
	SaveReading(reading *domain.SensorReading, sensorType string) error
	GetSensorData(id int) (*domain.SensorData, bool)
}
