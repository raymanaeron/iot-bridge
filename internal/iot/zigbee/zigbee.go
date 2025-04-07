package zigbee

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"iot-bridge/internal/store"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	deviceStates = make(map[string]map[string]string)
	stateMu      sync.RWMutex
	mqttClient   mqtt.Client
)

type ZigbeeDriver struct{}

func Init() {
	opts := mqtt.NewClientOptions().AddBroker("tcp://localhost:1883")
	opts.SetClientID("iot-bridge-zigbee")
	opts.OnConnect = func(c mqtt.Client) {
		log.Println("[Zigbee] Connected to MQTT")
		if token := c.Subscribe("zigbee2mqtt/+", 0, messageHandler); token.Wait() && token.Error() != nil {
			log.Println("[Zigbee] Failed to subscribe:", token.Error())
		}
	}
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("[Zigbee] MQTT connection error:", token.Error())
	}
}

func GetDriver() *ZigbeeDriver {
	return &ZigbeeDriver{}
}

func messageHandler(client mqtt.Client, msg mqtt.Message) {
	topic := msg.Topic()
	payload := msg.Payload()

	if strings.HasSuffix(topic, "/set") {
		return // Ignore set acknowledgments
	}

	deviceID := strings.TrimPrefix(topic, "zigbee2mqtt/")
	var state map[string]interface{}
	if err := json.Unmarshal(payload, &state); err != nil {
		log.Printf("[Zigbee] Invalid state for %s: %v", deviceID, err)
		return
	}

	stringState := make(map[string]string)
	for k, v := range state {
		stringState[k] = fmt.Sprintf("%v", v)
	}

	stateMu.Lock()
	deviceStates[deviceID] = stringState
	stateMu.Unlock()

	log.Printf("[Zigbee] Updated state for %s: %+v", deviceID, stringState)
}

func (z *ZigbeeDriver) GetState(device store.Device) (map[string]string, error) {
	stateMu.RLock()
	defer stateMu.RUnlock()
	s, ok := deviceStates[device.ID]
	if !ok {
		return nil, fmt.Errorf("no state for device %s", device.ID)
	}
	return s, nil
}

func (z *ZigbeeDriver) SetState(device store.Device, updates map[string]string) error {
	data, _ := json.Marshal(updates)
	topic := fmt.Sprintf("zigbee2mqtt/%s/set", device.ID)
	token := mqttClient.Publish(topic, 0, false, data)
	token.Wait()
	return token.Error()
}
