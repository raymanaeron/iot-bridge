package main

import (
	"iot-bridge/internal/api"
	"iot-bridge/internal/config"
	"iot-bridge/internal/iot"
	"iot-bridge/internal/iot/zigbee"
	llmfactory "iot-bridge/internal/llm"
	"iot-bridge/internal/store/factory"

	"log"
	"net/http"
)

func main() {
	config.LoadSettings()
	factory.Init()
	zigbee.Init()
	iot.Init()
	llmfactory.Init()
	router := api.NewRouter()
	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
