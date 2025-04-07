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

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("You: %s\n", req.Prompt))

	for _, action := range plan.Actions {
		fullURL := fmt.Sprintf("http://localhost:8080%s", action.Endpoint)

		reqBody := bytes.NewReader(action.Body)
		apiReq, err := http.NewRequest(action.Method, fullURL, reqBody)
		if err != nil {
			sb.WriteString(fmt.Sprintf("→ %s %s\n   failed (invalid request)\n", action.Method, action.Endpoint))
			continue
		}
		apiReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(apiReq)
		if err != nil {
			sb.WriteString(fmt.Sprintf("→ %s %s\n   failed (http error)\n", action.Method, action.Endpoint))
			continue
		}
		defer resp.Body.Close()

		sb.WriteString(fmt.Sprintf("→ %s %s\n   %s\n", action.Method, action.Endpoint, resp.Status))

		respText, _ := io.ReadAll(resp.Body)
		if len(respText) > 0 {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, respText, "   ", "  "); err != nil {
				// fallback if not JSON
				sb.WriteString(indent(string(respText), "   ") + "\n")
			} else {
				sb.WriteString(pretty.String() + "\n")
			}
		}

	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(sb.String()))
}

func indent(text string, prefix string) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
