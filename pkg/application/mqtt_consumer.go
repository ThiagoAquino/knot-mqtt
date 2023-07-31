package application

import (
	"encoding/json"
	"fmt"
	"github.com/CESARBR/knot-mqtt/internal/entities"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"os/signal"
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

func SubscribeTopic(client mqtt.Client, qos byte, transmissionChannel chan entities.CapturedData) {
	if token := client.Subscribe(topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		onMessageReceived(msg, transmissionChannel)
	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	log.Printf("Subscrição realizada no tópico: %s", topic)
}

func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData) {
	var capturedData entities.CapturedData

	err := json.Unmarshal([]byte(msg.Payload()), &capturedData)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return
	}

	// Imprimindo os dados decodificados
	fmt.Println("SensorId:", capturedData.ID)
	for _, row := range capturedData.Rows {
		fmt.Println("Value:", row.Value)
		fmt.Println("Timestamp:", row.Timestamp)
	}
	transmissionChannel <- capturedData
}

func WaitUntilShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Sinal de interrupção recebido. Desconectando...")
	fmt.Println("Desconectando...")
}

func VerifyError(err error) {
	if err != nil {
		panic(err)
	}
}
