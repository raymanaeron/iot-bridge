package zigbee

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	"iot-bridge/internal/store"
	"iot-bridge/internal/store/factory"

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

		// State updates
		if token := c.Subscribe("zigbee2mqtt/+", 0, messageHandler); token.Wait() && token.Error() != nil {
			log.Println("[Zigbee] Failed to subscribe to device state:", token.Error())
		}

		// Device discovery
		if token := c.Subscribe("zigbee2mqtt/bridge/devices", 0, deviceListHandler); token.Wait() && token.Error() != nil {
			log.Println("[Zigbee] Failed to subscribe to device list:", token.Error())
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
		return // Ignore set commands
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

func deviceListHandler(client mqtt.Client, msg mqtt.Message) {
	var devices []struct {
		FriendlyName string `json:"friendly_name"`
		ModelID      string `json:"model_id"`
		Description  string `json:"description"`
	}

	if err := json.Unmarshal(msg.Payload(), &devices); err != nil {
		log.Println("[Zigbee] Failed to parse device list:", err)
		return
	}

	for _, d := range devices {
		log.Printf("[Zigbee] Discovered device: %s (%s)", d.FriendlyName, d.ModelID)

		// Attempt to auto-register
		autoRegisterZigbeeDevice(store.Device{
			ID:       d.FriendlyName,
			Name:     d.Description,
			Type:     "bulb", // You can improve this with model/type mapping
			Protocol: "zigbee",
			Room:     "Unknown",
			State:    map[string]string{},
		})
	}
}

func autoRegisterZigbeeDevice(dev store.Device) {
	ds := factory.GetDeviceStore()
	if _, found := ds.Get(dev.ID); found {
		return // Already exists
	}
	if err := ds.Add(dev); err != nil {
		log.Printf("[Zigbee] Failed to add device %s: %v", dev.ID, err)
	} else {
		log.Printf("[Zigbee] Registered new device: %s", dev.ID)
	}
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

// Optional: Call this when /scan is hit
func StartPermitJoin(duration int) {
	payload := fmt.Sprintf(`{"value":true,"time":%d}`, duration)
	mqttClient.Publish("zigbee2mqtt/bridge/request/permit_join", 0, false, payload)
}
