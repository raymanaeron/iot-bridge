#!/bin/bash

baseUrl="http://localhost:8080"

echo "🧹 Deleting all existing devices..."
existing=$(curl -s "$baseUrl/devices" | jq -r '.[].id')
for id in $existing; do
  echo "❌ Deleting $id"
  curl -s -X DELETE "$baseUrl/devices/$id" > /dev/null
done
echo "✅ Clean slate ready."

echo ""
echo "➕ Adding test devices..."

add_device() {
  local id=$1
  local payload=$2
  response=$(curl -s -w "%{http_code}" -o /tmp/add_device_response.json \
    -X POST "$baseUrl/devices" \
    -H "Content-Type: application/json" \
    -d "$payload")

  if [[ "$response" == "201" ]]; then
    echo "✅ $id added"
  else
    echo "❌ Failed to add $id (HTTP $response):"
    cat /tmp/add_device_response.json
  fi
}

add_device "bulb1" '{
  "id": "bulb1",
  "name": "Living Room Bulb",
  "type": "bulb",
  "protocol": "zigbee",
  "room": "Living Room",
  "state": {}
}'

add_device "plug1" '{
  "id": "plug1",
  "name": "Coffee Machine Plug",
  "type": "smart_plug",
  "protocol": "zwave",
  "room": "Kitchen",
  "state": {}
}'

add_device "fan1" '{
  "id": "fan1",
  "name": "Ceiling Fan",
  "type": "fan",
  "protocol": "zigbee",
  "room": "Bedroom",
  "state": {}
}'

add_device "speaker1" '{
  "id": "speaker1",
  "name": "Kitchen Speaker",
  "type": "speaker",
  "protocol": "wifi",
  "room": "Kitchen",
  "state": {}
}'

add_device "coffee1" '{
  "id": "coffee1",
  "name": "Coffee Maker",
  "type": "smart_appliance",
  "protocol": "zwave",
  "room": "Kitchen",
  "state": {}
}'

echo ""
echo "🧠 Adding device capabilities..."

post_caps() {
  device_id=$1
  payload=$2
  response=$(curl -s -w "%{http_code}" -o /tmp/cap_resp.json \
    -X POST "$baseUrl/devices/$device_id/capabilities" \
    -H "Content-Type: application/json" \
    -d "$payload")

  if [[ "$response" == "204" ]]; then
    echo "✅ Capabilities set for $device_id"
  else
    echo "❌ Failed to set capabilities for $device_id (HTTP $response):"
    cat /tmp/cap_resp.json
  fi
}

post_caps bulb1 '[
  {
    "name": "power",
    "description": "Turn bulb on or off",
    "parameters": { "state": { "type": "string", "operations": ["on", "off"] } },
    "operations": ["on", "off"]
  },
  {
    "name": "brightness",
    "description": "Adjust brightness level",
    "parameters": { "level": { "type": "integer", "range": [0, 100] } }
  }
]'

post_caps plug1 '[
  {
    "name": "power",
    "description": "Turn plug on or off",
    "parameters": { "state": { "type": "string", "operations": ["on", "off"] } },
    "operations": ["on", "off"]
  }
]'

post_caps fan1 '[
  {
    "name": "power",
    "description": "Turn fan on or off",
    "parameters": { "state": { "type": "string", "operations": ["on", "off"] } },
    "operations": ["on", "off"]
  },
  {
    "name": "speed",
    "description": "Set fan speed",
    "parameters": { "level": { "type": "integer", "range": [0, 100] } }
  }
]'

post_caps speaker1 '[
  {
    "name": "volume",
    "description": "Adjust volume",
    "parameters": { "level": { "type": "integer", "range": [0, 100] } }
  },
  {
    "name": "playback",
    "description": "Play or pause music",
    "parameters": { "command": { "type": "string", "operations": ["play", "pause", "stop"] } }
  }
]'

post_caps coffee1 '[
  {
    "name": "power",
    "description": "Turn coffee maker on or off",
    "parameters": { "state": { "type": "string", "operations": ["on", "off"] } },
    "operations": ["on", "off"]
  },
  {
    "name": "brew",
    "description": "Start brewing",
    "parameters": { "cups": { "type": "integer", "range": [1, 12] } }
  }
]'

echo ""
echo "🎉 Devices and capabilities initialized!"
