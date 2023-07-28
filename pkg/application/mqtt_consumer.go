package application

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"knot-mqtt/internal/entities"
	"log"
	"os"
	"os/signal"
	"time"
)

const (
	mqttBroker     = "tcp://localhost:1883"
	mqttClientID   = "mqtt-subscriber"
	topic          = "knot"
	datetimeLayout = "2006-01-02 15:04:05"
)

func ConfigureClient() mqtt.Client {
	//Configure client
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID(mqttClientID)

	// Create MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Println("Conexão MQTT estabelecida")
	return client
}

func SubscribeTopic(client mqtt.Client, qos byte) {
	if token := client.Subscribe(topic, qos, onMessageReceived); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	log.Printf("Subscrição realizada no tópico: %s", topic)
}

func onMessageReceived(client mqtt.Client, msg mqtt.Message) {
	var capturedData entities.CapturedData

	err := json.Unmarshal([]byte(msg.Payload()), &capturedData)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return
	}

	// Imprimindo os dados decodificados
	fmt.Println("ID:", capturedData.ID)
	for _, data := range capturedData.Data {
		data.Timestamp = time.Now().Format(datetimeLayout)
		fmt.Println("SensorID:", data.SensorID)
		fmt.Println("Value:", data.Value)
		fmt.Println("Timestamp:", data.Timestamp)
	}

}

func EndListen() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Sinal de interrupção recebido. Desconectando...")
	fmt.Println("Desconectando...")
}
