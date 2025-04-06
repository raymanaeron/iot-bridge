package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Ollama struct{}

func NewOllama() LLMEngine {
	return &Ollama{}
}

func (o *Ollama) GeneratePlan(prompt string) (*Plan, error) {
	fullPrompt := fmt.Sprintf(`
You are an IoT planner.
Given a human command, respond ONLY with a JSON object like this:

{
  "actions": [
    {
      "method": "POST",
      "endpoint": "/devices/plug1/capabilities/power",
      "body": { "state": "on" }
    }
  ]
}

Do NOT explain. Do NOT wrap the response in markdown. Output ONLY valid JSON.

Now generate actions for this request:
"%s"
`, prompt)

	payload := map[string]interface{}{
		"model":  "phi",
		"prompt": fullPrompt,
		"stream": true,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var resultBuilder bytes.Buffer
	decoder := json.NewDecoder(resp.Body)

	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			fmt.Println("Decode error:", err)
			break
		}
		if text, ok := chunk["response"].(string); ok {
			resultBuilder.WriteString(text)
		}
	}

	finalOutput := resultBuilder.String()
	cleaned := extractPureJSON(finalOutput)

	var plan Plan
	if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
		return nil, fmt.Errorf("invalid LLM plan output: %v\nRAW OUTPUT:\n%s", err, finalOutput)
	}

	return &plan, nil
}

func extractPureJSON(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
