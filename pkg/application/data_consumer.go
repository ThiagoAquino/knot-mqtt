package application

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/CESARBR/knot-mqtt/internal/entities"
	"github.com/CESARBR/knot-mqtt/internal/gateways/knot"
	"github.com/sirupsen/logrus"
)

var consumerMutex *sync.Mutex = knot.GetMutex()

func DataConsumer(transmissionChannel chan entities.CapturedData, logger *logrus.Entry, knotIntegration *knot.Integration, pipeDevices chan map[string]entities.Device) {
	/*
		Receives the data collected from the database.
	*/
	devices := <-pipeDevices
	device := getDevice(devices)
	knotIntegration.Register(device)
	sensorIDTimestampMapping := make(map[int]string)

	for capturedData := range transmissionChannel {
		for _, row := range capturedData.Rows {
			if isMeasurementNew(sensorIDTimestampMapping, row, capturedData.ID) {
				sensorIDTimestampMapping = updateTagNameTimestampMapping(sensorIDTimestampMapping, row, capturedData.ID)
				go sentDataToKNoT(row, capturedData, device, knotIntegration)
				//Reset the sensors array to avoid data duplication.
				device.Data = nil
			}
		}
		select {
		case devices = <-pipeDevices:
			consumerMutex.Lock()
			device = getDevice(devices)
			consumerMutex.Unlock()
		default:
		}
	}
}

func getDevice(devices map[string]entities.Device) entities.Device {
	/*
		Returns the first and only device in the mapping.
	*/
	keys := make([]string, 0)
	for key := range devices {
		keys = append(keys, key)
	}
	const firstDeviceIndex = 0
	return devices[keys[firstDeviceIndex]]
}

func sentDataToKNoT(row entities.Row, capturedData entities.CapturedData, device entities.Device, knotIntegration *knot.Integration) {
	/*
		Structure the data collected from the database in the format expected by KNoT,
		and finally transmits the data to the KNoT Cloud
	*/
	data := entities.Data{}
	data = entities.Data{SensorID: capturedData.ID, Value: row.Value, TimeStamp: row.Timestamp}
	device.Data = append(device.Data, data)
	knotIntegration.Transmit(device)
}

func convertStringValueToNumeric(value string) (interface{}, error) {
	if integerValue, err := strconv.Atoi(value); err == nil {
		return integerValue, err
	}
	if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
		return floatValue, err
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue, err
	}
	if isEmptyString(value) {
		return 0, fmt.Errorf("type conversion error")
	}
	return value, nil
}

func isEmptyString(value string) bool {
	return value == ""
}

func isMeasurementNew(tagNameTimestampMapping map[int]string, row entities.Row, sensorID int) bool {
	// Checks if the timestamp of the current measurement is different from the previous one.
	// As the database returns the query result temporally ordered,
	// we just need to check if the current timestamp is different from the previous one.
	return tagNameTimestampMapping[sensorID] != row.Timestamp
}

func updateTagNameTimestampMapping(tagNameTimestampMapping map[int]string, row entities.Row, sensorID int) map[int]string {
	tagNameTimestampMapping[sensorID] = row.Timestamp
	return tagNameTimestampMapping
}
