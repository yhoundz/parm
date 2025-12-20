#!/usr/bin/env bash
set -euo pipefail

# Manual release script for parm
# Usage: ./scripts/release.sh <version>
# Example: ./scripts/release.sh v0.1.0

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$ROOT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1" >&2; exit 1; }

# Check dependencies
command -v go >/dev/null 2>&1 || error "go is required"
command -v git >/dev/null 2>&1 || error "git is required"
command -v gh >/dev/null 2>&1 || error "gh (GitHub CLI) is required. Install: https://cli.github.com/"

# Get version from argument or prompt
VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
    # Get latest tag as suggestion
    LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
    read -rp "Enter version (latest: $LATEST_TAG): " VERSION
fi

# Validate version format
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
    error "Invalid version format. Use semver: v1.0.0 or v1.0.0-beta"
fi

log "Releasing version: $VERSION"

# Check for uncommitted changes
if [[ -n "$(git status --porcelain)" ]]; then
    error "You have uncommitted changes. Please commit or stash them first."
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    error "Tag $VERSION already exists"
fi

# Create and push tag first (so make can pick up the version)
log "Creating git tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

# Build all platforms
log "Building release artifacts..."
make clean
VERSION="$VERSION" make release

# Verify artifacts exist
ARTIFACTS=(
    "bin/parm-linux-amd64.tar.gz"
    "bin/parm-linux-arm64.tar.gz"
    "bin/parm-darwin-amd64.tar.gz"
    "bin/parm-darwin-arm64.tar.gz"
    "bin/parm-windows-amd64.zip"
)

for artifact in "${ARTIFACTS[@]}"; do
    [[ -f "$artifact" ]] || error "Missing artifact: $artifact"
done

log "All artifacts built successfully"
ls -lh bin/

# Create GitHub release
log "Creating GitHub release..."
RELEASE_NOTES="Release $VERSION

## Changes
- Bug fixes and improvements

## Installation
\`\`\`sh
curl -fsSL https://raw.githubusercontent.com/aleister1102/parm/master/scripts/install.sh | sh
\`\`\`
"

gh release create "$VERSION" \
    "${ARTIFACTS[@]}" \
    --repo aleister1102/parm \
    --title "$VERSION" \
    --notes "$RELEASE_NOTES"

log "Release $VERSION published successfully!"
log "View at: https://github.com/aleister1102/parm/releases/tag/$VERSION"
