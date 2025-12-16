#!/bin/bash
set -e

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "Usage: ./release.sh v1.0.0"
  exit 1
fi

echo "Building hookrunner $VERSION..."

LDFLAGS="-s -w -X github.com/ashavijit/hookrunner/internal/version.Version=${VERSION#v} -X github.com/ashavijit/hookrunner/internal/version.GitCommit=$(git rev-parse --short HEAD) -X github.com/ashavijit/hookrunner/internal/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

mkdir -p dist

platforms=(
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "windows/amd64"
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

echo "Done! Binaries in dist/"
ls -la dist/
