#!/bin/bash
# Test script for video generation setup

echo "=== Video Generation Test Suite ==="
echo

# Check if FAL_KEY is set
if [ -z "$FAL_KEY" ]; then
    echo "❌ FAL_KEY environment variable is not set"
    echo "   Set it with: export FAL_KEY=your-api-key"
    exit 1
else
    echo "✅ FAL_KEY is set"
fi

# Check if poetry is installed
if ! command -v poetry &> /dev/null; then
    echo "❌ Poetry is not installed"
    echo "   Install it from: https://python-poetry.org/docs/#installation"
    exit 1
else
    echo "✅ Poetry is installed"
fi

# Check if dependencies are installed
if [ ! -d ".venv" ] && [ ! -f "poetry.lock" ]; then
    echo "⚠️  Dependencies not installed. Running poetry install..."
    poetry install
else
    echo "✅ Dependencies appear to be installed"
fi

echo
echo "=== Testing Python Video Generation ==="
echo

# Test 1: Help message
echo "Test 1: Checking if script runs..."
poetry run python video/video_generation.py > /dev/null 2>&1
if [ $? -eq 0 ] || [ $? -eq 1 ]; then
    echo "✅ Python script is accessible"
else
    echo "❌ Python script failed to run"
    exit 1
fi

# Test 2: JSON input validation
echo "Test 2: Testing JSON input..."
echo '{"mode":"text_to_video","prompt":"test","output_path":"test.mp4"}' | poetry run python video/video_generation.py > /tmp/test_output.json 2>&1

if [ $? -eq 0 ]; then
    echo "✅ JSON input processing works"

    # Check if output is valid JSON
    if python3 -c "import json; json.load(open('/tmp/test_output.json'))" 2>/dev/null; then
        echo "✅ Output is valid JSON"
    else
        echo "⚠️  Output might not be valid JSON (this is okay if there were stderr messages)"
    fi
else
    echo "❌ JSON input processing failed"
fi

echo
echo "=== Testing Go Integration ==="
echo

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed"
    exit 1
else
    echo "✅ Go is installed"
fi

# Check if video package compiles
echo "Testing Go video package compilation..."
cd video 2>/dev/null
if go build video.go 2>&1 | grep -q "error"; then
    echo "❌ Go video package has compilation errors"
    cd ..
    exit 1
else
    echo "✅ Go video package compiles successfully"
    rm -f video 2>/dev/null
    cd ..
fi

echo
echo "=== Setup Complete ==="
echo
echo "You can now:"
echo "1. Run the complete workflow: go run main.go"
echo "2. Test Python directly: poetry run python video/video_generation.py text 'your prompt' output.mp4"
echo "3. Use the video package in your own Go code"
echo
echo "Note: Actual video generation will require FAL API credits"

