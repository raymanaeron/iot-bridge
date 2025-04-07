package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iot-bridge/internal/llm"
	"iot-bridge/internal/store/factory"
	"net/http"
	"strings"
)

type LLMRequest struct {
	Prompt string `json:"prompt"`
}

type ActionResult struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Status   string `json:"status"`
}

// 🧠 Match a known device name to its ID
func resolveDeviceIDFromName(name string) (string, bool) {
	devices := factory.GetDeviceStore().GetAll()
	for _, d := range devices {
		if strings.EqualFold(d.Name, name) {
			return d.ID, true
		}
	}
	return "", false
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

	// 🧠 Patch any placeholder {id} with real ID from prompt
	for i, action := range plan.Actions {
		if strings.Contains(action.Endpoint, "{id}") {
			// Try matching any known device name inside the original prompt
			devices := factory.GetDeviceStore().GetAll()
			for _, d := range devices {
				if strings.Contains(strings.ToLower(req.Prompt), strings.ToLower(d.Name)) {
					plan.Actions[i].Endpoint = strings.Replace(action.Endpoint, "{id}", d.ID, 1)
					break
				}
			}
		}
	}

	var results []ActionResult
	for _, action := range plan.Actions {
		fullURL := fmt.Sprintf("http://localhost:8080%s", action.Endpoint)

		req, err := http.NewRequest(action.Method, fullURL, bytes.NewReader(action.Body))
		if err != nil {
			results = append(results, ActionResult{
				Endpoint: action.Endpoint,
				Method:   action.Method,
				Status:   "failed (invalid request)",
			})
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			results = append(results, ActionResult{
				Endpoint: action.Endpoint,
				Method:   action.Method,
				Status:   "failed (HTTP error)",
			})
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		status := "success"
		if resp.StatusCode >= 300 {
			status = fmt.Sprintf("failed (%s)", resp.Status)
		}

		results = append(results, ActionResult{
			Endpoint: action.Endpoint,
			Method:   action.Method,
			Status:   status,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"prompt":  req.Prompt,
		"actions": results,
	})
}
