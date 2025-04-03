package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"
	"multi/receive/internal/application"
	"net/http"

	"github.com/gorilla/websocket"
)

type WebSocketAdapter struct {
	sensorService *application.SensorService
	upgrader      websocket.Upgrader
}

func NewWebSocketAdapter(sensorService *application.SensorService) *WebSocketAdapter {
	return &WebSocketAdapter{
		sensorService: sensorService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (w *WebSocketAdapter) HandleWebSocket(resp http.ResponseWriter, req *http.Request) {
	conn, err := w.upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Printf("❌ Error upgrading to WebSocket: %v", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Error reading WebSocket message: %v", err)
			break
		}

		var request struct {
			ID int `json:"id"`
		}
		if err := json.Unmarshal(message, &request); err != nil {
			log.Printf("❌ Error parsing WebSocket request: %v", err)
			continue
		}

		sensorData, exists := w.sensorService.GetSensorData(request.ID)
		if !exists {
			conn.WriteJSON(map[string]string{
				"error": fmt.Sprintf("No data found for ESP32 ID: %d", request.ID),
			})
			continue
		}

		if err := conn.WriteJSON(sensorData); err != nil {
			log.Printf("❌ Error sending WebSocket response: %v", err)
			break
		}
	}
}
