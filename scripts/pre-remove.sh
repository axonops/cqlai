#!/bin/bash
# Pre-removal script for CQLAI

set -e

echo "Preparing to remove CQLAI..."

# Optional: Remind user about config files
if [ -f /etc/cqlai/cqlai.json ]; then
    echo "Note: Configuration file /etc/cqlai/cqlai.json will be preserved."
fi