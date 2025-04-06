package sqlite

import (
	"database/sql"
	"fmt"
	"iot-bridge/internal/store"
)

type SQLiteScanStore struct {
	db *sql.DB
}

func NewScanStore(db *sql.DB) store.ScanStore {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS discovered_devices (
		id TEXT PRIMARY KEY,
		name TEXT,
		type TEXT,
		protocol TEXT,
		signal INT
	);`)
	if err != nil {
		panic(fmt.Sprintf("Failed to create discovered_devices: %v", err))
	}
	return &SQLiteScanStore{db: db}
}

func (s *SQLiteScanStore) StartScan(protocols []string) {
	s.db.Exec("DELETE FROM discovered_devices")
	stmt, _ := s.db.Prepare(`INSERT INTO discovered_devices (id, name, type, protocol, signal) VALUES (?, ?, ?, ?, ?)`)
	stmt.Exec("bulb1", "Unregistered Bulb", "bulb", "zigbee", -42)
	stmt.Exec("plug1", "New Plug", "smart_plug", "zwave", -55)
}

func (s *SQLiteScanStore) GetScanResults() []store.DiscoveredDevice {
	rows, err := s.db.Query(`SELECT id, name, type, protocol, signal FROM discovered_devices`)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var devices []store.DiscoveredDevice
	for rows.Next() {
		var d store.DiscoveredDevice
		rows.Scan(&d.ID, &d.Name, &d.Type, &d.Protocol, &d.Signal)
		devices = append(devices, d)
	}
	return devices
}

func (s *SQLiteScanStore) FindDiscoveredDevice(id string) (store.DiscoveredDevice, bool) {
	row := s.db.QueryRow(`SELECT id, name, type, protocol, signal FROM discovered_devices WHERE id = ?`, id)
	var d store.DiscoveredDevice
	err := row.Scan(&d.ID, &d.Name, &d.Type, &d.Protocol, &d.Signal)
	if err != nil {
		return store.DiscoveredDevice{}, false
	}
	return d, true
}
