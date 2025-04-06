package llm

import (
	"fmt"
	"iot-bridge/internal/config"
)

var active LLMEngine

func Init() {
	switch config.LLMMode {
	case "mock":
		active = NewMockLLM()
		fmt.Println("LLM mode: mock")
	case "openai":
		active = NewOpenAI()
		fmt.Println("LLM mode: openai")
	case "ollama":
		active = NewOllama()
		fmt.Println("LLM mode: ollama")
	default:
		panic("Unsupported LLM mode: " + config.LLMMode)
	}
}

func GetEngine() LLMEngine {
	return active
}
