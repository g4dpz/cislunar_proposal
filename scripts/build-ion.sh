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

# Clean any stale config in the source tree
if [ -f "$ION_SRC/Makefile" ]; then
  echo "--- Cleaning stale ION-DTN config ---"
  make -C "$ION_SRC" distclean 2>/dev/null || true
fi

# Generate configure script if missing
if [ ! -f "$ION_SRC/configure" ]; then
  echo "--- Generating configure script (autoreconf) ---"
  (cd "$ION_SRC" && autoreconf -fi)
fi

# Configure
echo "--- Configuring ---"
(
  cd "$BUILD_DIR"
  CFLAGS="-g -O2 -Wno-error" "$ION_SRC/configure" \
    --prefix="$INSTALL_DIR" \
    --disable-sysctl-check \
    --enable-manpages=no
)

# Create missing directories for man pages
mkdir -p "$BUILD_DIR/ltp/kiss/doc"
mkdir -p "$BUILD_DIR/ltp/doc"
mkdir -p "$BUILD_DIR/ici/doc"
mkdir -p "$BUILD_DIR/dgr/doc"
mkdir -p "$BUILD_DIR/bpv7/doc"
mkdir -p "$BUILD_DIR/tc/doc"
mkdir -p "$BUILD_DIR/bss/doc"
mkdir -p "$BUILD_DIR/dtpc/doc"
mkdir -p "$BUILD_DIR/bssp/doc"
mkdir -p "$BUILD_DIR/ams/doc"
mkdir -p "$BUILD_DIR/cfdp/doc"
mkdir -p "$BUILD_DIR/nm/doc"
mkdir -p "$BUILD_DIR/restart/doc"

# Build
echo "--- Building (this may take a few minutes) ---"
make -C "$BUILD_DIR" -j4

# Install
echo "--- Installing ---"
make -C "$BUILD_DIR" install

# Verify key binaries
echo "--- Verifying key binaries ---"
for bin in ionadmin ltpadmin bpadmin bping bpsink bpsendfile bprecvfile ltpkisslsi ltpkisslso; do
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
