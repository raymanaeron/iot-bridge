package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"iot-bridge/internal/store/factory"
)

type OpenAI struct {
	apiKey string
	model  string
}

func NewOpenAI() LLMEngine {
	return &OpenAI{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		model:  "gpt-3.5-turbo",
	}
}

func (o *OpenAI) GeneratePlan(prompt string) (*Plan, error) {
	apiDoc := `
Available API endpoints:

- GET /devices → List all devices
- GET /devices/{id} → Get info for a device
- POST /devices → Add a new device
- PATCH /devices/{id} → Update device metadata (name, room, etc)
- DELETE /devices/{id} → Remove a device
- GET /devices/{id}/capabilities → List what actions a device supports
- POST /devices/{id}/capabilities/{capability} → Turn on/off, change brightness, set speed, etc.
- POST /scan → Start device scan
- GET /scan/results → Retrieve discovered devices
`

	deviceStore := factory.GetDeviceStore()
	devices := deviceStore.GetAll()

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Known devices and capabilities:\n")
	for _, d := range devices {
		contextBuilder.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Room: %s\n", d.ID, d.Name, d.Type, d.Room))
		for _, cap := range d.Capabilities {
			var paramList []string
			for k := range cap.Parameters {
				paramList = append(paramList, k)
			}
			contextBuilder.WriteString(fmt.Sprintf("  • %s (%s)\n", cap.Name, strings.Join(paramList, ", ")))
		}
	}

	systemPrompt := fmt.Sprintf(`
You are an IoT planner.

Given a natural language command, convert it into a sequence of REST API calls based on the following API specification:

%s

%s

Rules:
- Only use device IDs and capabilities provided in the context.
- Include the correct method, endpoint, and JSON body for each action.
- Ensure parameter names and value types match what is available.
- Respond only with a valid JSON object in this format:
  {
    "actions": [
      {
        "method": "POST",
        "endpoint": "/devices/plug1/capabilities/power",
        "body": { "state": "on" }
      }
    ]
  }
- Do not include explanations or markdown.
`, apiDoc, contextBuilder.String())

	payload := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "List all devices"},
			{"role": "assistant", "content": `{"actions":[{"method":"GET","endpoint":"/devices"}]}`},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.2,
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %v\n%s", err, string(raw))
	}

	clean := extractPureJSON(result.Choices[0].Message.Content)

	var plan Plan
	if err := json.Unmarshal([]byte(clean), &plan); err != nil {
		return nil, fmt.Errorf("invalid LLM plan output: %v\nRAW:\n%s", err, clean)
	}

	return &plan, nil
}

// extractPureJSON strips markdown ``` wrappers or formatting text if present
func extractPureJSON(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile("(?s)```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return s
}
