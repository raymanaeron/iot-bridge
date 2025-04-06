package llm

import (
	"fmt"
)

type OpenAI struct{}

func NewOpenAI() LLMEngine {
	return &OpenAI{}
}

func (o *OpenAI) GeneratePlan(prompt string) (*Plan, error) {
	return nil, fmt.Errorf("OpenAI integration not implemented yet")
}
