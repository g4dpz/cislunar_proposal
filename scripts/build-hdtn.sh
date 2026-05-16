#!/bin/bash
# Build HDTN from source with KISS CLA plugin
set -e

HDTN_SRC="${HDTN_SRC:-$HOME/HDTN}"
KISS_CLA_SRC="$(cd "$(dirname "$0")/.." && pwd)/plugins/kiss-cla"
BUILD_DIR="${HDTN_SRC}/build"
INSTALL_PREFIX="/usr/local"

echo "=== Building HDTN with KISS CLA Plugin ==="
echo "HDTN source: $HDTN_SRC"
echo "KISS CLA source: $KISS_CLA_SRC"

# Clone HDTN if not present
if [ ! -d "$HDTN_SRC" ]; then
    echo "Cloning HDTN..."
    git clone https://github.com/nasa/HDTN.git "$HDTN_SRC"
fi

# Copy KISS CLA plugin into HDTN source tree
echo "Installing KISS CLA plugin..."
mkdir -p "$HDTN_SRC/module/cla_kiss"
cp "$KISS_CLA_SRC/kiss_cla_plugin.h" "$HDTN_SRC/module/cla_kiss/"
cp "$KISS_CLA_SRC/kiss_cla_plugin.cpp" "$HDTN_SRC/module/cla_kiss/"
cp "$KISS_CLA_SRC/CMakeLists.txt" "$HDTN_SRC/module/cla_kiss/"

# Build HDTN
echo "Building HDTN..."
mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"
cmake .. -DCMAKE_INSTALL_PREFIX="$INSTALL_PREFIX" -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)

echo ""
echo "=== Build Complete ==="
echo "To install: cd $BUILD_DIR && sudo make install"
echo "Binary will be at: $INSTALL_PREFIX/bin/hdtn-one-process"
