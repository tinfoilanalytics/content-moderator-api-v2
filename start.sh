#!/bin/sh

# Start Ollama server in background
ollama serve &

# Wait for Ollama server to initialize
sleep 5

# Pull the required model
ollama pull llama-guard3:1b

# Execute the content moderation service
exec /contentmod