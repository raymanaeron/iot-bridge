package store

func (d DiscoveredDevice) ToDevice(name, room string) Device {
	return Device{
		ID:       d.ID,
		Name:     name,
		Room:     room,
		Type:     d.Type,
		Protocol: d.Protocol,
		State:    map[string]string{},
	}
}
