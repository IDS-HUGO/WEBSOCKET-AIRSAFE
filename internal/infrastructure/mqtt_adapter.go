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
	wsAdapter     *WebSocketAdapter
}

func NewMQTTAdapter(sensorService *application.SensorService, wsAdapter *WebSocketAdapter) *MQTTAdapter {
	return &MQTTAdapter{
		sensorService: sensorService,
		wsAdapter:     wsAdapter,
	}
}

func (m *MQTTAdapter) MessageHandler(client MQTT.Client, msg MQTT.Message) {
	payload := string(msg.Payload())

	log.Printf("📩 Mensaje recibido desde MQTT: '%s'", msg.Topic())
	log.Printf("📦 Payload recibido: %s", payload)

	if payload == "" {
		log.Println("⚠️ Mensaje vacío recibido, ignorando...")
		return
	}

	var sensorData domain.SensorData
	if err := json.Unmarshal([]byte(payload), &sensorData); err != nil {
		log.Printf("❌ Error procesando JSON: %v", err)
		return
	}

	// Print parsed data in a pretty format
	prettyJSON, err := json.MarshalIndent(sensorData, "", "    ")
	if err == nil {
		log.Printf("📊 Datos del sensor procesados:\n%s", string(prettyJSON))
	}

	// Save to cache/local
	if err := m.sensorService.ProcessSensorData(&sensorData); err != nil {
		log.Printf("❌ Error guardando datos del sensor: %v", err)
	}

	// Send to WebSocket
	m.wsAdapter.BroadcastSensorData(&sensorData)
}
