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
)

type OpenAI struct {
	apiKey string
	model  string
}

func NewOpenAI() LLMEngine {
	return &OpenAI{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		model:  "gpt-3.5-turbo", // or "gpt-4"
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
- POST /devices/{id}/capabilities/{capability} → Turn on/off, change brightness, etc.
- POST /scan → Start device scan
- GET /scan/results → Retrieve discovered devices
`

	systemPrompt := fmt.Sprintf(`
You are an IoT planner.

Given a human command, your job is to plan a series of REST API calls based on the following API spec:

%s

Rules:
- Respond ONLY with a JSON object like:
  {
    "actions": [
      {
        "method": "GET",
        "endpoint": "/devices"
      }
    ]
  }
- DO NOT include any explanation.
- DO NOT wrap your response in markdown.
- DO NOT output anything other than valid JSON.
`, apiDoc)

	// Few-shot learning: show what kind of response we expect
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

// extractPureJSON removes markdown formatting (e.g., ```json ... ```) and trims content
func extractPureJSON(s string) string {
	s = strings.TrimSpace(s)

	// Remove markdown triple backticks and optional "json" label
	re := regexp.MustCompile("(?s)```(?:json)?(.*?)```")
	matches := re.FindStringSubmatch(s)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// If no markdown, assume it's raw JSON already
	return s
}
