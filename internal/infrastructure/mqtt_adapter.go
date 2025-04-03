package infrastructure

import (
	"encoding/json"
	"log"
	"multi/receive/internal/application"
	"multi/receive/internal/domain"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTAdapter struct {
	sensorService *application.SensorService
}

func NewMQTTAdapter(sensorService *application.SensorService) *MQTTAdapter {
	return &MQTTAdapter{
		sensorService: sensorService,
	}
}

func (m *MQTTAdapter) MessageHandler(client MQTT.Client, msg MQTT.Message) {
	payload := string(msg.Payload())

	log.Printf("📩 Mensaje recibido desde el tópico: '%s'", msg.Topic())

	if payload == "" {
		log.Println("⚠️ Advertencia: Se recibió un mensaje vacío desde MQTT")
		return
	}

	var sensorData domain.SensorData
	if err := json.Unmarshal([]byte(payload), &sensorData); err != nil {
		log.Printf("❌ Error al procesar JSON: %v", err)
		return
	}

	if err := m.sensorService.ProcessSensorData(&sensorData); err != nil {
		log.Printf("❌ Error processing sensor data: %v", err)
	}
}
