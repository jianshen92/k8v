#!/bin/bash
# K8V Build Script - Optimized Production Build

set -e

VERSION=${VERSION:-"dev"}
OUTPUT=${OUTPUT:-"k8v"}

echo "🔨 Building k8v (optimized)..."
echo "Version: $VERSION"
echo "Output: $OUTPUT"
echo ""

# Build with optimizations:
# -ldflags="-s -w" removes debug symbols and DWARF information
#   -s: Omit the symbol table and debug information
#   -w: Omit the DWARF symbol table
# This reduces binary size by ~30% (63MB -> 44MB)

go build \
  -ldflags="-s -w -X main.Version=$VERSION" \
  -o "$OUTPUT" \
  cmd/k8v/main.go

SIZE=$(du -h "$OUTPUT" | cut -f1)
echo "✅ Build complete!"
echo "📦 Binary size: $SIZE"
echo ""
echo "Optional: Further compress with UPX (requires installation):"
echo "  sudo apt-get install upx-ucl   # Install UPX"
echo "  upx --best --lzma $OUTPUT      # Compress (can reduce by ~50%)"
echo ""
echo "Run with: ./$OUTPUT"
