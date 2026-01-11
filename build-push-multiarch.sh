#!/bin/bash
set -e

echo "=========================================="
echo "Building & Publishing Multi-Arch Image"
echo "=========================================="
echo "Image: rod16/whatsapp-multi-session:latest"
echo "Platforms: linux/amd64, linux/arm64"
echo ""

# Step 1: Initialize buildx
echo "Step 1: Initializing Docker buildx..."
docker buildx create --name multiarch --use 2>/dev/null || true
docker buildx inspect --bootstrap
echo "✓ Buildx initialized"
echo ""

# Step 2: Build and push multi-arch image
echo "Step 2: Building multi-arch image (this will take a while)..."
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t rod16/whatsapp-multi-session:latest \
  --push \
  .
echo ""
echo "✓ Multi-arch image built and pushed!"
echo ""
echo "=========================================="
echo "Summary"
echo "=========================================="
echo "Image: rod16/whatsapp-multi-session:latest"
echo "Platforms: linux/amd64, linux/arm64"
echo "Docker Hub: https://hub.docker.com/r/rod16/whatsapp-multi-session"
echo ""
echo "To pull and run:"
echo "  docker pull rod16/whatsapp-multi-session:latest"
echo "  docker run -p 8080:8080 rod16/whatsapp-multi-session:latest"
echo ""
