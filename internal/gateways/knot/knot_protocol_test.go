package knot

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/CESARBR/knot-thing-sql/internal/entities"
	"github.com/CESARBR/knot-thing-sql/internal/gateways/knot/network"
	"github.com/CESARBR/knot-thing-sql/internal/utils"
	"github.com/CESARBR/knot-thing-sql/pkg/logging"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func loadConfiguration() (map[string]entities.Device, entities.IntegrationKNoTConfig) {
	deviceConfiguration, err := utils.ConfigurationParser("../../configuration/device_config.yaml", make(map[string]entities.Device))
	if err != nil {
		panic(err)
	}

	knotConfiguration, err := utils.ConfigurationParser("../../configuration/knot_setup.yaml", entities.IntegrationKNoTConfig{})
	if err != nil {
		panic(err)
	}

	return deviceConfiguration, knotConfiguration
}

func setUp() entities.Device {
	knot_device := entities.Device{}
	knot_device.ID = "1"
	return knot_device
}

func createData(sensorID int) entities.Data {
	return entities.Data{
		SensorID:  sensorID,
		Value:     0,
		TimeStamp: "2022-01-02 07:45:04",
	}

}

func createDataWithEmptyTimestamp(sensorID int) entities.Data {
	return entities.Data{
		SensorID:  sensorID,
		Value:     0,
		TimeStamp: nil,
	}

}

func createDataWithInvalidValue(sensorID int) entities.Data {
	return entities.Data{
		SensorID:  sensorID,
		Value:     nil,
		TimeStamp: nil,
	}

}

func createConfig(sensorID int) entities.Config {
	return entities.Config{
		SensorID: sensorID,
		Schema:   entities.Schema{ValueType: 1, Unit: 0, TypeID: 65296},
		Event:    entities.Event{},
	}

}

func TestGivenEmptyStateThenCreateDevice(t *testing.T) {

	fakeDevice := setUp()
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.createDevice(fakeDevice)

	assert.NoError(t, err)
}

func TestGivenEmptyStateThenCreateDeviceWithKnotNewState(t *testing.T) {

	fakeDevice := setUp()
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.createDevice(fakeDevice)

	assert.NoError(t, err)
	assert.Equal(t, newProtocol.devices[fakeDevice.ID].State, entities.KnotNew)
}

func TestGivenNonEmptyStateThenFailToCreateDevice(t *testing.T) {

	fakeDevice := setUp()
	fakeDevice.State = "invalid"
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.createDevice(fakeDevice)
	assert.Error(t, err)
}

func TestGivenExistingDeviceThenTrue(t *testing.T) {
	fakeDevice := setUp()
	fakeDevice.State = ""
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	newProtocol.devices[fakeDevice.ID] = fakeDevice
	foundDevice := newProtocol.deviceExists(newProtocol.devices[fakeDevice.ID])
	assert.Equal(t, foundDevice, true)
}

func TestGivenNotExistingDeviceThenFalse(t *testing.T) {
	fakeDevice := setUp()
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	foundDevice := newProtocol.deviceExists(newProtocol.devices[fakeDevice.ID])
	assert.Equal(t, foundDevice, false)
}

func TestGivenValidDataThenSucess(t *testing.T) {
	fakeDevice := setUp()
	var fakeData []entities.Data
	for i := 1; i <= 10; i++ {
		fakeData = append(fakeData, createData(i))
	}

	fakeDevice.Data = fakeData

	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}

	err := newProtocol.checkData(fakeDevice)

	assert.NoError(t, err)
}

func TestGivenInvalidDataThenError(t *testing.T) {
	/*
		invalid starts with sensorID as 0. Why? I do not know.
	*/
	fakeDevice := setUp()
	var fakeData []entities.Data
	for i := 0; i < 2; i++ {
		fakeData = append(fakeData, createData(1))
	}
	fakeDevice.Data = fakeData
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.checkData(fakeDevice)
	assert.Error(t, err)
}

func TestGivenInvalidTimestampDataThenError(t *testing.T) {
	fakeDevice := setUp()
	var fakeData []entities.Data
	for i := 1; i <= 10; i++ {
		fakeData = append(fakeData, createDataWithEmptyTimestamp(i))
	}
	fakeDevice.Data = fakeData
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.checkData(fakeDevice)
	assert.Error(t, err)
}

func TestGivenInvalidValueDataThenError(t *testing.T) {
	fakeDevice := setUp()
	var fakeData []entities.Data
	for i := 1; i <= 10; i++ {
		fakeData = append(fakeData, createDataWithInvalidValue(i))
	}
	fakeDevice.Data = fakeData
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.checkData(fakeDevice)
	assert.Error(t, err)
}

func TestGivenValidConfigThenSucess(t *testing.T) {
	fakeDevice := setUp()
	var fakeConfig []entities.Config
	for i := 1; i <= 10; i++ {
		fakeConfig = append(fakeConfig, createConfig(i))
	}

	fakeDevice.Config = fakeConfig

	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.checkDeviceConfiguration(fakeDevice)
	assert.NoError(t, err)
}

func TestGivenInvalidConfigThenError(t *testing.T) {
	fakeDevice := setUp()
	var fakeConfig []entities.Config
	repeatedIdentifier := 0
	for i := 0; i < 10; i++ {
		fakeConfig = append(fakeConfig, createConfig(repeatedIdentifier))
	}
	fakeDevice.Config = fakeConfig
	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}
	err := newProtocol.checkDeviceConfiguration(fakeDevice)

	assert.Error(t, err)
}

func TestGenerateID(t *testing.T) {

	_, err := tokenIDGenerator()

	assert.NoError(t, err)
}

func TestGivenValidDeviceIDThenDeleteIt(t *testing.T) {
	fakeDevice := setUp()

	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}

	newProtocol.devices[fakeDevice.ID] = fakeDevice

	err := newProtocol.deleteDevice(fakeDevice.ID)

	assert.NoError(t, err)
}

func TestGivenInvalidDeviceIDThenError(t *testing.T) {
	fakeDevice := setUp()

	newProtocol := protocol{
		devices: make(map[string]entities.Device),
	}

	err := newProtocol.deleteDevice(fakeDevice.ID)

	assert.Error(t, err)
}

func sendDeviceToChannel(device entities.Device, channel chan entities.Device) {
	channel <- device
}

func createChannels() (chan map[string]entities.Device, chan entities.Device) {
	pipeDevices := make(chan map[string]entities.Device)
	deviceChannel := make(chan entities.Device)

	return pipeDevices, deviceChannel
}

func createLogger() *logrus.Entry {
	logrus := logging.NewLogrus("info", os.Stdout)
	logger := logrus.Get("Main")
	return logger
}

func createNullLogger() (*logrus.Entry, *test.Hook) {
	log, hook := test.NewNullLogger()
	logger := log.WithFields(logrus.Fields{
		"Context": "testing",
	})
	return logger, hook
}

func TestGivenInvalidConfigTimeout(t *testing.T) {

	conf := entities.IntegrationKNoTConfig{
		UserToken:               "",
		URL:                     "",
		EventRoutingKeyTemplate: "",
	}
	devices := make(map[string]entities.Device)
	pipeDevices, deviceChan := createChannels()
	logger := createLogger()
	msgChan := make(chan network.InMsg)
	var err error
	go func() {
		_, err = newProtocol(pipeDevices, conf, deviceChan, msgChan, logger, devices)
	}()

	timeout := time.Second * 60
	fmt.Println("Please, wait 60 seconds!")
	time.Sleep(timeout)

	if err == nil {
		err = fmt.Errorf("Timeout")
	}
	assert.Error(t, err)
}

func TestGivenDeviceWithoutTimeoutCheckTimeout(t *testing.T) {

	deviceConfiguration, knotConfiguration := loadConfiguration()
	pipeDevices, deviceChan := createChannels()
	logger := createLogger()
	msgChan := make(chan network.InMsg)
	knotProtocol, _ := newProtocol(pipeDevices, knotConfiguration, deviceChan, msgChan, logger, deviceConfiguration)
	device := deviceConfiguration["2801d924dcf1fc52"]
	device.Error = ""
	deviceWithCheckedTimeout := knotProtocol.checkTimeout(device, logger)
	assert.Equal(t, deviceWithCheckedTimeout.Error, device.Error)
}

func TestGivenDeviceTimeoutCheckTimeoutStateNewWaitReg(t *testing.T) {

	deviceConfiguration, knotConfiguration := loadConfiguration()
	oldDevice := deviceConfiguration["2801d924dcf1fc52"]
	oldDevice.State = entities.KnotWaitReg
	deviceConfiguration[oldDevice.ID] = oldDevice
	pipeDevices, deviceChan := createChannels()
	logger, hook := createNullLogger()
	msgChan := make(chan network.InMsg)
	knotProtocol, _ := newProtocol(pipeDevices, knotConfiguration, deviceChan, msgChan, logger, deviceConfiguration)
	device := deviceConfiguration["2801d924dcf1fc52"]
	device.Error = "timeOut"
	device.State = entities.KnotNew
	_ = knotProtocol.checkTimeout(device, logger)
	expectedErrorMessage := "error: timeOut"
	assert.Equal(t, hook.LastEntry().Message, expectedErrorMessage)
}

func TestGivenDeviceTimeoutCheckTimeoutStateRegisteredWaitAuth(t *testing.T) {

	deviceConfiguration, knotConfiguration := loadConfiguration()
	oldDevice := deviceConfiguration["2801d924dcf1fc52"]
	oldDevice.State = entities.KnotWaitAuth
	deviceConfiguration[oldDevice.ID] = oldDevice
	pipeDevices, deviceChan := createChannels()
	logger, hook := createNullLogger()
	msgChan := make(chan network.InMsg)
	knotProtocol, _ := newProtocol(pipeDevices, knotConfiguration, deviceChan, msgChan, logger, deviceConfiguration)
	device := deviceConfiguration["2801d924dcf1fc52"]
	device.Error = "timeOut"
	device.State = entities.KnotRegistered

	_ = knotProtocol.checkTimeout(device, logger)
	expectedErrorMessage := "error: timeOut"
	assert.Equal(t, hook.LastEntry().Message, expectedErrorMessage)
}

func TestGivenDeviceTimeoutCheckTimeoutStateAuthWaitConfig(t *testing.T) {

	deviceConfiguration, knotConfiguration := loadConfiguration()
	oldDevice := deviceConfiguration["2801d924dcf1fc52"]
	oldDevice.State = entities.KnotWaitConfig
	deviceConfiguration[oldDevice.ID] = oldDevice
	pipeDevices, deviceChan := createChannels()
	logger, hook := createNullLogger()
	msgChan := make(chan network.InMsg)
	knotProtocol, _ := newProtocol(pipeDevices, knotConfiguration, deviceChan, msgChan, logger, deviceConfiguration)
	device := deviceConfiguration["2801d924dcf1fc52"]
	device.Error = "timeOut"
	device.State = entities.KnotAuth

	_ = knotProtocol.checkTimeout(device, logger)
	expectedErrorMessage := "error: timeOut"
	assert.Equal(t, hook.LastEntry().Message, expectedErrorMessage)
}

func TestGivenDeviceTimeoutCheckTimeoutStateNew(t *testing.T) {

	deviceConfiguration, knotConfiguration := loadConfiguration()
	pipeDevices, deviceChan := createChannels()
	logger := createLogger()
	msgChan := make(chan network.InMsg)
	knotProtocol, _ := newProtocol(pipeDevices, knotConfiguration, deviceChan, msgChan, logger, deviceConfiguration)
	device := deviceConfiguration["2801d924dcf1fc52"]
	device.Error = "timeOut"

	deviceWithCheckedTimeout := knotProtocol.checkTimeout(device, logger)
	assert.Equal(t, deviceWithCheckedTimeout.State, entities.KnotOff)
}

type testCase struct {
	bidingKey       string
	expectedMessage string
	testName        string
	errorMessage    string
}

var testCases = []testCase{
	{bidingKey: network.BindingKeyRegistered, expectedMessage: "received a registration response with no error", testName: "TestGivenBindingKeyRegisteredDeviceWithoutErrorHandleKnotAMQP"},
	{bidingKey: network.BindingKeyRegistered, expectedMessage: "received a registration response with a error", testName: "TestGivenBindingKeyRegisteredDeviceWithErrorHandleKnotAMQP", errorMessage: "Testing error"},
	{bidingKey: network.BindingKeyUpdatedConfig, expectedMessage: "received a config update response with a error", testName: "TestGivenBindingKeyUpdatedConfigWithoutErrorKnotAMQP", errorMessage: "testing error"},
}

func TestHandleKnotAMQP(t *testing.T) {

	for _, test := range testCases {
		logger, hook := createNullLogger()
		deviceChan := make(chan entities.Device)
		msgChan := make(chan network.InMsg)
		testAMQPMessage := network.InMsg{
			RoutingKey: test.bidingKey,
		}
		testDevice := entities.Device{ID: "42", Error: test.errorMessage}
		testAMQPMessage.Body, _ = json.Marshal(testDevice)
		newBidingKeyActionMapping := NewBidingKeyActionMapping()
		go handlerKnotAMQP(msgChan, deviceChan, logger, newBidingKeyActionMapping)
		msgChan <- testAMQPMessage
		<-deviceChan
		assert.Equal(t, hook.LastEntry().Message, test.expectedMessage, test.testName)
	}

}

type testCaseWithState struct {
	bidingKey       string
	expectedMessage string
	testName        string
	errorMessage    string
	state           string
}

var testCasesWithState = []testCaseWithState{
	{bidingKey: network.BindingKeyRegistered, expectedMessage: "received a registration response with a error", testName: "TestGivenBindingKeyRegisteredDeviceAlreadyRegisteredHandleKnotAMQP", errorMessage: "thing is already registered", state: entities.KnotAlreadyReg},
	{bidingKey: network.BindingKeyUnregistered, expectedMessage: "received a unregistration response", testName: "TestGivenBindingKeyUnregisteredHandleKnotAMQP", errorMessage: "", state: entities.KnotForceDelete},
	{bidingKey: network.ReplyToAuthMessages, expectedMessage: "received a authentication response with no error", testName: "TestGivenReplyToAuthMessagesHandleWithoutErrorKnotAMQP", errorMessage: "", state: entities.KnotAuth},
	{bidingKey: network.ReplyToAuthMessages, expectedMessage: "received a authentication response with a error", testName: "TestGivenReplyToAuthMessagesHandleWithoutErrorKnotAMQP", errorMessage: "testing error", state: entities.KnotForceDelete},
	{bidingKey: network.BindingKeyUpdatedConfig, expectedMessage: "received a config update response with no error", testName: "TestGivenBindingKeyUpdatedConfigWithoutErrorKnotAMQP", errorMessage: "", state: entities.KnotReady},
}

func TestGivenBindingKeyRegisteredDeviceAlreadyRegisteredHandleKnotAMQP(t *testing.T) {

	for _, test := range testCasesWithState {
		logger, hook := createNullLogger()
		deviceChan := make(chan entities.Device)
		msgChan := make(chan network.InMsg)
		testAMQPMessage := network.InMsg{
			RoutingKey: test.bidingKey,
		}
		testDevice := entities.Device{ID: "42", Error: test.errorMessage}
		testAMQPMessage.Body, _ = json.Marshal(testDevice)
		newBidingKeyActionMapping := NewBidingKeyActionMapping()
		go handlerKnotAMQP(msgChan, deviceChan, logger, newBidingKeyActionMapping)
		msgChan <- testAMQPMessage
		device := <-deviceChan
		assert.Equal(t, device.Error, test.errorMessage)
		assert.Equal(t, hook.LastEntry().Message, test.expectedMessage)
		assert.Equal(t, device.State, test.state)
	}
}
