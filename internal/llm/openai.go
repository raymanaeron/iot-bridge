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
		contextBuilder.WriteString(fmt.Sprintf("- Name: \"%s\", ID: \"%s\", Type: %s, Room: %s\n", d.Name, d.ID, d.Type, d.Room))
		for _, cap := range d.Capabilities {
			var paramList []string
			for k := range cap.Parameters {
				paramList = append(paramList, k)
			}
			writability := "read-only"
			if cap.Writable {
				writability = "writable"
			}
			contextBuilder.WriteString(fmt.Sprintf("  • %s (%s) — %s\n", cap.Name, strings.Join(paramList, ", "), writability))
		}
	}

	systemPrompt := fmt.Sprintf(`You are an IoT planner.

You convert natural language commands into precise REST API plans.

Use this API spec:

%s

%s

Instructions:
- Use exact device names or IDs from the context above. Do NOT guess or assume.
- Prefer writable capabilities for POST requests. Read-only capabilities should never be set.
- Match parameter names and types carefully from context.
- Output only a JSON object like this:
  {
    "actions": [
      {
        "method": "POST",
        "endpoint": "/devices/plug1/capabilities/power",
        "body": { "state": "on" }
      }
    ]
  }
- No explanations. No markdown. Only pure JSON.
`, apiDoc, contextBuilder.String())

	payload := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},

			// Few-shot examples
			{"role": "user", "content": "List all devices"},
			{"role": "assistant", "content": `{"actions":[{"method":"GET","endpoint":"/devices"}]}`},

			{"role": "user", "content": "Turn on Living Room Bulb"},
			{"role": "assistant", "content": `{
  "actions": [
    {
      "method": "POST",
      "endpoint": "/devices/bulb1/capabilities/power",
      "body": { "state": "on" }
    }
  ]
}`},

			// Actual user input
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
