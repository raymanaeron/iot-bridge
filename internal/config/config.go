package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

var DemoMode bool
var LLMMode string

func LoadSettings() {
	_ = godotenv.Load(".env") // ignore error if .env does not exist

	val := os.Getenv("DEMO_MODE")
	demoMode, err := strconv.ParseBool(val)
	if err != nil {
		log.Println("Invalid DEMO_MODE value, defaulting to true")
		demoMode = true
	}
	DemoMode = demoMode
	log.Printf("DemoMode = %v\n", DemoMode)

	LLMMode = os.Getenv("LLM_MODE")
	if LLMMode == "" {
		LLMMode = "mock"
	}
	log.Printf("LLMMode = %v\n", LLMMode)
}
