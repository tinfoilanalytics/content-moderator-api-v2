#!/usr/bin/env bash
set -e

echo "=== Starting wrapper build process ==="
echo "Building user image..."
docker build -f Dockerfile.user -t userimage .
echo "✓ User image built successfully"

echo "Inspecting user image configuration..."
USER_ENTRYPOINT_JSON=$(docker inspect userimage --format='{{json .Config.Entrypoint}}')
USER_CMD_JSON=$(docker inspect userimage --format='{{json .Config.Cmd}}')

echo "Raw Entrypoint JSON: $USER_ENTRYPOINT_JSON"
echo "Raw Cmd JSON: $USER_CMD_JSON"

# We can parse them now with jq (assuming you have jq installed locally)
USER_ENTRYPOINT_SHELL=$(echo "$USER_ENTRYPOINT_JSON" | jq -r '.[]' 2>/dev/null | xargs)
USER_CMD_SHELL=$(echo "$USER_CMD_JSON" | jq -r '.[]' 2>/dev/null | xargs)

echo "=== Parsed configuration ==="
echo "Entrypoint shell form: $USER_ENTRYPOINT_SHELL"
echo "Cmd shell form:        $USER_CMD_SHELL"

echo "=== Building final wrapper image ==="
echo "Build args:"
echo "  USER_ENTRYPOINT_SHELL=$USER_ENTRYPOINT_SHELL"
echo "  USER_CMD_SHELL=$USER_CMD_SHELL"

docker build \
  --build-arg USER_ENTRYPOINT_SHELL="$USER_ENTRYPOINT_SHELL" \
  --build-arg USER_CMD_SHELL="$USER_CMD_SHELL" \
  -f Dockerfile.wrapper \
  -t finalimage \
  .

echo "✓ Final wrapper image built successfully"
echo "=== Build process complete ==="

