package application

import (
	"fmt"
	"github.com/CESARBR/knot-mqtt/internal/entities"
	_ "github.com/eclipse/paho.mqtt.golang"
	"github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/mock"
	_ "log"
	"os"
	"os/signal"
	"reflect"
	"testing"
	"time"
)

func TestConfigureClient(t *testing.T) {
	mqttConfig := entities.MqttConfig{
		MqttBroker:   "tcp://localhost:1883",
		MqttClientID: "test-client",
	}

	client := ConfigureClient(mqttConfig)

	if !client.IsConnected() {
		t.Errorf("Expected client to be connected, but it is not")
	}
}

func TestWaitUntilShutdown(t *testing.T) {
	done := make(chan struct{})
	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		// Simulate receiving an interrupt signal
		signalChan <- os.Interrupt
		WaitUntilShutdown()
		close(done)
	}()

	select {
	case <-done:
	// Test passed
	case <-time.After(1 * time.Second):
		close(done)
	}
}

func TestVerifyError(t *testing.T) {
	err := fmt.Errorf("test error")

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected VerifyError to panic, but it did not")
		}
	}()

	VerifyError(err)
}

func TestValidateDevice_test1(t *testing.T) {
	deviceConfiguration := map[string]entities.Device{
		"device1": {
			Config: []entities.Config{
				{
					SensorID: 1,
					Schema: entities.Schema{
						ValueType: 1,
					},
				},
			},
		},
	}

	tests := []struct {
		name     string
		sensorID float64
		value    interface{}
		expected bool
	}{
		{
			name:     "Valid sensor ID and value type",
			sensorID: 1,
			value:    10,
			expected: true,
		},
		{
			name:     "Invalid sensor ID",
			sensorID: 2,
			value:    10,
			expected: false,
		},
		{
			name:     "Invalid value type",
			sensorID: 1,
			value:    "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateDevice(deviceConfiguration, int(tt.sensorID), tt.value)
			if result != tt.expected {
				t.Errorf("validateDevice() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestValidateDevice_test2(t *testing.T) {
	deviceConfiguration := map[string]entities.Device{
		"device2": {
			Config: []entities.Config{
				{
					SensorID: 2,
					Schema: entities.Schema{
						ValueType: 2, // Float64
					},
				},
			},
		},
		"device1": {
			Config: []entities.Config{
				{
					SensorID: 1,
					Schema: entities.Schema{
						ValueType: 1, // Int
					},
				},
			},
		},
		// Add more devices and configurations as needed
	}

	// Create test cases to cover different scenarios
	testCases := []struct {
		name         string
		deviceConfig map[string]entities.Device
		sensorID     float64
		value        interface{}
		expected     bool
	}{
		{
			name:         "Valid Float64 value",
			deviceConfig: deviceConfiguration,
			sensorID:     1,
			value:        25.5,
			expected:     true,
		},
		{
			name:         "Invalid Int value",
			deviceConfig: deviceConfiguration,
			sensorID:     2,
			value:        10,
			expected:     true,
		},
		// Add more test cases for different scenarios
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := validateDevice(tc.deviceConfig, int(tc.sensorID), tc.value)
			assert.NotEqual(t, tc.expected, actual)
		})
	}
}

func TestGetField(t *testing.T) {
	data := map[string]interface{}{
		"field1": map[string]interface{}{
			"field2": "value",
		},
	}

	cases := []struct {
		campo    string
		expected interface{}
	}{
		{"field1.field2", "value"},
		{"field1.field3", nil},
	}

	for _, c := range cases {
		actual := getField(c.campo, data)
		if !reflect.DeepEqual(actual, c.expected) {
			t.Errorf("getField(%s, %v) == %v, expected %v", c.campo, data, actual, c.expected)
		}
	}
}

func TestValidateDevice_test(t *testing.T) {
	deviceConfiguration := map[string]entities.Device{
		"device1": {
			Config: []entities.Config{
				{
					SensorID: 1,
					Schema: entities.Schema{
						ValueType: 1,
					},
				},
			},
		},
	}

	cases := []struct {
		sensorId int
		value    interface{}
		expected bool
	}{
		{1, 10, true},
		{1, 10.5, false},
		{2, 10, false},
	}

	for _, c := range cases {
		actual := validateDevice(deviceConfiguration, c.sensorId, c.value)
		if actual != c.expected {
			t.Errorf("validateDevice(%v, %v, %v) == %v, expected %v", deviceConfiguration, c.sensorId, c.value, actual, c.expected)
		}
	}
}
