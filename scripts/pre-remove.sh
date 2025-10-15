#!/bin/bash
# Pre-removal script for CQLAI

set -e

echo "Preparing to remove CQLAI..."

# Remind user about config files
if [ -f ~/.cqlai.json ]; then
    echo "Note: User configuration file ~/.cqlai.json will be preserved."
fi