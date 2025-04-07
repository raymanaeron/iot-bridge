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
		model:  "gpt-3.5-turbo", // or gpt-4
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

	// Get live context from DB
	deviceStore := factory.GetDeviceStore()
	devices := deviceStore.GetAll()

	var contextBlock strings.Builder
	contextBlock.WriteString("Known devices and their capabilities:\n")
	for _, d := range devices {
		contextBlock.WriteString(fmt.Sprintf("- ID: %s, Name: %s, Type: %s, Room: %s\n", d.ID, d.Name, d.Type, d.Room))
		for _, cap := range d.Capabilities {
			paramNames := make([]string, 0, len(cap.Parameters))
			for k := range cap.Parameters {
				paramNames = append(paramNames, k)
			}
			contextBlock.WriteString(fmt.Sprintf("  • %s (%s)\n", cap.Name, strings.Join(paramNames, ", ")))
		}
	}

	systemPrompt := fmt.Sprintf(`
You are an IoT planner.

Given a natural language command, convert it into a sequence of REST API calls based on the following API spec:

%s

%s

Rules:
- Each action must include the correct HTTP method, endpoint, and JSON body (if needed).
- Refer to capabilities and parameter names from the context above.
- Always respond with a JSON object like:
  {
    "actions": [
      {
        "method": "POST",
        "endpoint": "/devices/fan1/capabilities/speed",
        "body": { "level": 75 }
      }
    ]
  }
- DO NOT explain anything.
- DO NOT wrap output in markdown.
- DO NOT include any extra text.
`, apiDoc, contextBlock.String())

	payload := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},

			// Few-shot example: list devices
			{"role": "user", "content": "List all devices"},
			{"role": "assistant", "content": `{"actions":[{"method":"GET","endpoint":"/devices"}]}`},

			// Few-shot example: set fan speed
			{"role": "user", "content": "Set ceiling fan speed to 75%"},
			{"role": "assistant", "content": `{
  "actions": [
    {
      "method": "POST",
      "endpoint": "/devices/fan1/capabilities/speed",
      "body": { "level": 75 }
    }
  ]
}`},

			// Actual user prompt
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

// extractPureJSON strips ```json and extra text if OpenAI returns markdown
func extractPureJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown triple-backticks
	re := regexp.MustCompile("(?s)```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return s
}
