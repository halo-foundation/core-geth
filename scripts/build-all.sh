#!/bin/bash
# Build Halo Chain geth for all platforms

set -e

echo "🔨 Building Halo Chain for All Platforms"
echo "=========================================="
echo ""

# Check if Go is available
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed or not in PATH"
    echo "   Please install Go 1.21+ from https://go.dev/dl/"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}')
echo "✅ Go version: $GO_VERSION"
echo ""

# Create output directory
OUTPUT_DIR="release"
mkdir -p "$OUTPUT_DIR"

# Build Linux
echo "📦 Building for Linux (amd64)..."
make clean
make geth
cp build/bin/geth "$OUTPUT_DIR/geth-linux-amd64"
echo "✅ Linux build complete: $OUTPUT_DIR/geth-linux-amd64"
echo ""

# Build Windows
echo "📦 Building for Windows (amd64)..."
if ! command -v x86_64-w64-mingw32-gcc &> /dev/null; then
    echo "⚠️  MinGW not found, skipping Windows build"
    echo "   Install with: sudo apt install gcc-mingw-w64-x86-64"
else
    make clean
    GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc make geth
    cp build/bin/geth.exe "$OUTPUT_DIR/geth-windows-amd64.exe"
    echo "✅ Windows build complete: $OUTPUT_DIR/geth-windows-amd64.exe"
fi
echo ""

# Build macOS Intel
echo "📦 Building for macOS (amd64)..."
make clean
GOOS=darwin GOARCH=amd64 make geth
cp build/bin/geth "$OUTPUT_DIR/geth-darwin-amd64"
echo "✅ macOS Intel build complete: $OUTPUT_DIR/geth-darwin-amd64"
echo ""

# Build macOS ARM64
echo "📦 Building for macOS (arm64)..."
make clean
GOOS=darwin GOARCH=arm64 make geth
cp build/bin/geth "$OUTPUT_DIR/geth-darwin-arm64"
echo "✅ macOS ARM64 build complete: $OUTPUT_DIR/geth-darwin-arm64"
echo ""

echo "✅ All builds complete!"
echo ""
echo "📦 Binaries in $OUTPUT_DIR/:"
ls -lh "$OUTPUT_DIR"/geth-*
echo ""

# Create checksums
echo "🔐 Creating checksums..."
cd "$OUTPUT_DIR"
sha256sum geth-* > SHA256SUMS
cd ..
echo "✅ Checksums created: $OUTPUT_DIR/SHA256SUMS"
echo ""

cat "$OUTPUT_DIR/SHA256SUMS"
