#!/bin/sh
set -e

echo "[start-all.sh] Starting Ollama in background ..."
/bin/ollama serve &

if [ -z "$USER_ENTRYPOINT_SHELL" ] && [ -z "$USER_CMD_SHELL" ]; then
  echo "[start-all.sh] No user entrypoint or cmd found; launching a shell."
  exec /bin/sh
fi

# The user’s final command is the concatenation of entrypoint + cmd
FINAL_CMD="$USER_ENTRYPOINT_SHELL $USER_CMD_SHELL"

echo "[start-all.sh] Execing: $FINAL_CMD"
set -- $FINAL_CMD
exec "$@"
