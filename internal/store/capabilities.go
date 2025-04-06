package store

type Capability struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Operations  []string               `json:"operations"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

func GetCapabilitiesForType(deviceType string) []Capability {
	switch deviceType {
	case "bulb":
		return []Capability{
			{
				Name:        "power",
				Description: "Turn the bulb on or off",
				Operations:  []string{"on", "off"},
			},
			{
				Name:        "brightness",
				Description: "Adjust brightness (0-100)",
				Operations:  []string{"set"},
				Parameters: map[string]interface{}{
					"level": map[string]interface{}{
						"type":  "integer",
						"range": []int{0, 100},
					},
				},
			},
			{
				Name:        "color",
				Description: "Change bulb color using RGB",
				Operations:  []string{"set"},
				Parameters: map[string]interface{}{
					"rgb": map[string]interface{}{
						"type":   "array",
						"length": 3,
						"range":  []int{0, 255},
					},
				},
			},
		}
	case "switch", "smart_plug":
		return []Capability{
			{
				Name:        "power",
				Description: "Turn the device on or off",
				Operations:  []string{"on", "off"},
			},
		}
	default:
		return []Capability{}
	}
}
