package application

import (
	"multi/receive/internal/domain"
	"multi/receive/internal/ports/output"
	"time"
)

type SensorService struct {
	repository output.SensorRepository
}

func NewSensorService(repository output.SensorRepository) *SensorService {
	return &SensorService{
		repository: repository,
	}
}

func (s *SensorService) ProcessSensorData(data *domain.SensorData) error {
	data.Timestamp = time.Now()

	if err := s.repository.SaveSensorData(data); err != nil {
		return err
	}

	reading := domain.SensorReading{
		ESP32ID:   data.ID,
		Timestamp: data.Timestamp,
	}

	// Process each sensor reading
	readings := map[string]float64{
		"temperatura_dht": data.TemperaturaDHT,
		"humedad":         data.Humedad,
		"temperatura_bmp": data.TemperaturaBMP,
		"presion":         data.Presion,
		"calidad_aire":    data.CalidadAire,
		"gas_inflamable":  data.GasInflamable,
	}

	for sensorType, value := range readings {
		reading.Value = value
		if err := s.repository.SaveReading(&reading, sensorType); err != nil {
			return err
		}
	}

	return nil
}

func (s *SensorService) GetSensorData(id int) (*domain.SensorData, bool) {
	return s.repository.GetSensorData(id)
}
