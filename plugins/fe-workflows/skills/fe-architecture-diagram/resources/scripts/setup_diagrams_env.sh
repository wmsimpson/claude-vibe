#!/bin/bash
# Setup script for fe-architecture-diagram skill
# Creates isolated Python environment with diagrams package in ~/.vibe/diagrams/

set -e

VIBE_DIR="$HOME/.vibe"
DIAGRAMS_DIR="$VIBE_DIR/diagrams"
VENV_DIR="$DIAGRAMS_DIR/.venv"

echo "=== Architecture Diagram Environment Setup ==="
echo ""

# Check for Homebrew
if ! command -v brew &> /dev/null; then
    echo "ERROR: Homebrew not found. Please install Homebrew first:"
    echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
    exit 1
fi

# Check for uv
if ! command -v uv &> /dev/null; then
    echo "ERROR: uv not found. Please run /configure-vibe first to install uv."
    exit 1
fi

# Install Graphviz if not present
echo "Checking Graphviz..."
if ! command -v dot &> /dev/null; then
    echo "Installing Graphviz via Homebrew..."
    brew install graphviz
else
    echo "Graphviz already installed: $(dot -V 2>&1 | head -1)"
fi

# Create directories
echo ""
echo "Creating directories..."
mkdir -p "$DIAGRAMS_DIR"

# Create virtual environment with uv
echo ""
echo "Creating Python virtual environment in $VENV_DIR..."
if [ -d "$VENV_DIR" ]; then
    echo "Virtual environment already exists, updating..."
else
    cd "$DIAGRAMS_DIR"
    uv venv
fi

# Install diagrams package
echo ""
echo "Installing diagrams package..."
cd "$DIAGRAMS_DIR"
uv pip install diagrams

# Verify installation
echo ""
echo "Verifying installation..."
"$VENV_DIR/bin/python" -c "import diagrams; print(f'diagrams version: {diagrams.__version__}')"
"$VENV_DIR/bin/python" -c "from diagrams import Diagram; print('Diagram class imported successfully')"

# Check for optional Mermaid CLI
echo ""
echo "Checking optional Mermaid CLI..."
if command -v mmdc &> /dev/null; then
    echo "Mermaid CLI already installed: $(mmdc --version 2>&1 | head -1)"
else
    echo "Mermaid CLI not found (optional). Install with:"
    echo "  npm install -g @mermaid-js/mermaid-cli"
fi

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Usage:"
echo "  # Run a diagram script"
echo "  $VENV_DIR/bin/python your_diagram.py"
echo ""
echo "  # Or activate the environment"
echo "  source $VENV_DIR/bin/activate"
echo "  python your_diagram.py"
echo ""
