# Install Ollama
> On RP 4/5 (64 bit only)

curl -fsSL https://ollama.com/install.sh | sh

> Then start the Ollama service

ollama serve

> Pull a tiny model like phi

ollama pull phi

> or

ollama pull tinyllama

> To upload .env into the PI from Windows

scp .env rayman@raspberrypi.local:/home/rayman/iot-bridge/

> To run the LLM Test Script - you need these

sudo apt update
sudo apt install jq -y
