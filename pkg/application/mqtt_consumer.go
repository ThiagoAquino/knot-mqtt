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
	"strconv"
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

func SubscribeTopic(client mqtt.Client, qos byte, transmissionChannel chan entities.CapturedData, mqttConfiguration entities.MqttConfig, deviceConfiguration map[string]entities.Device, mqttDeviceConfiguration entities.DeviceConfig) {
	if token := client.Subscribe(mqttConfiguration.Topic, qos, func(client mqtt.Client, msg mqtt.Message) {
		onMessageReceived(msg, transmissionChannel, deviceConfiguration, mqttDeviceConfiguration)
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

//	func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttDeviceConfiguration entities.DeviceConfig) {
//		var capturedData struct {
//			ID   int                      `json:"id"`
//			Data []map[string]interface{} `json:"data"`
//		}
//
//		hexMap := map[string]int{
//			"int":    1,
//			"float":  2,
//			"bool":   3,
//			"raw":    4,
//			"int64":  5,
//			"uint":   6,
//			"double": 7,
//		}
//
//		err := json.Unmarshal([]byte(msg.Payload()), &capturedData)
//		if err != nil {
//			fmt.Println("Erro ao converter JSON:", err)
//			return
//		}
//
//		// Encontrar o tipo de dispositivo com base no sensorId
//		var deviceType string
//		for _, config := range deviceConfiguration {
//			for _, device := range config.Config {
//				sensorType := device.Schema.ValueType
//				for _, dataMap := range capturedData.Data {
//					if device.SensorID == int(dataMap["sensorId"].(float64)) {
//						for key, value := range hexMap {
//							if sensorType == value {
//								deviceType = key
//							}
//						}
//					}
//				}
//			}
//		}
//
//		// Construir a estrutura final removendo os campos indesejados
//		var finalData entities.CapturedData
//		dataId := capturedData.ID
//		for _, dataMap := range capturedData.Data {
//			var dataRow entities.Row
//
//			sensorId, _ := dataMap["sensorId"].(float64)
//			finalData.ID = int(sensorId)
//
//			switch deviceType {
//
//			case "int":
//				value, _ := dataMap["value"].(float64)
//				dataRow.Value = int(value)
//			case "float":
//				value, _ := dataMap["value"].(float64)
//				dataRow.Value = value
//			case "bool":
//				value, _ := dataMap["value"].(bool)
//				dataRow.Value = value
//			case "raw":
//				value, _ := dataMap["value"].(string)
//				dataRow.Value = value
//			case "int64":
//				value, _ := dataMap["value"].(float64)
//				dataRow.Value = int64(value)
//			case "uint":
//				value, _ := dataMap["value"].(float64)
//				dataRow.Value = uint(value)
//			case "double":
//				value, _ := dataMap["value"].(float64)
//				dataRow.Value = value
//			default:
//				fmt.Println("Erro no id")
//			}
//
//			timestamp, _ := dataMap["timestamp"].(string)
//			dataRow.Timestamp = timestamp
//
//			// Outros campos podem ser tratados da mesma maneira
//
//			finalData.Rows = append(finalData.Rows, dataRow)
//		}
//
//		// Imprimir os dados decodificados
//		fmt.Println("CapturedData:", dataId)
//		for _, row := range finalData.Rows {
//			fmt.Println("SensorId:", finalData.ID)
//			fmt.Println("Value:", row.Value)
//			fmt.Println("Timestamp:", row.Timestamp)
//		}
//
//		transmissionChannel <- finalData
//	}
func onMessageReceived(msg mqtt.Message, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device, mqttDeviceConfiguration entities.DeviceConfig) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(msg.Payload()), &data)
	fmt.Println(data)
	if err != nil {
		fmt.Println("Erro ao converter JSON:", err)
		return
	}
	for _, config := range mqttDeviceConfiguration.Config {
		var finalData entities.CapturedData
		value, _ := data[config.Value].(string)
		name, _ := data[config.Name].(string)
		times, _ := data[config.Timestamp].(string)
		sensorType, _ := data[config.SensorType].(float64)
		fmt.Println("Value:", value)
		fmt.Println("Name:", name)
		fmt.Println("times:", times)
		fmt.Println("sensorType:", sensorType)
		var dataRow entities.Row
		dataRow.Value, _ = strconv.ParseFloat(value, 64)
		dataRow.Timestamp = times
		dataRow.ID = int(math.Round(sensorType))
		finalData.ID = int(math.Round(sensorType))
		finalData.Rows = append(finalData.Rows, dataRow)
		// Imprimir os dados decodificados
		for _, row := range finalData.Rows {
			fmt.Println("SensorId:", finalData.ID)
			fmt.Println("Value:", row.Value)
			fmt.Println("Timestamp:", row.Timestamp)
		}
		transmissionChannel <- finalData
	}

}

//type T struct {
//	Id   string `json:"id"`
//	Data []struct {
//		SensorId  int    `json:"sensorId"`
//		Value     string `json:"value"`
//		Timestamp string `json:"timestamp"`
//	} `json:"data"`
//}
//
//{"id":"0d7cd9d221395e1e","data":[{"sensorId":1,"value":"12","timestamp":"2023-08-17 08:06:04"}]}
