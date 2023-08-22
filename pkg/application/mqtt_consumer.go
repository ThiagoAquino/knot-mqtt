package application

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	mqtt "github.com/eclipse/paho.mqtt.golang"
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

func SubscribeTopic(client mqtt.Client, transmissionChannel chan entities.CapturedData, transmissionChannelTopic chan mqtt.Message, mqttConfiguration entities.MqttConfig, deviceConfiguration map[string]entities.Device, mqttConfigSensor []entities.SensorDetail) {
	if token := client.Subscribe(mqttConfiguration.Topic, mqttConfiguration.MqttQoS, func(client mqtt.Client, msg mqtt.Message) {
		transmissionChannelTopic <- msg
	}); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("Subscrição realizada no tópico: %s", mqttConfiguration.Topic)
	go processMessages(transmissionChannelTopic, transmissionChannel, deviceConfiguration, mqttConfigSensor)
}

func processMessages(transmissionChannelTopic chan mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttConfigSensor []entities.SensorDetail) {
	for msg := range transmissionChannelTopic {
		for i, config := range mqttConfigSensor {
			if msg.Topic() == config.Topic {
				onMessageReceived(msg, transmissionChannel, deviceConfiguration, config, i+1)

			}
		}
	}
}

func WaitUntilShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Sinal de interrupção recebido. Desconectando...")
	log.Println("Desconectando...")
}

func VerifyError(err error) {
	if err != nil {
		panic(err)
	}
}

func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttConfigSensor entities.SensorDetail, idSensor int) {
	var data map[string]interface{}

	err := json.Unmarshal([]byte(msg.Payload()), &data)

	if err != nil {
		log.Println("Erro ao converter JSON:", err)
		return
	}

	var finalData entities.CapturedData

	value := getField(mqttConfigSensor.Value, data)
	timestamp := getField(mqttConfigSensor.Timestamp, data)

	if value == nil || timestamp == nil {
		return
	}

	timestampParse := timestamp.(string)

	if validateDevice(deviceConfiguration, idSensor, value) {
		finalData.ID = idSensor

		var dataRow entities.Row
		dataRow.Value = value
		dataRow.Timestamp = timestampParse
		finalData.Rows = append(finalData.Rows, dataRow)

		transmissionChannel <- finalData
	} else {
		log.Printf("Erro: O dado do sensor %v está diferente do configurado no device_config", idSensor)
	}

}

func getField(campo string, data map[string]interface{}) interface{} {
	parts := strings.Split(campo, ".")
	field := data[parts[0]]

	for _, part := range parts[1:] {
		switch value := field.(type) {
		case map[string]interface{}:
			field = value[part]
		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil {
				log.Println("Erro:", err)
				return nil
			}
			if index < 0 || index >= len(value) {
				log.Printf("Erro: Índice %v incompativel com a configuração no mqtt_device_config", index)
				return nil
			}
			field = value[index]
		default:
			log.Println("Erro: Tipo desconhecido ", reflect.TypeOf(field))
			return nil
		}
	}

	return field
}

func validateDevice(deviceConfiguration map[string]entities.Device, sensorId int, value interface{}) bool {
	hexMap := map[int]string{
		1: "int",
		2: "float64",
		3: "bool",
		4: "string",
		5: "int64",
		6: "uint",
		7: "double",
	}

	isValid := false
	for _, device := range deviceConfiguration {
		for _, config := range device.Config {
			typeOf := reflect.TypeOf(value).Name()
			if config.SensorID == sensorId && hexMap[config.Schema.ValueType] == typeOf {
				isValid = true
			}
		}
	}
	return isValid
}
