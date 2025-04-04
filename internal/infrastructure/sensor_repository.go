package infrastructure

import (
	"multi/receive/internal/domain"
	"sync"
)

type SensorRepository struct {
	sensorCache    map[int]domain.SensorData
	sensorCacheMux sync.RWMutex
}

func NewSensorRepository() *SensorRepository {
	return &SensorRepository{
		sensorCache: make(map[int]domain.SensorData),
	}
}

func (r *SensorRepository) SaveSensorData(data *domain.SensorData) error {
	r.sensorCacheMux.Lock()
	r.sensorCache[data.ID] = *data
	r.sensorCacheMux.Unlock()
	return nil
}

func (r *SensorRepository) GetSensorData(id int) (*domain.SensorData, bool) {
	r.sensorCacheMux.RLock()
	defer r.sensorCacheMux.RUnlock()

	data, exists := r.sensorCache[id]
	if !exists {
		return nil, false
	}
	return &data, true
}

func (r *SensorRepository) SaveReading(reading *domain.SensorReading, sensorType string) error {
	return nil
}
