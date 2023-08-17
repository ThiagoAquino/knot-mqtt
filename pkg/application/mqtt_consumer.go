package application

import (
	"encoding/json"
	"fmt"
	"github.com/CESARBR/knot-mqtt/internal/entities"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"math"
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

func SubscribeTopic(client mqtt.Client, qos byte, transmissionChannel chan entities.CapturedData, mqttConfiguration entities.MqttConfig, deviceConfiguration map[string]entities.Device, mqttConfigSensor entities.ConfigSensor) {
	if token := client.Subscribe(mqttConfiguration.Topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		onMessageReceived(msg, transmissionChannel, deviceConfiguration, mqttConfigSensor)
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

func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttConfigSensor entities.ConfigSensor) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(msg.Payload()), &data)
	if err != nil {
		log.Println("Erro ao converter JSON:", err)
		return
	}

	var finalData entities.CapturedData

	capturedDataArray, _ := data["data"].([]interface{})
	for _, capturedData := range capturedDataArray {
		capturedDataMap, _ := capturedData.(map[string]interface{})
		sensorId, _ := capturedDataMap[mqttConfigSensor.Sensor.ID].(float64)
		value, _ := capturedDataMap[mqttConfigSensor.Sensor.Value]
		times, _ := capturedDataMap[mqttConfigSensor.Sensor.Timestamp].(string)

		finalData.ID = int(math.Round(sensorId))

		var dataRow entities.Row
		dataRow.Value = value
		dataRow.Timestamp = times
		finalData.Rows = append(finalData.Rows, dataRow)

		fmt.Println("SensorId:", finalData.ID)
		// Imprimir os dados decodificados
		for _, row := range finalData.Rows {
			fmt.Println("Value:", row.Value)
			fmt.Println("Timestamp:", row.Timestamp)
		}
	}
	transmissionChannel <- finalData
}
