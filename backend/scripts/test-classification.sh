#!/bin/bash

# Script to run the interactive classification test tool

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Check if Python 3 is available
if ! command -v python3 &> /dev/null; then
    if ! command -v python &> /dev/null; then
        echo "❌ Error: Python 3 is not installed"
        echo "Please install Python 3 to use this script"
        exit 1
    else
        PYTHON_CMD="python"
    fi
else
    PYTHON_CMD="python3"
fi

# Check if requests library is installed
if ! $PYTHON_CMD -c "import requests" 2>/dev/null; then
    echo "⚠️  Warning: 'requests' library is not installed"
    echo "Installing requests..."
    $PYTHON_CMD -m pip install requests --quiet
    if [ $? -ne 0 ]; then
        echo "❌ Error: Failed to install 'requests' library"
        echo "Please install it manually: pip install requests"
        exit 1
    fi
fi

# Run the Python script
$PYTHON_CMD "$SCRIPT_DIR/test-classification.py"

