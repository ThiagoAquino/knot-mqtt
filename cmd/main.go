package main

import (
	"fmt"
	"github.com/CESARBR/knot-mqtt/internal/entities"
	"github.com/CESARBR/knot-mqtt/internal/gateways/knot"
	"github.com/CESARBR/knot-mqtt/internal/utils"
	"github.com/CESARBR/knot-mqtt/pkg/application"
	"github.com/CESARBR/knot-mqtt/pkg/logging"
	_ "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
)

func main() {
	startPprof()
	applicationConfiguration, deviceConfiguration, knotConfiguration, mqttConfiguration := loadConfiguration()

	log := setupLogger(applicationConfiguration.LogFilepath)
	logger := log.Get("Main")

	transmissionChannel := make(chan entities.CapturedData, len(applicationConfiguration.PertinentTags))

	//Create and Configure client
	client := application.ConfigureClient(mqttConfiguration)
	defer client.Disconnect(250)

	// Configura o QoS para 2 (entrega exatamente uma vez)
	qos := byte(2)

	// Inscreve-se no tópico "/topico/subtopico" com o QoS configurado
	application.SubscribeTopic(client, qos, transmissionChannel, mqttConfiguration)
	fmt.Println()

	pipeDevices := make(chan map[string]entities.Device)
	knotIntegration, err := knot.NewKNoTIntegration(pipeDevices, knotConfiguration, logger, deviceConfiguration)
	application.VerifyError(err)
	go application.DataConsumer(transmissionChannel, log.Get("Data consumer"), knotIntegration, pipeDevices)
	application.WaitUntilShutdown()
}

func loadConfiguration() (entities.Application, map[string]entities.Device, entities.IntegrationKNoTConfig, entities.MqttConfig) {
	applicationConfiguration, err := utils.ConfigurationParser("internal/configuration/application_configuration.yaml", entities.Application{})
	application.VerifyError(err)
	deviceConfiguration, err := utils.ConfigurationParser("internal/configuration/device_config.yaml", make(map[string]entities.Device))
	application.VerifyError(err)
	knotConfiguration, err := utils.ConfigurationParser("internal/configuration/knot_setup.yaml", entities.IntegrationKNoTConfig{})
	application.VerifyError(err)
	mqttConfiguration, err := utils.ConfigurationParser("internal/configuration/mqtt_setup.yaml", entities.MqttConfig{})
	application.VerifyError(err)
	return applicationConfiguration, deviceConfiguration, knotConfiguration, mqttConfiguration
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
