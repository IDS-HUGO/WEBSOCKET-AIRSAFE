package main

import (
	"log"
	"multi/receive/internal/application"
	"multi/receive/internal/infrastructure"
	"net/http"
	"os"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No se pudo cargar .env, usando variables del sistema")
	}

	repository := infrastructure.NewSensorRepository()
	sensorService := application.NewSensorService(repository)

	wsAdapter := infrastructure.NewWebSocketAdapter(sensorService)

	smsService := infrastructure.NewTwilioSMSService()
	mqttAdapter := infrastructure.NewMQTTAdapter(sensorService, wsAdapter, smsService)

	broker := os.Getenv("RABBITMQ_URL")
	topic := os.Getenv("RABBITMQ_QUEUE_IN")

	log.Printf("🔄 Conectando a broker MQTT: %s", broker)
	log.Printf("📩 Topic a suscribir: %s", topic)

	opts := MQTT.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID("COLAEVENTDRIVE_" + time.Now().Format("20060102150405"))
	opts.SetDefaultPublishHandler(mqttAdapter.MessageHandler)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(5 * time.Second)

	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		log.Printf("❌ Conexión perdida: %v", err)
	})

	opts.SetOnConnectHandler(func(client MQTT.Client) {
		log.Println("✅ Conectado exitosamente al broker MQTT")
		token := client.Subscribe(topic, 1, mqttAdapter.MessageHandler)
		token.Wait()
		if token.Error() != nil {
			log.Printf("❌ Error al suscribirse: %v", token.Error())
		} else {
			log.Printf("✅ Suscrito exitosamente al topic: %s", topic)
		}
	})

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Error al conectar: %v", token.Error())
	}
	defer client.Disconnect(250)

	if token := client.Subscribe(topic, 1, mqttAdapter.MessageHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("❌ Error al suscribirse: %v", token.Error())
	}

	http.HandleFunc("/ws", wsAdapter.HandleWebSocket)
	wsPort := os.Getenv("WEBSOCKET_PORT")
	if wsPort == "" {
		wsPort = "8080"
	}

	go func() {
		log.Printf("🌐 Servidor WebSocket iniciado en puerto %s", wsPort)
		if err := http.ListenAndServe(":"+wsPort, nil); err != nil {
			log.Fatalf("❌ Error en WebSocket: %v", err)
		}
	}()

	log.Println("✅ Servidor iniciado y esperando mensajes...")
	select {}
}
