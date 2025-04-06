package iot

import "iot-bridge/internal/store"

type DeviceDriver interface {
	GetState(device store.Device) (map[string]string, error)
	SetState(device store.Device, updates map[string]string) error
}
