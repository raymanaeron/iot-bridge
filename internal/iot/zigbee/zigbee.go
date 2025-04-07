package zigbee

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
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
		return
	}

	deviceID := strings.TrimPrefix(topic, "zigbee2mqtt/")
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		log.Printf("[Zigbee] Invalid state for %s: %v", deviceID, err)
		return
	}

	stringState := make(map[string]string)
	for k, v := range raw {
		stringState[k] = fmt.Sprintf("%v", v)
	}

	stateMu.Lock()
	deviceStates[deviceID] = stringState
	stateMu.Unlock()

	ds := factory.GetDeviceStore()
	_, found := ds.Get(deviceID)
	if !found {
		log.Printf("[Zigbee] Discovered device: %s", deviceID)
		newDevice := store.Device{
			ID:           deviceID,
			Name:         deviceID,
			Type:         "zigbee",
			Protocol:     "zigbee",
			Room:         "unknown",
			State:        stringState,
			Capabilities: inferCapabilitiesFromPayload(raw),
		}
		if err := ds.Add(newDevice); err != nil {
			log.Printf("[Zigbee] Failed to add device %s: %v", deviceID, err)
		} else {
			log.Printf("[Zigbee] Registered new device: %s", deviceID)
		}
	} else {
		if err := ds.UpdateState(deviceID, stringState); err != nil {
			log.Printf("[Zigbee] Failed to update state for %s: %v", deviceID, err)
		}
	}
}

func inferCapabilitiesFromPayload(payload map[string]interface{}) []store.Capability {
	var caps []store.Capability
	for key, value := range payload {
		caps = append(caps, store.Capability{
			Name:        key,
			Description: fmt.Sprintf("Auto-discovered capability for '%s'", key),
			Writable:    inferWritable(key, value),
			Parameters: map[string]interface{}{
				key: inferParamSpec(value),
			},
		})
	}
	return caps
}

func inferParamSpec(value interface{}) map[string]interface{} {
	typ := reflect.TypeOf(value)
	spec := map[string]interface{}{}

	switch typ.Kind() {
	case reflect.Float64:
		spec["type"] = "integer"
		spec["range"] = []int{0, 100}
	case reflect.Bool:
		spec["type"] = "boolean"
	case reflect.String:
		spec["type"] = "string"
	default:
		spec["type"] = "string"
	}
	return spec
}

func inferWritable(key string, value interface{}) bool {
	// Heuristically mark as writable if its current value is a string and appears to accept discrete commands
	strVal, ok := value.(string)
	if !ok {
		return false
	}

	// Common keys that are likely writable based on behavior
	controlKeywords := []string{"state", "power", "command", "mode", "level", "brightness", "speed", "volume"}
	for _, keyword := range controlKeywords {
		if strings.Contains(strings.ToLower(key), keyword) {
			return true
		}
	}

	// If the value is short and uppercase, likely a command
	if len(strVal) <= 6 && strVal == strings.ToUpper(strVal) {
		return true
	}

	return false
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
