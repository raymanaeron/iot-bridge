package llm

import "encoding/json"

type MockLLM struct{}

func NewMockLLM() LLMEngine {
	return &MockLLM{}
}

func (m *MockLLM) GeneratePlan(prompt string) (*Plan, error) {
	if prompt == "turn on the living room bulb" {
		body, _ := json.Marshal(map[string]string{"state": "on"})
		return &Plan{
			Actions: []PlannedAction{
				{
					Method:   "POST",
					Endpoint: "/devices/bulb1/capabilities/power",
					Body:     body,
				},
			},
		}, nil
	}
	return &Plan{}, nil
}
