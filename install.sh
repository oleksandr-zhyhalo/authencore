#!/bin/bash

set -e

# Constants
GITHUB_REPO="oleksandr-zhyhalo/authencore"
BINARY_NAME="authencore"
BINARY_PATH="/usr/local/bin/${BINARY_NAME}"
CONFIG_DIR="/etc/authencore"
LOG_DIR="/var/log/authencore"

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print error message and exit
error() {
    echo -e "${RED}Error: $1${NC}" >&2
    exit 1
}

# Print success message
success() {
    echo -e "${GREEN}$1${NC}"
}

# Print info message
info() {
    echo -e "${YELLOW}$1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Get system information
get_arch() {
    local arch
    arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        arm64|aarch64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
}

get_os() {
    local os
    os=$(uname -s)
    case $os in
        Linux)
            echo "linux"
            ;;
        Darwin)
            echo "darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
}

# Get appropriate group name based on OS
get_group_name() {
    if [ "$(get_os)" = "darwin" ]; then
        echo "staff"
    else
        echo "${SUDO_USER}"
    fi
}

# Download and install binary
install_binary() {
    local version tmp_dir asset_name download_url
    local os arch

    # Get latest release version
    if command_exists curl; then
        version=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep -o '"tag_name": ".*"' | cut -d'"' -f4)
    elif command_exists wget; then
        version=$(wget -qO- https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep -o '"tag_name": ".*"' | cut -d'"' -f4)
    else
        error "Either curl or wget is required"
    fi

    [ -z "$version" ] && error "Could not determine latest version"

    os=$(get_os)
    arch=$(get_arch)

    # Create temporary directory
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT

    # Download binary
    asset_name="authencore-${os}-${arch}.tar.gz"
    download_url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${asset_name}"

    info "Downloading ${asset_name}..."
    if command_exists curl; then
        curl -sL "$download_url" -o "$tmp_dir/authencore.tar.gz" || error "Failed to download authencore"
        tar xzf "$tmp_dir/authencore.tar.gz" -C "$tmp_dir" || error "Failed to extract archive"
    elif command_exists wget; then
        wget -qO "$tmp_dir/authencore.tar.gz" "$download_url" || error "Failed to download authencore"
        tar xzf "$tmp_dir/authencore.tar.gz" -C "$tmp_dir" || error "Failed to extract archive"
    fi

    # Verify and install files
    [ ! -f "$tmp_dir/authencore" ] && error "Binary not found in downloaded archive"
    [ ! -f "$tmp_dir/authencore.yml.sample" ] && error "Config sample not found in downloaded archive"

    # Install binary
    info "Installing binary to ${BINARY_PATH}..."
    install -m 755 "$tmp_dir/authencore" "$BINARY_PATH"

    # Install config
    if [ ! -f "${CONFIG_DIR}/authencore.yml" ]; then
        info "Installing default configuration..."
        install -m 600 -o "${SUDO_USER}" -g "$(get_group_name)" "$tmp_dir/authencore.yml.sample" "${CONFIG_DIR}/authencore.yml"
    else
        info "Config file already exists, installing sample as reference..."
        install -m 600 -o "${SUDO_USER}" -g "$(get_group_name)" "$tmp_dir/authencore.yml.sample" "${CONFIG_DIR}/authencore.yml.sample"
    fi

    success "Downloaded and installed $version"
}

# Main installation process
main() {
    # Check if running with sudo
    if [ "$EUID" -ne 0 ]; then
        error "Please run with sudo"
    fi

    # Get the user who ran sudo
    SUDO_USER="${SUDO_USER:-$USER}"
    if [ "$SUDO_USER" = "root" ]; then
        error "Please run with sudo instead of as root directly"
    fi

    info "Installing authencore..."

    # Create required directories
    mkdir -p "${CONFIG_DIR}"
    mkdir -p "${LOG_DIR}"

    # Set directory permissions
    chown "${SUDO_USER}:$(get_group_name)" "${CONFIG_DIR}"
    chmod 750 "${CONFIG_DIR}"

    chown "${SUDO_USER}:$(get_group_name)" "${LOG_DIR}"
    chmod 750 "${LOG_DIR}"

    # Install binary and config
    install_binary

    success "Installation completed successfully!"
    echo
    echo "To use authencore, add the following to your AWS CLI config (~/.aws/config):"
    echo
    echo "[profile your-profile]"
    echo "credential_process = /usr/local/bin/authencore"
}

main