#!/bin/bash
# Post-installation script for CQLAI

set -e

# Create config directory if it doesn't exist
mkdir -p /etc/cqlai

# Copy example config if no config exists
if [ ! -f /etc/cqlai/cqlai.json ]; then
    if [ -f /etc/cqlai/cqlai.json.example ]; then
        cp /etc/cqlai/cqlai.json.example /etc/cqlai/cqlai.json
        chmod 644 /etc/cqlai/cqlai.json
    fi
fi

# Ensure binary is executable
chmod 755 /usr/bin/cqlai

echo "CQLAI has been successfully installed."
echo ""
echo "To get started:"
echo "  1. Configure your connection: edit /etc/cqlai/cqlai.json or ~/.cqlai.json"
echo "  2. Run: cqlai"
echo ""
echo "For more information, visit: https://github.com/axonops/cqlai"