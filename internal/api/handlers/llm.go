package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"iot-bridge/internal/llm"
	"net/http"
)

type LLMRequest struct {
	Prompt string `json:"prompt"`
}

type ActionResult struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Status   string `json:"status"`
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
