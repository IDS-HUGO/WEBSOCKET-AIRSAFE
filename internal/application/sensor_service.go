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
	return s.repository.SaveSensorData(data)
}

func (s *SensorService) GetSensorData(id int) (*domain.SensorData, bool) {
	return s.repository.GetSensorData(id)
}
