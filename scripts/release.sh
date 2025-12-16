#!/bin/bash
set -e

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh v1.0.0"
  exit 1
fi

echo "Building hookrunner $VERSION for all platforms..."

LDFLAGS="-s -w -X github.com/ashavijit/hookrunner/internal/version.Version=${VERSION#v} -X github.com/ashavijit/hookrunner/internal/version.GitCommit=$(git rev-parse --short HEAD) -X github.com/ashavijit/hookrunner/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

mkdir -p dist

platforms=(
  "linux/amd64"
  "linux/arm64"
  "linux/386"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
  "windows/386"
  "windows/arm64"
  "freebsd/amd64"
  "freebsd/arm64"
  "openbsd/amd64"
  "netbsd/amd64"
)

for platform in "${platforms[@]}"; do
  IFS='/' read -r GOOS GOARCH <<< "$platform"

  output="dist/hookrunner-${GOOS}-${GOARCH}"
  if [ "$GOOS" = "windows" ]; then
    output="${output}.exe"
  fi

  echo "Building $output..."
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "$LDFLAGS" -o "$output" ./cmd/hookrunner
done

echo "Creating checksums..."
cd dist
sha256sum hookrunner-* > checksums.txt
cd ..

echo ""
echo "Done! Built for:"
echo "  - Linux (amd64, arm64, 386)"
echo "  - macOS (amd64, arm64)"
echo "  - Windows (amd64, 386, arm64)"
echo "  - FreeBSD (amd64, arm64)"
echo "  - OpenBSD (amd64)"
echo "  - NetBSD (amd64)"
echo ""
ls -la dist/
