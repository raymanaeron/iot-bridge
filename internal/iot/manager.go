package iot

import (
	"iot-bridge/internal/config"
)

var activeDriver DeviceDriver

func Init() {
	if config.DemoMode {
		activeDriver = NewMockDriver()
	} else {
		// Default to mock — replace with real protocol later
		activeDriver = NewMockDriver()
	}
}

/*
func GetDriverFor(device store.Device) DeviceDriver {
	return activeDriver
}
*/
