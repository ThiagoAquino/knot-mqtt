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

func ConfigureClient(mqttConfiguration entities.MqttConfig) mqtt.Client {
	//Configure client
	opts := mqtt.NewClientOptions().AddBroker(mqttConfiguration.MqttBroker)
	opts.SetClientID(mqttConfiguration.MqttClientID)

	// Create MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Println("Conexão MQTT estabelecida")
	return client
}

func SubscribeTopic(client mqtt.Client, qos byte, transmissionChannel chan entities.CapturedData, mqttConfiguration entities.MqttConfig) {
	if token := client.Subscribe(mqttConfiguration.Topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		onMessageReceived(msg, transmissionChannel)
	}); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
		os.Exit(1)
	}
	log.Printf("Subscrição realizada no tópico: %s", mqttConfiguration.Topic)
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

/** Com Payload genérico, descartando os campos que não são utilizados. Usando map[string] interface.*/

func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData) {
	var capturedData struct {
		ID   int                      `json:"id"`
		Data []map[string]interface{} `json:"data"`
	}

	err := json.Unmarshal([]byte(msg.Payload()), &capturedData)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return
	}

	// Construir a estrutura final removendo os campos indesejados
	var finalData entities.CapturedData
	dataId := capturedData.ID
	for _, dataMap := range capturedData.Data {
		var dataRow entities.Row

		sensorId, _ := dataMap["sensorId"].(float64)
		finalData.ID = int(sensorId)

		value, _ := dataMap["value"].(float64)
		dataRow.Value = value

		timestamp, _ := dataMap["timestamp"].(string)
		dataRow.Timestamp = timestamp

		// Outros campos podem ser tratados da mesma maneira

		finalData.Rows = append(finalData.Rows, dataRow)
	}

	// Imprimir os dados decodificados
	fmt.Println("CapturedData:", dataId)
	for _, row := range finalData.Rows {
		fmt.Println("SensorId:", finalData.ID)
		fmt.Println("Value:", row.Value)
		fmt.Println("Timestamp:", row.Timestamp)
	}

	transmissionChannel <- finalData
}
