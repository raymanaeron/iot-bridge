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
	
	Your job is to convert a user's instruction into a structured JSON action plan.
	
	Respond ONLY with a JSON object. Do NOT wrap it in triple-backticks. Do NOT explain anything.
	
	Example format:
	
	{
	  "actions": [
		{
		  "method": "POST",
		  "endpoint": "/devices/plug1/capabilities/power",
		  "body": { "state": "on" }
		}
	  ]
	}
	
	Now generate actions for this command:
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
		jsonBlock := s[start : end+1]

		// Clean up common markdown junk
		jsonBlock = strings.TrimPrefix(jsonBlock, "```json")
		jsonBlock = strings.TrimPrefix(jsonBlock, "```")
		jsonBlock = strings.TrimSuffix(jsonBlock, "```")

		return strings.TrimSpace(jsonBlock)
	}
	return ""
}
