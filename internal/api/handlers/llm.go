package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"iot-bridge/internal/llm"
	"iot-bridge/internal/store"
)

type LLMRequest struct {
	Prompt string `json:"prompt"`
}

type ActionResult struct {
	Method   string `json:"method"`
	Endpoint string `json:"endpoint"`
	Status   string `json:"status"`
}

type DeviceWithCapabilities struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Type         string             `json:"type"`
	Room         string             `json:"room"`
	Capabilities []store.Capability `json:"capabilities"`
}

func HandleLLMRequest(w http.ResponseWriter, r *http.Request) {
	var req LLMRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Prompt == "" {
		http.Error(w, "Missing or invalid prompt", http.StatusBadRequest)
		return
	}

	engine := llm.GetEngine()
	plan, err := engine.GeneratePlan(req.Prompt)
	if err != nil {
		http.Error(w, "LLM error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []ActionResult
	var formattedDevices []DeviceWithCapabilities

	for _, action := range plan.Actions {
		fullURL := fmt.Sprintf("http://localhost:8080%s", action.Endpoint)

		var method = strings.ToUpper(action.Method)
		var body io.Reader
		if len(action.Body) > 0 {
			body = bytes.NewReader(action.Body)
		}

		req, err := http.NewRequest(method, fullURL, body)
		if err != nil {
			results = append(results, ActionResult{
				Method:   method,
				Endpoint: action.Endpoint,
				Status:   "failed (invalid request)",
			})
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			results = append(results, ActionResult{
				Method:   method,
				Endpoint: action.Endpoint,
				Status:   "failed (HTTP error)",
			})
			continue
		}
		defer resp.Body.Close()

		status := "success"
		if resp.StatusCode >= 300 {
			status = fmt.Sprintf("failed (%s)", resp.Status)
		}

		results = append(results, ActionResult{
			Method:   method,
			Endpoint: action.Endpoint,
			Status:   status,
		})

		// Custom logic: If this is GET /devices → read all devices
		if method == "GET" && action.Endpoint == "/devices" && resp.StatusCode == 200 {
			var devices []store.Device
			json.NewDecoder(resp.Body).Decode(&devices)

			// pre-fill devices for capability fetch
			for _, d := range devices {
				formattedDevices = append(formattedDevices, DeviceWithCapabilities{
					ID:   d.ID,
					Name: d.Name,
					Type: d.Type,
					Room: d.Room,
				})
			}
		}

		// Populate capabilities into formattedDevices if this is a GET /devices/{id}/capabilities
		if method == "GET" && strings.Contains(action.Endpoint, "/capabilities") && resp.StatusCode == 200 {
			parts := strings.Split(action.Endpoint, "/")
			if len(parts) >= 3 {
				deviceID := parts[2]
				var capsResp struct {
					Capabilities []store.Capability `json:"capabilities"`
				}
				json.NewDecoder(resp.Body).Decode(&capsResp)

				for i := range formattedDevices {
					if formattedDevices[i].ID == deviceID {
						formattedDevices[i].Capabilities = capsResp.Capabilities
					}
				}
			}
		}
	}

	// Response structure
	respData := map[string]interface{}{
		"prompt":  req.Prompt,
		"actions": results,
	}
	if len(formattedDevices) > 0 {
		respData["devices"] = formattedDevices
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respData)
}
