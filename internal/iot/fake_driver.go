package iot

import (
	"fmt"
	"iot-bridge/internal/store"
)

type FakeDriver struct{}

func (f *FakeDriver) GetState(device store.Device) (map[string]string, error) {
	// Just return whatever is already stored — in real life, this would hit Zigbee, MQTT, etc.
	return device.State, nil
}

func (f *FakeDriver) SetState(device store.Device, updates map[string]string) error {
	// Simulate setting values without validating hardcoded capability names
	for k, v := range updates {
		fmt.Printf("🔧 Set %s = %s on device %s\n", k, v, device.ID)
	}
	return nil
}

var defaultDriver = &FakeDriver{}

/*
// GetDriverFor returns the default driver for now.
func GetDriverFor(device store.Device) DeviceDriver {
	return defaultDriver
}
*/
