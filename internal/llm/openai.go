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

	var apiDoc = `
	Available API endpoints:

	- GET /devices → List all devices
	- GET /devices/{id} → Get info for a device
	- POST /devices → Add a new device
	- PATCH /devices/{id} → Update device metadata
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
	- Respond ONLY with a JSON object that includes "actions".
	- DO NOT explain your response.
	- DO NOT wrap the output in markdown.
	`, apiDoc)

	/*
	   	systemPrompt := `
	   You are an IoT planner.

	   Device IDs like "plug1" or "bulb2" are permanent identifiers and do not change.

	   Users may rename a device by updating its 'name' field via:
	   PATCH /devices/{id}

	   But the ID remains the same and must still be used in all future actions.

	   Respond ONLY with a JSON object like this:

	   {
	     "actions": [
	       {
	         "method": "PATCH",
	         "endpoint": "/devices/plug1",
	         "body": { "name": "multi-plug" }
	       },
	       {
	         "method": "POST",
	         "endpoint": "/devices/plug1/capabilities/power",
	         "body": { "state": "on" }
	       }
	     ]
	   }

	   Do NOT wrap the output in markdown.
	   Do NOT include any explanation.
	   Respond ONLY with the JSON object.
	   `
	*/

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
