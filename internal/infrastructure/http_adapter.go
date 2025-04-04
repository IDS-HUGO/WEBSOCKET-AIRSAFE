package infrastructure

import (
	"bytes"
	"encoding/json"
	"log"
	"multi/receive/internal/application"
	"multi/receive/internal/domain"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type WebSocketAdapter struct {
	sensorService *application.SensorService
	upgrader      websocket.Upgrader
	clients       map[*websocket.Conn]bool
	mu            sync.Mutex
}

func NewWebSocketAdapter(sensorService *application.SensorService) *WebSocketAdapter {
	return &WebSocketAdapter{
		sensorService: sensorService,
		clients:       make(map[*websocket.Conn]bool),
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

	w.mu.Lock()
	w.clients[conn] = true
	w.mu.Unlock()

	log.Println("✅ Cliente WebSocket conectado")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("❌ Cliente WebSocket desconectado: %v", err)
			w.mu.Lock()
			delete(w.clients, conn)
			w.mu.Unlock()
			break
		}

		var wsRequest domain.WebSocketRequest
		if err := json.Unmarshal(message, &wsRequest); err != nil {
			log.Printf("❌ Error parsing request: %v", err)
			sendErrorResponse(conn, "Invalid request format")
			continue
		}

		log.Printf("📥 Received request for ID: %d with phone: %s", wsRequest.ID, wsRequest.PhoneNumber)

		sensorData, exists := w.sensorService.GetSensorData(wsRequest.ID)
		if !exists {
			log.Printf("⚠️ No sensor data found for ID: %d", wsRequest.ID)
			sendErrorResponse(conn, "Sensor data not found")
			continue
		}

		response := domain.WebSocketResponse{
			Data: sensorData,
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Printf("❌ Error sending response: %v", err)
			continue
		}

		log.Printf("✅ Sent sensor data for ID %d to client", wsRequest.ID)

		go w.sendToExternalAPIs(sensorData, wsRequest.PhoneNumber)
	}
}

func sendErrorResponse(conn *websocket.Conn, message string) {
	response := domain.WebSocketResponse{
		Error: message,
	}
	if err := conn.WriteJSON(response); err != nil {
		log.Printf("❌ Error sending error response: %v", err)
	}
}

type APIPayload struct {
	ID          int     `json:"id"`
	PhoneNumber string  `json:"phone_number"`
	Value       float64 `json:"value"`
}

func (w *WebSocketAdapter) sendToExternalAPIs(data *domain.SensorData, phoneNumber string) {
	endpoints := map[string]float64{
		os.Getenv("API_URL_TEMPERATURA_DHT"): data.TemperaturaDHT,
		os.Getenv("API_URL_HUMEDAD"):         data.Humedad,
		os.Getenv("API_URL_TEMPERATURA_BMP"): data.TemperaturaBMP,
		os.Getenv("API_URL_PRESION"):         data.Presion,
		os.Getenv("API_URL_CALIDAD_AIRE"):    data.CalidadAire,
		os.Getenv("API_URL_GAS_INFLAMABLE"):  data.GasInflamable,
	}

	for url, value := range endpoints {
		if url == "" {
			continue
		}

		payload := APIPayload{
			ID:          data.ID,
			PhoneNumber: phoneNumber,
			Value:       value,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			log.Printf("❌ Error creating payload for %s: %v", url, err)
			continue
		}

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("❌ Error making POST request to %s: %v", url, err)
			continue
		}

		log.Printf("✅ Data sent successfully to %s", url)
		resp.Body.Close()
	}
}

func (w *WebSocketAdapter) BroadcastSensorData(data *domain.SensorData) {
	w.mu.Lock()
	defer w.mu.Unlock()

	payload, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Error al serializar JSON: %v", err)
		return
	}

	for conn := range w.clients {
		err := conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			log.Printf("❌ Error enviando mensaje WebSocket: %v", err)
			conn.Close()
			delete(w.clients, conn)
		}
	}
}
