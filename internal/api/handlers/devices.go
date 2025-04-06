package handlers

import (
	"encoding/json"
	"net/http"

	"iot-bridge/internal/iot"
	"iot-bridge/internal/store"
	"iot-bridge/internal/store/factory"

	"github.com/go-chi/chi/v5"
)

type AddFromScanRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Room string `json:"room"`
}

type PatchDeviceRequest struct {
	Name string `json:"name,omitempty"`
	Room string `json:"room,omitempty"`
}

func AddDeviceFromScan(w http.ResponseWriter, r *http.Request) {
	var req AddFromScanRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.ID == "" || req.Name == "" || req.Room == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	found, ok := factory.GetScanStore().FindDiscoveredDevice(req.ID)
	if !ok {
		http.Error(w, "Device not found in scan results", http.StatusNotFound)
		return
	}

	device := found.ToDevice(req.Name, req.Room)
	deviceStore := factory.GetDeviceStore()
	if err := deviceStore.Add(device); err != nil {
		http.Error(w, "Failed to add device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetDevices(w http.ResponseWriter, r *http.Request) {
	devices := factory.GetDeviceStore().GetAll()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func AddDevice(w http.ResponseWriter, r *http.Request) {
	var d store.Device
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil || d.ID == "" {
		http.Error(w, "Invalid device JSON or missing ID", http.StatusBadRequest)
		return
	}
	if err := factory.GetDeviceStore().Add(d); err != nil {
		http.Error(w, "Failed to add device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func GetDeviceByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	device, ok := factory.GetDeviceStore().Get(id)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	driver := iot.GetDriverFor(device)
	if state, err := driver.GetState(device); err == nil {
		device.State = state
	}

	json.NewEncoder(w).Encode(device)
}

func PatchDevice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	store := factory.GetDeviceStore()
	device, ok := store.Get(id)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	var req PatchDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		device.Name = req.Name
	}
	if req.Room != "" {
		device.Room = req.Room
	}

	if err := store.Add(device); err != nil {
		http.Error(w, "Failed to update device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteDevice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	store := factory.GetDeviceStore()
	_, ok := store.Get(id)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}
	if err := store.Delete(id); err != nil {
		http.Error(w, "Failed to delete device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
