#!/bin/bash

API=http://localhost:8080/llm
HEADER="Content-Type: application/json"

echo "🔎 Asking: 'List all devices'"
curl -s -X POST $API -H "$HEADER" -d '{"prompt": "List all devices"}' | jq
echo ""

echo "🔍 Asking: 'Is there a fan in the bedroom?'"
curl -s -X POST $API -H "$HEADER" -d '{"prompt": "Is there a fan in the bedroom?"}' | jq
echo ""

echo "📡 Asking: 'Scan for new devices'"
curl -s -X POST $API -H "$HEADER" -d '{"prompt": "Scan for new devices"}' | jq
echo ""

echo "✏️ Asking: 'Rename device fan1 to ceiling fan'"
curl -s -X POST $API -H "$HEADER" -d '{"prompt": "Rename fan1 to ceiling fan"}' | jq
echo ""

echo "💨 Asking: 'Set ceiling fan speed to 75%'"
curl -s -X POST $API -H "$HEADER" -d '{"prompt": "Set ceiling fan speed to 75"}' | jq
echo ""
