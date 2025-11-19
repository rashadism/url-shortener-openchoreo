#!/bin/bash

# Build and push Docker images to Docker Hub
# Usage: ./build-and-push.sh [DOCKER_USERNAME] [--no-cache]
# Examples:
#   ./build-and-push.sh                    # Use default username (rashadxyz)
#   ./build-and-push.sh myusername         # Use custom username
#   ./build-and-push.sh --no-cache         # Use default username with --no-cache
#   ./build-and-push.sh myusername --no-cache  # Use custom username with --no-cache

set -e

# Configuration
DOCKER_USERNAME="rashadxyz"  # Default username
VERSION="${VERSION:-demo}"
NO_CACHE=""

# Parse arguments
for arg in "$@"; do
    if [[ "$arg" == "--no-cache" ]]; then
        NO_CACHE="--no-cache"
    elif [[ "$arg" != --* ]]; then
        # First non-flag argument is the Docker username
        DOCKER_USERNAME="$arg"
    fi
done

echo "======================================"
echo "Building and pushing URL Shortener images"
echo "Docker Hub Username: $DOCKER_USERNAME"
echo "Version: $VERSION"
echo "======================================"

# Function to build and push an image
build_and_push() {
    local service_name=$1
    local service_dir=$2
    local image_name="${DOCKER_USERNAME}/url-shortener-${service_name}:${VERSION}"

    echo ""
    echo "Building $service_name..."
    echo "Image: $image_name"

    docker build $NO_CACHE -t "$image_name" -f "${service_dir}/Dockerfile" "${service_dir}"

    echo "Pushing $image_name to Docker Hub..."
    docker push "$image_name"

    echo "✓ Successfully built and pushed $image_name"
}

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if logged into Docker Hub
if ! docker info | grep -q "Username: $DOCKER_USERNAME"; then
    echo "Warning: You may not be logged into Docker Hub as $DOCKER_USERNAME"
    echo "Please run: docker login"
    read -p "Continue anyway? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Build and push all services
build_and_push "api" "api-service"
build_and_push "analytics" "analytics-service"
build_and_push "frontend" "frontend"

echo ""
echo "======================================"
echo "✓ All images built and pushed successfully!"
echo "======================================"
echo ""
echo "Images pushed:"
echo "  - ${DOCKER_USERNAME}/url-shortener-api:${VERSION}"
echo "  - ${DOCKER_USERNAME}/url-shortener-analytics:${VERSION}"
echo "  - ${DOCKER_USERNAME}/url-shortener-frontend:${VERSION}"
echo ""
echo "Usage examples:"
echo "  ./build-and-push.sh myusername              # Use custom Docker Hub username"
echo "  VERSION=v1.0.0 ./build-and-push.sh          # Use specific version tag"
echo "  ./build-and-push.sh myusername --no-cache   # Custom username with clean build"
