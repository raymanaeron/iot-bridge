package handlers

import (
	"encoding/json"
	"net/http"

	"iot-bridge/internal/store/factory"
)

func StartScan(w http.ResponseWriter, r *http.Request) {
	var req ScanRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.Protocols) == 0 {
		req.Protocols = []string{"zigbee", "zwave"}
	}
	factory.GetScanStore().StartScan(req.Protocols)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":    "scanning",
		"protocols": req.Protocols,
	})
}

func GetScanResults(w http.ResponseWriter, r *http.Request) {
	results := factory.GetScanStore().GetScanResults()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type ScanRequest struct {
	Protocols []string `json:"protocols"`
}
