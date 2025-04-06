package iot

import (
	"fmt"
	"iot-bridge/internal/store"
	"strings"
)

type MockDriver struct{}

func NewMockDriver() DeviceDriver {
	return &MockDriver{}
}

func (d *MockDriver) GetState(device store.Device) (map[string]string, error) {
	if device.State == nil {
		return map[string]string{}, nil
	}
	return device.State, nil
}

func (d *MockDriver) SetState(device store.Device, updates map[string]string) error {
	for k, v := range updates {
		switch strings.ToLower(k) {
		case "power":
			if v != "on" && v != "off" {
				return fmt.Errorf("invalid power value: %s", v)
			}
		case "brightness", "color":
			continue
		default:
			return fmt.Errorf("unsupported capability: %s", k)
		}
	}
	return nil
}
