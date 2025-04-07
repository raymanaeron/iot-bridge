package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"iot-bridge/internal/iot"
	storemodel "iot-bridge/internal/store"
	"iot-bridge/internal/store/factory"
)

func GetCapabilities(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	store := factory.GetDeviceStore()
	device, ok := store.Get(deviceID)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// capabilities := storemodel.GetCapabilitiesForType(device.Type)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":   device.ID,
		"type": device.Type,
		"name": device.Name,
		//"capabilities": capabilities,
		"capabilities": device.Capabilities,
	})
}

func UpdateCapabilities(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	store := factory.GetDeviceStore()

	var newCaps []storemodel.Capability
	if err := json.NewDecoder(r.Body).Decode(&newCaps); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	device, ok := store.Get(deviceID)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	device.Capabilities = newCaps
	if err := store.Add(device); err != nil {
		http.Error(w, "Failed to update capabilities", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func InvokeCapability(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	capabilityName := chi.URLParam(r, "capability")

	store := factory.GetDeviceStore()
	device, ok := store.Get(deviceID)
	if !ok {
		http.Error(w, "Device not found", http.StatusNotFound)
		return
	}

	// Get capability definition
	//caps := storemodel.GetCapabilitiesForType(device.Type)
	caps := device.Capabilities

	var selectedCap *storemodel.Capability
	for _, cap := range caps {
		if cap.Name == capabilityName {
			selectedCap = &cap
			break
		}
	}
	if selectedCap == nil {
		http.Error(w, "Unsupported capability", http.StatusBadRequest)
		return
	}

	// Parse input
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Validate against capability definition
	validated := map[string]string{}
	for paramName, paramSpecRaw := range selectedCap.Parameters {
		spec := paramSpecRaw.(map[string]interface{})

		value, ok := input[paramName]
		if !ok {
			http.Error(w, fmt.Sprintf("Missing required parameter: %s", paramName), http.StatusBadRequest)
			return
		}

		switch spec["type"] {
		case "integer":
			var valFloat float64
			switch v := value.(type) {
			case float64:
				valFloat = v
			case int:
				valFloat = float64(v)
			case int64:
				valFloat = float64(v)
			default:
				http.Error(w, fmt.Sprintf("Parameter '%s' must be a number", paramName), http.StatusBadRequest)
				return
			}

			// Range check
			if bounds, ok := spec["range"].([]interface{}); ok && len(bounds) == 2 {
				min := int(bounds[0].(float64))
				max := int(bounds[1].(float64))
				if int(valFloat) < min || int(valFloat) > max {
					http.Error(w, fmt.Sprintf("Parameter '%s' out of range", paramName), http.StatusBadRequest)
					return
				}
			}
			validated[paramName] = fmt.Sprintf("%d", int(valFloat))

		case "array":
			arr, ok := value.([]interface{})
			if !ok {
				http.Error(w, fmt.Sprintf("Parameter '%s' must be an array", paramName), http.StatusBadRequest)
				return
			}
			length := int(spec["length"].(float64))
			if len(arr) != length {
				http.Error(w, fmt.Sprintf("Parameter '%s' must be an array of length %d", paramName, length), http.StatusBadRequest)
				return
			}
			validated[paramName] = fmt.Sprintf("[%v,%v,%v]", int(arr[0].(float64)), int(arr[1].(float64)), int(arr[2].(float64)))

		default:
			strVal := fmt.Sprintf("%v", value)
			if len(selectedCap.Operations) > 0 {
				valid := false
				for _, op := range selectedCap.Operations {
					if strVal == op {
						valid = true
						break
					}
				}
				if !valid {
					http.Error(w, fmt.Sprintf("Invalid value for '%s'", paramName), http.StatusBadRequest)
					return
				}
			}
			validated[paramName] = strVal
		}
	}

	// Send to device
	driver := iot.GetDriverFor(device)
	if err := driver.SetState(device, validated); err != nil {
		http.Error(w, fmt.Sprintf("Failed to communicate with device: %v", err), http.StatusBadGateway)
		return
	}

	if err := store.UpdateState(deviceID, validated); err != nil {
		http.Error(w, "Failed to persist device state", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "success",
		"capability": capabilityName,
		"new_state":  validated,
	})
}

func GetCapability(deviceType string, capabilityName string) (storemodel.Capability, bool) {
	for _, cap := range storemodel.GetCapabilitiesForType(deviceType) {
		if cap.Name == capabilityName {
			return cap, true
		}
	}
	return storemodel.Capability{}, false
}
