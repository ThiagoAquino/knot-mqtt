package main

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	"github.com/CESARBR/knot-mqtt/internal/gateways/knot"
	"github.com/CESARBR/knot-mqtt/internal/utils"
	"github.com/CESARBR/knot-mqtt/pkg/application"
	"github.com/CESARBR/knot-mqtt/pkg/logging"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	startPprof()
	deviceConfiguration, knotConfiguration, mqttConfiguration, mqttDeviceConfiguration := loadConfiguration()
	log := setupLogger(mqttConfiguration.LogFilepath)
	logger := log.Get("Main")
	transmissionChannel := make(chan entities.CapturedData, mqttConfiguration.AmountTags)

	client := application.ConfigureClient(mqttConfiguration)
	defer client.Disconnect(250)

	subscribeTopicBySensorDetails(mqttDeviceConfiguration.SensorDetails, mqttConfiguration, client, transmissionChannel, deviceConfiguration)

	pipeDevices := make(chan map[string]entities.Device)
	knotIntegration, err := knot.NewKNoTIntegration(pipeDevices, knotConfiguration, logger, deviceConfiguration)
	application.VerifyError(err)
	go application.DataConsumer(transmissionChannel, log.Get("Data consumer"), knotIntegration, pipeDevices)
	application.WaitUntilShutdown()
}

func subscribeTopicBySensorDetails(mqttDeviceConfigSensorDetails []entities.SensorDetail, mqttConfiguration entities.MqttConfig, client mqtt.Client, transmissionChannel chan entities.CapturedData, deviceConfiguration map[string]entities.Device) {
	for i := 0; i < len(mqttDeviceConfigSensorDetails); i++ {
		transmissionChannelTopic := make(chan mqtt.Message)
		mqttConfiguration.Topic = mqttDeviceConfigSensorDetails[i].Topic
		application.SubscribeTopic(client, transmissionChannel, transmissionChannelTopic, mqttConfiguration, deviceConfiguration, mqttDeviceConfigSensorDetails)
	}
}

func loadConfiguration() (map[string]entities.Device, entities.IntegrationKNoTConfig, entities.MqttConfig, entities.DeviceConfig) {
	deviceConfiguration, err := utils.ConfigurationParser("internal/configuration/device_config.yaml", make(map[string]entities.Device))
	application.VerifyError(err)
	knotConfiguration, err := utils.ConfigurationParser("internal/configuration/knot_setup.yaml", entities.IntegrationKNoTConfig{})
	application.VerifyError(err)
	mqttConfiguration, err := utils.ConfigurationParser("internal/configuration/mqtt_setup.yaml", entities.MqttConfig{})
	application.VerifyError(err)
	mqttDeviceConfiguration, err := utils.ConfigurationParser("internal/configuration/mqtt_device_config.yaml", entities.DeviceConfig{})
	application.VerifyError(err)
	return deviceConfiguration, knotConfiguration, mqttConfiguration, mqttDeviceConfiguration
}

func startPprof() {
	go func() {
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()
}

func setupLogger(logFilepath string) *logging.Logrus {
	var log *logging.Logrus
	file, err := os.OpenFile(filepath.Clean(logFilepath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err == nil {
		log = logging.NewLogrus("info", file)
	} else {
		log = logging.NewLogrus("info", os.Stdout)
	}
	return log
}
