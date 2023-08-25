package application

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func ConfigureClient(mqttConfiguration entities.MqttConfig) mqtt.Client {
	//Configure client
	opts := mqtt.NewClientOptions().AddBroker(mqttConfiguration.MqttBroker)
	opts.SetClientID(mqttConfiguration.MqttClientID)
	opts.SetUsername(mqttConfiguration.MqttUser)
	opts.SetPassword(mqttConfiguration.MqttPass)

	// Create MQTT client
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	log.Println("Established MQTT connection")
	return client
}

func SubscribeTopic(client mqtt.Client, transmissionChannel chan entities.CapturedData, transmissionChannelTopic chan mqtt.Message, mqttConfiguration entities.MqttConfig, deviceConfiguration map[string]entities.Device, mqttConfigSensor []entities.SensorDetail) {
	if token := client.Subscribe(mqttConfiguration.Topic, mqttConfiguration.MqttQoS, func(client mqtt.Client, msg mqtt.Message) {
		transmissionChannelTopic <- msg
	}); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
		os.Exit(1)
	}

	log.Printf("Subscription made to the topic: %s", mqttConfiguration.Topic)
	go processMessages(transmissionChannelTopic, transmissionChannel, deviceConfiguration, mqttConfigSensor)
}

func processMessages(transmissionChannelTopic chan mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttConfigSensor []entities.SensorDetail) {
	for msg := range transmissionChannelTopic {
		for idSensor, config := range mqttConfigSensor {
			if msg.Topic() == config.Topic {
				onMessageReceived(msg, transmissionChannel, deviceConfiguration, config, idSensor+1)
			}
		}
	}
}

func WaitUntilShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	log.Println("Interrupt signal received. Disconnecting...")
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
		log.Println("Error to parse JSON:", err)
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

		fmt.Println("SensorId:", finalData.ID)
		// Imprimir os dados decodificados
		for _, row := range finalData.Rows {
			fmt.Println("Value:", row.Value)
			fmt.Println("Timestamp:", row.Timestamp)
		}

		transmissionChannel <- finalData
	} else {
		log.Printf("Error: Sensor data %v  is different from that configured in device_config", idSensor)
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
				log.Println("Error:", err)
				return nil
			}
			if index < 0 || index >= len(value) {
				log.Printf("Error: Index %v  not compatible with configuration in mqtt_device_config", index)
				return nil
			}
			field = value[index]
		default:
			log.Println("Error: unknown type ", reflect.TypeOf(field))
			return nil
		}
	}

	return field
}

var deviceConfigLock sync.RWMutex

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

	deviceConfigLock.RLock()
	defer deviceConfigLock.RUnlock()

	typeOf := reflect.TypeOf(value).Name()

	for _, device := range deviceConfiguration {
		for _, config := range device.Config {
			if config.SensorID == sensorId && hexMap[config.Schema.ValueType] == typeOf {
				return true
			}
		}
	}
	return false
}
