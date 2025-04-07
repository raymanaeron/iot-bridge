package iot

import (
	"iot-bridge/internal/iot/zigbee"
	"iot-bridge/internal/store"
)

type DeviceDriver interface {
	GetState(device store.Device) (map[string]string, error)
	SetState(device store.Device, updates map[string]string) error
}

func GetDriverFor(device store.Device) DeviceDriver {
	switch device.Protocol {
	case "zigbee":
		return zigbee.GetDriver()
		// Add other protocols here (zwave, matter, etc.) as needed
	}
	return nil // or panic/log
}
