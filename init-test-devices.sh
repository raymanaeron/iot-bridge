#!/bin/bash

baseUrl="http://localhost:8080"

echo "🧹 Deleting all existing devices..."
existing=$(curl -s "$baseUrl/devices" | jq -r '.[].id')
for id in $existing; do
  echo "❌ Deleting $id"
  curl -s -X DELETE "$baseUrl/devices/$id"
done

echo "✅ Clean slate ready."

echo ""
echo "➕ Adding test devices..."

curl -s -X POST "$baseUrl/devices" -H "Content-Type: application/json" -d '{
  "id": "bulb1",
  "name": "Living Room Bulb",
  "type": "bulb",
  "protocol": "zigbee",
  "room": "Living Room",
  "state": {}
}' && echo "💡 bulb1 added"

curl -s -X POST "$baseUrl/devices" -H "Content-Type: application/json" -d '{
  "id": "plug1",
  "name": "Coffee Machine Plug",
  "type": "smart_plug",
  "protocol": "zwave",
  "room": "Kitchen",
  "state": {}
}' && echo "🔌 plug1 added"

curl -s -X POST "$baseUrl/devices" -H "Content-Type: application/json" -d '{
  "id": "fan1",
  "name": "Ceiling Fan",
  "type": "fan",
  "protocol": "zigbee",
  "room": "Bedroom",
  "state": {}
}' && echo "🌪️ fan1 added"

curl -s -X POST "$baseUrl/devices" -H "Content-Type: application/json" -d '{
  "id": "speaker1",
  "name": "Kitchen Speaker",
  "type": "speaker",
  "protocol": "wifi",
  "room": "Kitchen",
  "state": {}
}' && echo "🔊 speaker1 added"

curl -s -X POST "$baseUrl/devices" -H "Content-Type: application/json" -d '{
  "id": "coffee1",
  "name": "Coffee Maker",
  "type": "smart_appliance",
  "protocol": "zwave",
  "room": "Kitchen",
  "state": {}
}' && echo "☕ coffee1 added"

echo ""
echo "🧠 Adding device capabilities..."

post_caps() {
  device_id=$1
  payload=$2
  curl -s -X POST "$baseUrl/devices/$device_id/capabilities" \
    -H "Content-Type: application/json" -d "$payload" \
    && echo "✅ Capabilities set for $device_id"
}

# Capabilities for each device type
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
