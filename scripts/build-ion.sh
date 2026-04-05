#!/bin/bash
# Build ION-DTN with KISS CLA from source.
# Run from the project root directory.
set -e

ION_SRC="$(pwd)/ION-DTN"
BUILD_DIR="$(pwd)/ion-build"
INSTALL_DIR="$(pwd)/ion-install"

echo "=== Building ION-DTN with KISS CLA ==="
echo "Source:  $ION_SRC"
echo "Build:   $BUILD_DIR"
echo "Install: $INSTALL_DIR"

mkdir -p "$BUILD_DIR"

# Configure
echo "--- Configuring ---"
(
  cd "$BUILD_DIR"
  CFLAGS="-g -O2 -Wno-error" "$ION_SRC/configure" \
    --prefix="$INSTALL_DIR" \
    --disable-sysctl-check
)

# Create missing directories for man pages
mkdir -p "$BUILD_DIR/ltp/kiss/doc"

# Build
echo "--- Building (this may take a few minutes) ---"
make -C "$BUILD_DIR" -j4

# Install
echo "--- Installing ---"
make -C "$BUILD_DIR" install

# Verify key binaries
echo "--- Verifying key binaries ---"
for bin in ionadmin ltpadmin bpadmin bping bpsink bpsendfile bprecvfile ltpkisscli ltpkissclo; do
  if [ -f "$INSTALL_DIR/bin/$bin" ]; then
    echo "  OK: $bin"
  else
    echo "  MISSING: $bin"
    exit 1
  fi
done

echo ""
echo "=== ION-DTN build complete ==="
echo "Binaries installed to: $INSTALL_DIR/bin/"
echo "Libraries installed to: $INSTALL_DIR/lib/"
echo ""
echo "To use ION-DTN, add to your environment:"
echo "  export PATH=\"$INSTALL_DIR/bin:\$PATH\""
echo "  export DYLD_LIBRARY_PATH=\"$INSTALL_DIR/lib:\$DYLD_LIBRARY_PATH\""
