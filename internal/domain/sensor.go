package domain

import "time"

type SensorData struct {
	ID             int       `json:"id"`
	TemperaturaDHT float64   `json:"temperatura_dht"`
	Humedad        float64   `json:"humedad"`
	TemperaturaBMP float64   `json:"temperatura_bmp"`
	Presion        float64   `json:"presion"`
	CalidadAire    float64   `json:"calidad_aire"`
	GasInflamable  float64   `json:"gas_inflamable"`
	Timestamp      time.Time `json:"timestamp"`
}

type SensorReading struct {
	ESP32ID   int       `json:"esp32_id"`
	Value     float64   `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}
