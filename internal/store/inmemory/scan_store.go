package inmemory

import "iot-bridge/internal/store"

type InMemoryScanStore struct {
	discovered []store.DiscoveredDevice
}

func NewScanStore() store.ScanStore {
	return &InMemoryScanStore{}
}

func (s *InMemoryScanStore) StartScan(protocols []string) {
	s.discovered = []store.DiscoveredDevice{
		{ID: "bulb1", Name: "Unregistered Bulb", Type: "bulb", Protocol: "zigbee", Signal: -42},
		{ID: "plug1", Name: "New Plug", Type: "smart_plug", Protocol: "zwave", Signal: -55},
	}
}

func (s *InMemoryScanStore) GetScanResults() []store.DiscoveredDevice {
	return s.discovered
}

func (s *InMemoryScanStore) FindDiscoveredDevice(id string) (store.DiscoveredDevice, bool) {
	for _, d := range s.discovered {
		if d.ID == id {
			return d, true
		}
	}
	return store.DiscoveredDevice{}, false
}
