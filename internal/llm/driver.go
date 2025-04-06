package llm

import (
	"encoding/json"
)

type PlannedAction struct {
	Method   string          `json:"method"`
	Endpoint string          `json:"endpoint"`
	Body     json.RawMessage `json:"body"`
}

type Plan struct {
	Actions []PlannedAction `json:"actions"`
}

type LLMEngine interface {
	GeneratePlan(prompt string) (*Plan, error)
}
