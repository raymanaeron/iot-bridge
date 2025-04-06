package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"iot-bridge/internal/store"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func New() store.DeviceStore {
	dbPath := filepath.Join(".", "devices.db")
	os.MkdirAll(filepath.Dir(dbPath), 0755)

	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		panic(fmt.Sprintf("Failed to open SQLite DB: %v", err))
	}

	createTable := `
	CREATE TABLE IF NOT EXISTS devices (
		id TEXT PRIMARY KEY,
		name TEXT,
		type TEXT,
		protocol TEXT,
		room TEXT,
		state TEXT
	);
	`
	if _, err := db.Exec(createTable); err != nil {
		panic(fmt.Sprintf("Failed to initialize schema: %v", err))
	}

	return &SQLiteStore{db: db}
}

func (s *SQLiteStore) Add(device store.Device) error {
	stateJSON, _ := json.Marshal(device.State)

	_, err := s.db.Exec(`INSERT OR REPLACE INTO devices (id, name, type, protocol, room, state) VALUES (?, ?, ?, ?, ?, ?)`,
		device.ID, device.Name, device.Type, device.Protocol, device.Room, string(stateJSON))
	return err
}

func (s *SQLiteStore) GetAll() []store.Device {
	rows, err := s.db.Query(`SELECT id, name, type, protocol, room, state FROM devices`)
	if err != nil {
		return []store.Device{}
	}
	defer rows.Close()

	var devices []store.Device
	for rows.Next() {
		var d store.Device
		var stateJSON string
		if err := rows.Scan(&d.ID, &d.Name, &d.Type, &d.Protocol, &d.Room, &stateJSON); err == nil {
			json.Unmarshal([]byte(stateJSON), &d.State)
			devices = append(devices, d)
		}
	}
	return devices
}

func (s *SQLiteStore) Get(id string) (store.Device, bool) {
	row := s.db.QueryRow(`SELECT id, name, type, protocol, room, state FROM devices WHERE id = ?`, id)

	var d store.Device
	var stateJSON string
	err := row.Scan(&d.ID, &d.Name, &d.Type, &d.Protocol, &d.Room, &stateJSON)
	if err != nil {
		return store.Device{}, false
	}
	json.Unmarshal([]byte(stateJSON), &d.State)
	return d, true
}

func (s *SQLiteStore) UpdateState(id string, updates map[string]string) error {
	device, found := s.Get(id)
	if !found {
		return errors.New("device not found")
	}

	for k, v := range updates {
		device.State[k] = v
	}
	return s.Add(device)
}

func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

func (s *SQLiteStore) Delete(id string) error {
	_, err := s.db.Exec(`DELETE FROM devices WHERE id = ?`, id)
	return err
}
