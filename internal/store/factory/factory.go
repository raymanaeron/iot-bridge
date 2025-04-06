package factory

import (
	"iot-bridge/internal/config"
	"iot-bridge/internal/store"
	"iot-bridge/internal/store/inmemory"
	"iot-bridge/internal/store/sqlite"
)

var activeStore store.DeviceStore
var scanStore store.ScanStore

func Init() {
	if config.DemoMode {
		activeStore = inmemory.New()
		scanStore = inmemory.NewScanStore()
	} else {
		sqlStore := sqlite.New()
		activeStore = sqlStore
		scanStore = sqlite.NewScanStore(sqlStore.(*sqlite.SQLiteStore).DB())
	}
}

func GetDeviceStore() store.DeviceStore {
	return activeStore
}

func GetScanStore() store.ScanStore {
	return scanStore
}
