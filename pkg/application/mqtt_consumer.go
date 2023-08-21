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
	"reflect"
	"strconv"
	"strings"
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

func SubscribeTopic(client mqtt.Client, qos byte, transmissionChannel chan entities.CapturedData, mqttConfiguration entities.MqttConfig, deviceConfiguration map[string]entities.Device, mqttConfigSensor entities.SensorDetail) {
	if token := client.Subscribe(mqttConfiguration.Topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		onMessageReceived(msg, transmissionChannel, deviceConfiguration, mqttConfigSensor)
	}); token.Wait() && token.Error() != nil {
		log.Println(token.Error())
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

func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttConfigSensor entities.SensorDetail) {
	var data map[string]interface{}

	err := json.Unmarshal([]byte(msg.Payload()), &data)

	if err != nil {
		log.Println("Erro ao converter JSON:", err)
		return
	}

	var finalData entities.CapturedData

	sensorId := getField(mqttConfigSensor.ID, data)
	value := getField(mqttConfigSensor.Value, data)
	timestamp := getField(mqttConfigSensor.Timestamp, data)

	if sensorId == nil || value == nil || timestamp == nil {
		return
	}

	sensorIdParse := sensorId.(float64)
	timestampParse := timestamp.(string)

	fmt.Println("sensorId: ", sensorId)
	fmt.Println("value: ", value)
	fmt.Println("timestamp: ", timestamp)

	if validateDevice(deviceConfiguration, sensorIdParse, value) {
		finalData.ID = int(math.Round(sensorIdParse))

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
		log.Printf("Erro: O dado do sensor %v está diferente do configurado no device_config", sensorId)
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
				fmt.Println("Erro:", err)
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

func validateDevice(deviceConfiguration map[string]entities.Device, sensorId float64, value interface{}) bool {
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
			if config.SensorID == int(sensorId) && hexMap[config.Schema.ValueType] == typeOf {
				isValid = true
			}
		}
	}
	return isValid
}
