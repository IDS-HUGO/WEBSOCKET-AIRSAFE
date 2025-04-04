package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"
	"multi/receive/internal/application"
	"multi/receive/internal/domain"
	"multi/receive/internal/ports/output"
	"os"
	"strconv"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

type MQTTAdapter struct {
	sensorService *application.SensorService
	wsAdapter     *WebSocketAdapter
	smsService    output.SMSService
}

func NewMQTTAdapter(sensorService *application.SensorService, wsAdapter *WebSocketAdapter, smsService output.SMSService) *MQTTAdapter {
	return &MQTTAdapter{
		sensorService: sensorService,
		wsAdapter:     wsAdapter,
		smsService:    smsService,
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

	var mqttPayload struct {
		SensorID  string `json:"sensor_id"`
		Timestamp string `json:"timestamp"`
		Data      struct {
			TemperaturaDHT float64 `json:"temperatura_dht"`
			HumedadDHT     float64 `json:"humedad_dht"`
			TemperaturaBMP float64 `json:"temperatura_bmp"`
			Presion        float64 `json:"presion"`
			CalidadAire    float64 `json:"calidad_aire"`
			GasInflamable  float64 `json:"gas_inflamable"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(payload), &mqttPayload); err != nil {
		log.Printf("❌ Error procesando JSON: %v", err)
		return
	}

	var sensorID int
	fmt.Sscanf(mqttPayload.SensorID, "sensor_%d", &sensorID)

	sensorData := &domain.SensorData{
		ID:             sensorID,
		TemperaturaDHT: mqttPayload.Data.TemperaturaDHT,
		Humedad:        mqttPayload.Data.HumedadDHT,
		TemperaturaBMP: mqttPayload.Data.TemperaturaBMP,
		Presion:        mqttPayload.Data.Presion,
		CalidadAire:    mqttPayload.Data.CalidadAire,
		GasInflamable:  mqttPayload.Data.GasInflamable,
	}

	prettyJSON, err := json.MarshalIndent(sensorData, "", "    ")
	if err == nil {
		log.Printf("📊 Datos del sensor procesados:\n%s", string(prettyJSON))
	}

	minLevel, _ := strconv.ParseFloat(os.Getenv("GAS_ALERT_MIN"), 64)
	maxLevel, _ := strconv.ParseFloat(os.Getenv("GAS_ALERT_MAX"), 64)

	if sensorData.GasInflamable >= minLevel && sensorData.GasInflamable <= maxLevel {
		alertMsg := fmt.Sprintf("⚠️ ALERTA: Nivel de gas inflamable peligroso detectado: %.2f PPM en el sensor %d",
			sensorData.GasInflamable, sensorData.ID)
		phoneNumber := os.Getenv("ALERT_PHONE_NUMBER")

		if err := m.smsService.SendAlert(phoneNumber, alertMsg); err != nil {
			log.Printf("❌ Error sending alert SMS: %v", err)
		} else {
			log.Printf("✅ Alert SMS sent to %s for gas level %.2f PPM", phoneNumber, sensorData.GasInflamable)
		}
	}

	if err := m.sensorService.ProcessSensorData(sensorData); err != nil {
		log.Printf("❌ Error guardando datos del sensor: %v", err)
		return
	}

	log.Printf("✅ Datos guardados correctamente para ID: %d", sensorData.ID)
	m.wsAdapter.BroadcastSensorData(sensorData)
}
