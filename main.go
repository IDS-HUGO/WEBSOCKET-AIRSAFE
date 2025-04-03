package main

import (
	"log"
	"multi/receive/internal/application"
	"multi/receive/internal/infrastructure"
	"net/http"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No se pudo cargar el archivo .env, usando variables del sistema")
	}

	// Initialize repositories and services
	repository := infrastructure.NewSensorRepository()
	sensorService := application.NewSensorService(repository)
	mqttAdapter := infrastructure.NewMQTTAdapter(sensorService)
	wsAdapter := infrastructure.NewWebSocketAdapter(sensorService)

	// Setup MQTT
	broker := os.Getenv("RABBITMQ_URL")
	topic := os.Getenv("RABBITMQ_QUEUE_IN")

	if broker == "" || topic == "" {
		log.Fatal("❌ ERROR: RABBITMQ_URL o RABBITMQ_QUEUE_IN no están configurados en el .env")
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("COLAEVENTDRIVE")
	opts.SetDefaultPublishHandler(mqttAdapter.MessageHandler)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Error al conectar con el broker MQTT: %v", token.Error())
	}
	defer client.Disconnect(250)

	if token := client.Subscribe(topic, 1, mqttAdapter.MessageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Error al suscribirse al tópico: %v", token.Error())
	}

	// Setup WebSocket
	http.HandleFunc("/ws", wsAdapter.HandleWebSocket)
	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		wsPort = "8080"
	}

	go func() {
		log.Printf("🌐 WebSocket server starting on port %s...", wsPort)
		if err := http.ListenAndServe(":"+wsPort, nil); err != nil {
			log.Fatalf("❌ WebSocket server error: %v", err)
		}
	}()

	log.Println(" [*] ✅ Esperando mensajes en MQTT. Presiona CTRL+C para salir.")
	select {}
}
