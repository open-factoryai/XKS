#!/bin/bash

# XKS Installation Script
# Installs xks CLI tool for AKS cluster management

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="${XKS_INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="xks"
REPO_URL="https://github.com/horizon-ch/xks.git"
TEMP_DIR="/tmp/xks-install"

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_command() {
    if ! command -v "$1" &> /dev/null; then
        log_error "$1 is required but not installed"
        return 1
    fi
}

check_go_version() {
    local go_version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    local required_version="1.21"
    
    if ! echo "$go_version $required_version" | awk '{exit ($1 >= $2)}'; then
        log_error "Go version $required_version or higher is required (found: $go_version)"
        return 1
    fi
}

cleanup() {
    if [[ -d "$TEMP_DIR" ]]; then
        log_info "Cleaning up temporary files..."
        rm -rf "$TEMP_DIR"
    fi
}

install_xks() {
    log_info "Starting XKS installation..."
    
    # Check prerequisites
    log_info "Checking prerequisites..."
    check_command "go" || exit 1
    check_command "git" || exit 1
    check_command "az" || log_warning "Azure CLI not found. Install it from https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    
    check_go_version || exit 1
    
    # Create temporary directory
    log_info "Creating temporary directory..."
    mkdir -p "$TEMP_DIR"
    
    # Clone repository
    log_info "Cloning repository..."
    git clone "$REPO_URL" "$TEMP_DIR" || {
        log_error "Failed to clone repository"
        exit 1
    }
    
    # Build binary
    log_info "Building XKS binary..."
    cd "$TEMP_DIR"
    go mod tidy || {
        log_error "Failed to download dependencies"
        exit 1
    }
    
    go build -o "$BINARY_NAME" || {
        log_error "Failed to build binary"
        exit 1
    }
    
    # Install binary
    log_info "Installing binary to $INSTALL_DIR..."
    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_info "Installing with sudo (directory not writable)..."
        sudo cp "$BINARY_NAME" "$INSTALL_DIR/" || {
            log_error "Failed to install binary"
            exit 1
        }
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        cp "$BINARY_NAME" "$INSTALL_DIR/" || {
            log_error "Failed to install binary"
            exit 1
        }
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Verify installation
    if command -v "$BINARY_NAME" &> /dev/null; then
        local version
        version=$("$BINARY_NAME" --version 2>/dev/null || echo "unknown")
        log_success "XKS installed successfully!"
        log_info "Version: $version"
        log_info "Location: $INSTALL_DIR/$BINARY_NAME"
    else
        log_warning "Installation completed but $BINARY_NAME not found in PATH"
        log_info "You may need to add $INSTALL_DIR to your PATH"
    fi
}

show_usage() {
    cat << EOF
XKS Installation Script

Usage: $0 [OPTIONS]

Options:
    -h, --help              Show this help message
    -d, --install-dir DIR   Installation directory (default: /usr/local/bin)
    --uninstall            Uninstall XKS

Environment Variables:
    XKS_INSTALL_DIR        Installation directory

Examples:
    $0                     # Install to /usr/local/bin
    $0 -d ~/.local/bin     # Install to custom directory
    $0 --uninstall         # Remove XKS

EOF
}

uninstall_xks() {
    log_info "Uninstalling XKS..."
    
    local binary_path="$INSTALL_DIR/$BINARY_NAME"
    
    if [[ -f "$binary_path" ]]; then
        if [[ -w "$INSTALL_DIR" ]]; then
            rm "$binary_path"
        else
            sudo rm "$binary_path"
        fi
        log_success "XKS uninstalled successfully"
    else
        log_warning "XKS binary not found at $binary_path"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -d|--install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --uninstall)
            uninstall_xks
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Trap cleanup on exit
trap cleanup EXIT

# Run installation
install_xks

log_info "Next steps:"
log_info "1. Create a .env file with your Azure credentials"
log_info "2. Run 'xks --help' to see available commands"
log_info "3. Check the README for configuration details"
