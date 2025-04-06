package inmemory

import (
	"iot-bridge/internal/store"
	"sync"
)

type InMemoryStore struct {
	mu      sync.RWMutex
	devices map[string]store.Device
}

func New() store.DeviceStore {
	return &InMemoryStore{
		devices: make(map[string]store.Device),
	}
}

func (s *InMemoryStore) Add(d store.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[d.ID] = d
	return nil
}

func (s *InMemoryStore) GetAll() []store.Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []store.Device
	for _, d := range s.devices {
		list = append(list, d)
	}
	return list
}

func (s *InMemoryStore) Get(id string) (store.Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.devices[id]
	return d, ok
}

func (s *InMemoryStore) UpdateState(id string, updates map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.devices[id]
	if !ok {
		return nil
	}
	for k, v := range updates {
		d.State[k] = v
	}
	s.devices[id] = d
	return nil
}

func (s *InMemoryStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.devices, id)
	return nil
}
