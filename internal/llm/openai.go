package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type OpenAI struct {
	apiKey string
	model  string
}

func NewOpenAI() LLMEngine {
	return &OpenAI{
		apiKey: os.Getenv("OPENAI_API_KEY"),
		model:  "gpt-3.5-turbo", // You can switch to gpt-4 if needed
	}
}

func (o *OpenAI) GeneratePlan(prompt string) (*Plan, error) {
	systemPrompt := `
You are an IoT planner.

Given a natural language command, convert it into a sequence of REST API actions.

Respond ONLY with a JSON object like this:

{
  "actions": [
    {
      "method": "PATCH",
      "endpoint": "/devices/plug1",
      "body": { "name": "flower_plug" }
    },
    {
      "method": "POST",
      "endpoint": "/devices/flower_plug/capabilities/power",
      "body": { "state": "on" }
    }
  ]
}

Do NOT explain your response.
Do NOT wrap the output in markdown (no triple backticks).
Respond with ONLY the JSON plan.
`

	payload := map[string]interface{}{
		"model": o.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
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
