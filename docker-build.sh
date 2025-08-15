#!/bin/bash

# Docker build script with ARM64/Apple Silicon support
# Usage: ./docker-build.sh [options]

set -e

# Default values
IMAGE_NAME="riskmatrix"
TAG="latest"
PLATFORM=""
PUSH=false
BUILDX_SETUP=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Help function
show_help() {
    echo "Docker build script for RiskMatrix with ARM64/Apple Silicon support"
    echo ""
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  -t, --tag TAG         Set image tag (default: latest)"
    echo "  -p, --platform PLAT   Specify platform (auto-detect if not set)"
    echo "                        Options: linux/amd64, linux/arm64, multi"
    echo "  --push               Push image after build (requires registry)"
    echo "  --setup-buildx       Set up Docker buildx for multi-platform builds"
    echo "  -h, --help           Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                   # Auto-detect platform and build"
    echo "  $0 -p linux/arm64    # Build for ARM64"
    echo "  $0 -p multi --push   # Build multi-platform and push"
    echo "  $0 --setup-buildx    # Set up buildx for multi-platform builds"
}

# Detect current platform
detect_platform() {
    local arch=$(uname -m)
    local os=$(uname -s)
    
    case "$arch" in
        x86_64)
            echo "linux/amd64"
            ;;
        arm64|aarch64)
            echo "linux/arm64"
            ;;
        *)
            echo "linux/amd64"  # Default fallback
            ;;
    esac
}

# Set up Docker buildx
setup_buildx() {
    echo -e "${BLUE}Setting up Docker buildx for multi-platform builds...${NC}"
    
    # Check if buildx is available
    if ! docker buildx version >/dev/null 2>&1; then
        echo -e "${RED}Error: Docker buildx is not available. Please update Docker.${NC}"
        exit 1
    fi
    
    # Create and use a new builder instance
    if docker buildx ls | grep -q "multiplatform"; then
        echo -e "${YELLOW}Buildx instance 'multiplatform' already exists.${NC}"
        docker buildx use multiplatform
    else
        echo -e "${GREEN}Creating new buildx instance 'multiplatform'...${NC}"
        docker buildx create --name multiplatform --use --platform linux/amd64,linux/arm64
    fi
    
    # Bootstrap the builder
    echo -e "${BLUE}Bootstrapping builder...${NC}"
    docker buildx inspect --bootstrap
    
    echo -e "${GREEN}Buildx setup complete!${NC}"
}

# Build function
build_image() {
    local platform=$1
    local tag=$2
    local push_flag=$3
    
    echo -e "${BLUE}Building RiskMatrix Docker image...${NC}"
    echo -e "${BLUE}Platform: ${platform}${NC}"
    echo -e "${BLUE}Tag: ${IMAGE_NAME}:${tag}${NC}"
    
    case "$platform" in
        "multi")
            echo -e "${YELLOW}Building multi-platform image (AMD64 + ARM64)...${NC}"
            if [ "$push_flag" = true ]; then
                docker buildx build \
                    --platform linux/amd64,linux/arm64 \
                    -t "${IMAGE_NAME}:${tag}" \
                    --push .
            else
                docker buildx build \
                    --platform linux/amd64,linux/arm64 \
                    -t "${IMAGE_NAME}:${tag}" \
                    --load .
            fi
            ;;
        *)
            echo -e "${YELLOW}Building for platform: ${platform}${NC}"
            if command -v docker buildx >/dev/null 2>&1; then
                # Use buildx for better cross-platform support
                docker buildx build \
                    --platform "$platform" \
                    -t "${IMAGE_NAME}:${tag}" \
                    --load .
            else
                # Fallback to regular docker build
                docker build \
                    -t "${IMAGE_NAME}:${tag}" .
            fi
            
            if [ "$push_flag" = true ]; then
                echo -e "${BLUE}Pushing image...${NC}"
                docker push "${IMAGE_NAME}:${tag}"
            fi
            ;;
    esac
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--tag)
            TAG="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        --setup-buildx)
            BUILDX_SETUP=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
echo -e "${GREEN}RiskMatrix Docker Build Script${NC}"
echo -e "${GREEN}=============================${NC}"

# Set up buildx if requested
if [ "$BUILDX_SETUP" = true ]; then
    setup_buildx
    exit 0
fi

# Auto-detect platform if not specified
if [ -z "$PLATFORM" ]; then
    PLATFORM=$(detect_platform)
    echo -e "${YELLOW}Auto-detected platform: ${PLATFORM}${NC}"
fi

# Check if multi-platform build requires buildx setup
if [ "$PLATFORM" = "multi" ]; then
    if ! docker buildx ls | grep -q "multiplatform"; then
        echo -e "${YELLOW}Multi-platform build requires buildx setup.${NC}"
        echo -e "${BLUE}Run: $0 --setup-buildx${NC}"
        echo -e "${YELLOW}Setting up buildx automatically...${NC}"
        setup_buildx
    else
        docker buildx use multiplatform
    fi
fi

# Validate Docker is running
if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running or not accessible.${NC}"
    exit 1
fi

# Check if Dockerfile exists
if [ ! -f "Dockerfile" ]; then
    echo -e "${RED}Error: Dockerfile not found in current directory.${NC}"
    exit 1
fi

# Build the image
build_image "$PLATFORM" "$TAG" "$PUSH"

echo -e "${GREEN}Build completed successfully!${NC}"
echo -e "${BLUE}Image: ${IMAGE_NAME}:${TAG}${NC}"

# Show next steps
echo -e "${YELLOW}Next steps:${NC}"
if [ "$PUSH" != true ]; then
    echo -e "  ${BLUE}Run the container:${NC} docker run -p 8080:8080 ${IMAGE_NAME}:${TAG}"
    echo -e "  ${BLUE}Or use compose:${NC} docker-compose up -d"
fi
echo -e "  ${BLUE}View logs:${NC} docker logs riskmatrix"
echo -e "  ${BLUE}Access app:${NC} http://localhost:8080"
echo -e ""
echo -e "${GREEN}Features included:${NC}"
echo -e "  ✅ Automatic MITRE ATT&CK data import on first run"
echo -e "  ✅ Multi-platform support (AMD64/ARM64)"
echo -e "  ✅ Non-root user for security"
echo -e "  ✅ Health checks enabled"