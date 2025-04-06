package llm

type PlannedAction struct {
	Method   string `json:"method"`
	Endpoint string `json:"endpoint"`
	Body     []byte `json:"body"`
}

type Plan struct {
	Actions []PlannedAction `json:"actions"`
}

type LLMEngine interface {
	GeneratePlan(prompt string) (*Plan, error)
}
