package store

type Device struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Protocol     string            `json:"protocol"`
	Room         string            `json:"room"`
	State        map[string]string `json:"state"`
	Capabilities []Capability      `json:"capabilities"`
}

type DeviceStore interface {
	Add(device Device) error
	GetAll() []Device
	Get(id string) (Device, bool)
	UpdateState(id string, updates map[string]string) error
	Delete(id string) error
}
