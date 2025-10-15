#!/bin/bash
# Post-installation script for CQLAI

set -e

# Ensure binary is executable
chmod 755 /usr/bin/cqlai

echo "CQLAI has been successfully installed."
echo ""
echo "To get started:"
echo "  1. Copy example config: cp /usr/share/doc/cqlai/cqlai.json.example ~/.cqlai.json"
echo "  2. Edit your configuration: nano ~/.cqlai.json"
echo "  3. Run: cqlai"
echo ""
echo "Configuration can also be placed in:"
echo "  - ./cqlai.json (current directory)"
echo "  - ~/.cqlai.json (user home)"
echo "  - ~/.config/cqlai/config.json (XDG config)"
echo ""
echo "For more information, visit: https://github.com/axonops/cqlai"