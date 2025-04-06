package store

import "log"

type DiscoveredDevice struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
	Signal   int    `json:"signal_strength"`
}

type ScanStore interface {
	StartScan(protocols []string)
	GetScanResults() []DiscoveredDevice
	FindDiscoveredDevice(id string) (DiscoveredDevice, bool)
}

var discoveredDevices []DiscoveredDevice

func StartScan(protocols []string) {
	log.Println("⚙️  StartScan called. Simulating discovery...")

	discoveredDevices = []DiscoveredDevice{
		{ID: "bulb1", Name: "Unregistered Bulb", Type: "bulb", Protocol: "zigbee", Signal: -42},
		{ID: "plug1", Name: "New Plug", Type: "smart_plug", Protocol: "zwave", Signal: -55},
	}
}

func GetScanResults() []DiscoveredDevice {
	if discoveredDevices == nil {
		return []DiscoveredDevice{}
	}
	return discoveredDevices
}

func FindDiscoveredDevice(id string) (DiscoveredDevice, bool) {
	for _, d := range discoveredDevices {
		if d.ID == id {
			return d, true
		}
	}
	return DiscoveredDevice{}, false
}
