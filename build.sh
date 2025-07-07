#!/bin/bash

APP_NAME=GoCloudflareDDNS
VERSION="$(git describe --tags --always 2>/dev/null || echo 'v0.0.1')"
BUILD_DIR=build
CHECKSUM_FILE="$BUILD_DIR/checksums.txt"

mkdir -p "$BUILD_DIR"
: >"$CHECKSUM_FILE"

platforms=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

for platform in "${platforms[@]}"; do
    IFS="/" read -r GOOS GOARCH <<<"$platform"
    output_name="${APP_NAME}-$VERSION-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        output_name+='.exe'
    fi

    echo "Building $output_name..."
    env GOOS=$GOOS GOARCH=$GOARCH go build -ldflags "-X main.version=$VERSION -s -w" -o "build/$output_name"

    archive_name=""
    if [ "$GOOS" = "windows" ]; then
        archive_name="${output_name}.zip"

        echo "Creating archive $archive_name..."
        zip -j "$BUILD_DIR/$archive_name" "$BUILD_DIR/$output_name"
    
    else
        archive_name="${output_name}.tar.gz"

        echo "Creating archive $archive_name..."
        tar -czvf "$BUILD_DIR/$archive_name" "$BUILD_DIR/$output_name"
    fi

    echo "Creating checksum for $output_name..."
    sha256sum "$BUILD_DIR/$archive_name" >> "$CHECKSUM_FILE"
    sha256sum "$BUILD_DIR/$output_name" >> "$CHECKSUM_FILE"
    echo "" >> "$CHECKSUM_FILE"

    echo "Cleaning up..."
    rm "$BUILD_DIR/$output_name"

done
