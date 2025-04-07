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
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./web/index.html")
		return
	}

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

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("You: %s\n", req.Prompt))

	for _, action := range plan.Actions {
		method := strings.ToUpper(action.Method)
		endpoint := action.Endpoint
		fullURL := fmt.Sprintf("http://localhost:8080%s", endpoint)

		var reqBody io.Reader
		if len(action.Body) > 0 {
			reqBody = bytes.NewReader(action.Body)
		}

		httpReq, err := http.NewRequest(method, fullURL, reqBody)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("→ %s %s\n   failed (invalid request)\n\n", method, endpoint))
			continue
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			buffer.WriteString(fmt.Sprintf("→ %s %s\n   failed (HTTP error)\n\n", method, endpoint))
			continue
		}
		defer resp.Body.Close()

		status := "success"
		if resp.StatusCode >= 300 {
			status = fmt.Sprintf("failed (%s)", resp.Status)
		}
		buffer.WriteString(fmt.Sprintf("→ %s %s\n   %s\n", method, endpoint, status))

		// If GET /devices, also print devices with capabilities
		if method == "GET" && endpoint == "/devices" && resp.StatusCode == 200 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			var devices []map[string]interface{}
			json.Unmarshal(bodyBytes, &devices)

			for _, d := range devices {
				id := d["id"].(string)
				name := d["name"].(string)
				deviceType := d["type"].(string)
				room := d["room"].(string)

				buffer.WriteString(fmt.Sprintf("\n• %s - %s (%s in %s)\n", id, name, deviceType, room))

				// Fetch capabilities
				capURL := fmt.Sprintf("http://localhost:8080/devices/%s/capabilities", id)
				capResp, err := http.Get(capURL)
				if err != nil || capResp.StatusCode != 200 {
					buffer.WriteString("   (failed to retrieve capabilities)\n")
					continue
				}
				defer capResp.Body.Close()

				var capBody struct {
					Capabilities []struct {
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"capabilities"`
				}
				json.NewDecoder(capResp.Body).Decode(&capBody)

				for _, cap := range capBody.Capabilities {
					buffer.WriteString(fmt.Sprintf("  - %s: %s\n", cap.Name, cap.Description))
				}
			}
			buffer.WriteString("\n")
		} else {
			// Drain the response body for other endpoints
			io.Copy(io.Discard, resp.Body)
		}

		buffer.WriteString("\n")
	}

	// Return plain text or JSON
	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"prompt":  req.Prompt,
			"summary": buffer.String(),
		})
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(buffer.String()))
}
