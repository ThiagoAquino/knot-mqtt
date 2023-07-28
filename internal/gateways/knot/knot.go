package knot

import (
	"github.com/CESARBR/knot-mqtt/internal/entities"
	"github.com/CESARBR/knot-mqtt/internal/gateways/knot/network"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Integration struct {
	protocol Protocol
}

var deviceChan = make(chan entities.Device)
var msgChan = make(chan network.InMsg)

func NewKNoTIntegration(pipeDevices chan map[string]entities.Device, conf entities.IntegrationKNoTConfig, log *logrus.Entry, devices map[string]entities.Device) (*Integration, error) {
	var err error
	KNoTInteration := Integration{}

	KNoTInteration.protocol, err = newProtocol(pipeDevices, conf, deviceChan, msgChan, log, devices)
	if err != nil {
		return nil, errors.Wrap(err, "new knot protocol")
	}

	return &KNoTInteration, nil
}

// HandleUplinkEvent sends an UplinkEvent.
func (i *Integration) HandleDevice(device entities.Device) {
	device.State = ""
	deviceChan <- device
}

func (integration *Integration) Close() error {
	return integration.protocol.Close()
}

func (i Integration) Transmit(device entities.Device) {
	i.HandleDevice(device)
}

func (i Integration) Register(device entities.Device) {
	i.HandleDevice(device)
}
