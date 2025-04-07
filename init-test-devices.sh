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
echo "🎉 Test devices initialized!"
