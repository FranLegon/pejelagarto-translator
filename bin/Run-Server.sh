#!/bin/bash
# Run-Server.sh
# Launch script for pejelagarto-translator server on Linux/macOS

# Determine the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Determine the binary name
if [ -f "${SCRIPT_DIR}/pejelagarto-translator" ]; then
    BINARY="${SCRIPT_DIR}/pejelagarto-translator"
elif [ -f "${SCRIPT_DIR}/piper-server" ]; then
    BINARY="${SCRIPT_DIR}/piper-server"
else
    echo "Error: Binary not found in ${SCRIPT_DIR}"
    echo "Expected: pejelagarto-translator or piper-server"
    exit 1
fi

# Make sure binary is executable
chmod +x "$BINARY"

# Run the server with ngrok credentials
"$BINARY" -ngrok_token '34QfuhfXXNQmIe0TbFH67RmNZZZ_7TtoYMAdwwgdYV1JFE1z6' -ngrok_domain 'emptiest-unwieldily-kiana.ngrok-free.dev'
