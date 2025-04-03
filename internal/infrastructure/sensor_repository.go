package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"multi/receive/internal/domain"
	"net/http"
	"os"
	"strings"
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

func (r *SensorRepository) SaveReading(reading *domain.SensorReading, sensorType string) error {
	envKey := fmt.Sprintf("API_URL_%s", strings.ToUpper(strings.ReplaceAll(sensorType, "-", "_")))
	apiURL := os.Getenv(envKey)

	if apiURL == "" {
		return fmt.Errorf("❌ ERROR: URL de API no configurada para %s", sensorType)
	}

	jsonData, err := json.Marshal(reading)
	if err != nil {
		return fmt.Errorf("❌ Error al convertir datos a JSON: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("❌ Error creando request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("❌ Error enviando datos: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("❌ Error de API: status code %d para %s", resp.StatusCode, sensorType)
	}

	log.Printf("✅ Datos enviados exitosamente a %s: %.2f", sensorType, reading.Value)
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
