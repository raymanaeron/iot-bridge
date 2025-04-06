package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Ollama struct{}

func NewOllama() LLMEngine {
	return &Ollama{}
}

func (o *Ollama) GeneratePlan(prompt string) (*Plan, error) {
	fullPrompt := fmt.Sprintf(`
You are an IoT planner.
Given a human command, respond ONLY with a JSON object like:

{
  "actions": [
    {
      "method": "POST",
      "endpoint": "/devices/plug1/capabilities/power",
      "body": { "state": "on" }
    }
  ]
}

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

	// Read streaming JSON chunks
	var resultBuilder bytes.Buffer
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var chunk map[string]string
		if err := decoder.Decode(&chunk); err != nil {
			fmt.Println("Error decoding chunk:", err)
			break
		}
		fmt.Printf("Chunk: %+v\n", chunk)

		if text, ok := chunk["response"]; ok {
			resultBuilder.WriteString(text)
		}
	}

	finalOutput := resultBuilder.String()

	var plan Plan
	if err := json.Unmarshal([]byte(finalOutput), &plan); err != nil {
		return nil, fmt.Errorf("invalid LLM plan output: %v\nRAW OUTPUT:\n%s", err, finalOutput)
	}

	return &plan, nil
}
