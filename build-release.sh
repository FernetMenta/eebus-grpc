#!/usr/bin/env bash

set -e

IMAGE_NAME="fernetmenta/eebus-grpc"

# Default settings: Build everything, but DO NOT upload by default
BUILD_BINARIES=true
BUILD_DOCKER=true
UPLOAD_GITHUB=false

# Help text
show_help() {
    echo "Usage: $0 [OPTIONS] [TAG]"
    echo ""
    echo "Options:"
    echo "  -b, --binaries    Build binaries only"
    echo "  -d, --docker      Build Docker container only"
    echo "  -u, --upload      Enable uploading binaries to GitHub (disabled by default)"
    echo "  -h, --help        Show this help text"
    echo ""
    echo "Examples:"
    echo "  $0 v2.0.1                # Build everything (Binaries + Docker), no upload"
    echo "  $0 -u v2.0.1             # Build everything and upload binaries to GitHub"
    echo "  $0 --binaries -u v2.0.1  # Build & upload binaries only"
    echo "  $0 -d v2.0.1             # Build Docker image only"
    exit 0
}

# Parameter parsing
POSITIONAL_ARGS=()

while [[ $# -gt 0 ]]; do
  case $1 in
    -b|--binaries)
      BUILD_BINARIES=true
      BUILD_DOCKER=false
      shift
      ;;
    -d|--docker)
      BUILD_BINARIES=false
      BUILD_DOCKER=true
      shift
      ;;
    -u|--upload)
      UPLOAD_GITHUB=true
      shift
      ;;
    --no-upload)
      UPLOAD_GITHUB=false
      shift
      ;;
    -h|--help)
      show_help
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # Capture the tag
      shift
      ;;
  esac
done

# Reset positional arguments for the tag
set -- "${POSITIONAL_ARGS[@]}"
TAG=$1

# 1. Determine Tag
if [ -z "$TAG" ]; then
    GIT_TAG=$(git describe --tags --abbrev=0 2>/dev/null || true)
    
    if [ -n "$GIT_TAG" ]; then
        read -p "No tag specified. Use Git tag '$GIT_TAG'? (y/N): " confirm
        if [[ "$confirm" =~ ^[Nn]$ ]]; then
            read -p "Please enter tag (e.g., v2.0.1): " TAG
        else
            TAG=$GIT_TAG
        fi
    else
        read -p "Please enter tag (e.g., v2.0.1): " TAG
    fi
fi

if [ -z "$TAG" ]; then
    echo "Error: No tag specified!"
    exit 1
fi

echo "=========================================="
echo " Starting release process for version: $TAG"
echo "=========================================="

# 2. Build Binaries
if [ "$BUILD_BINARIES" = true ]; then
    echo ""
    echo "--> Compiling Go binaries..."
    mkdir -p dist
    rm -f dist/*

    echo "   [1/3] Windows amd64..."
    GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$TAG" -o "dist/eebus-grpc-windows-amd64.exe" ./cmd/main.go

    echo "   [2/3] Linux amd64..."
    GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$TAG" -o "dist/eebus-grpc-linux-amd64" ./cmd/main.go

    echo "   [3/3] Linux arm64..."
    GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$TAG" -o "dist/eebus-grpc-linux-arm64" ./cmd/main.go

    echo "Binaries successfully placed in /dist."

    # GitHub Upload
    if [ "$UPLOAD_GITHUB" = true ]; then
        echo ""
        echo "--> Uploading binaries to GitHub..."

        if command -v gh &> /dev/null; then
            if gh release view "$TAG" &>/dev/null; then
                echo "Release $TAG already exists. Overwriting files..."
                gh release upload "$TAG" dist/* --clobber
            else
                echo "Creating new GitHub Release $TAG..."
                gh release create "$TAG" dist/* --title "Release $TAG" --generate-notes
            fi
            echo "GitHub upload completed!"
        else
            echo "WARNING: 'gh' CLI not found. Skipping upload."
        fi
    else
        echo "--> Skipping GitHub upload (disabled by default, pass -u or --upload to enable)."
    fi
fi

# 3. Docker Build
if [ "$BUILD_DOCKER" = true ]; then
    echo ""
    echo "--> Building Docker Multi-Arch Container..."

    docker buildx inspect >/dev/null 2>&1 || docker buildx create --use

    docker buildx build \
      --platform linux/amd64,linux/arm64 \
      -t "${IMAGE_NAME}:${TAG}" \
      -t "${IMAGE_NAME}:latest" \
      .
fi

echo ""
echo "=========================================="
echo " Done!"
echo "=========================================="