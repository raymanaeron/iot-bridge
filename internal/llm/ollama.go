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
	payload := map[string]string{
		"model":  "phi",
		"prompt": fmt.Sprintf("Given this command, generate a list of JSON actions to control an IoT system: %s", prompt),
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// LLM responds with stringified JSON — now decode that
	var plan Plan
	if err := json.Unmarshal([]byte(response.Response), &plan); err != nil {
		return nil, fmt.Errorf("invalid LLM plan output: %v", err)
	}

	return &plan, nil
}
