#!/bin/bash

################################################################################
# HomeChain V10 Deployment Script
# Automated deployment for validator nodes
################################################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CHAIN_ID="homechain_9000-1"
BINARY_NAME="homechaind"
REPO_URL="https://github.com/homechain/homechain.git"
VERSION="v1.0.0"
MONIKER="${1:-homechain-validator}"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="$HOME/.homechain"

################################################################################
# Functions
################################################################################

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_root() {
    if [ "$EUID" -ne 0 ]; then 
        log_error "Please run as root or with sudo"
        exit 1
    fi
}

install_dependencies() {
    log_info "Installing dependencies..."
    apt update && apt upgrade -y
    apt install -y build-essential git wget jq curl gcc make \
        libssl-dev pkg-config protobuf-compiler libclang-dev \
        libargon2-dev
    log_info "Dependencies installed"
}

install_go() {
    log_info "Installing Go 1.24.2..."
    if command -v go &> /dev/null; then
        log_warn "Go is already installed: $(go version)"
    else
        wget https://go.dev/dl/go1.24.2.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.24.2.linux-amd64.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        source ~/.bashrc
        rm go1.24.2.linux-amd64.tar.gz
        log_info "Go 1.24.2 installed"
    fi
}

build_binary() {
    log_info "Building binary from source..."
    
    if [ -d "homechain" ]; then
        cd homechain
        git pull
    else
        git clone $REPO_URL homechain
        cd homechain
    fi
    
    git checkout $VERSION
    make build
    
    # Install binary
    sudo cp build/$BINARY_NAME $INSTALL_DIR/
    sudo chmod +x $INSTALL_DIR/$BINARY_NAME
    
    # Verify
    $BINARY_NAME version
    log_info "Binary built and installed"
}

init_node() {
    log_info "Initializing node..."
    
    # Remove existing data if present
    if [ -d "$DATA_DIR" ]; then
        log_warn "Removing existing data directory"
        rm -rf "$DATA_DIR"
    fi
    
    $BINARY_NAME init $MONIKER --chain-id $CHAIN_ID
    log_info "Node initialized"
}

generate_keys() {
    log_info "Generating keys..."
    
    # Validator key
    $BINARY_NAME keys add validator --keyring-backend file
    
    # Miner key
    $BINARY_NAME keys add miner --keyring-backend file
    
    # Node runner key
    $BINARY_NAME keys add noderunner --keyring-backend file
    
    log_info "Keys generated. SAVE YOUR MNEMONICS!"
}

configure_p2p() {
    log_info "Configuring P2P..."
    
    PUBLIC_IP=$(curl -s ifconfig.me)
    
    # Update config.toml
    sed -i "s/^laddr = .*/laddr = \"tcp:\/\/0.0.0.0:26656\"/" $DATA_DIR/config/config.toml
    sed -i "s/^external_address = .*/external_address = \"$PUBLIC_IP:26656\"/" $DATA_DIR/config/config.toml
    sed -i "s/^max_packet_msg_payload_size = .*/max_packet_msg_payload_size = 1048576/" $DATA_DIR/config/config.toml
    sed -i "s/^send_rate = .*/send_rate = 20480000/" $DATA_DIR/config/config.toml
    sed -i "s/^recv_rate = .*/recv_rate = 20480000/" $DATA_DIR/config/config.toml
    
    log_info "P2P configured with public IP: $PUBLIC_IP"
}

configure_app() {
    log_info "Configuring application..."
    
    # Update app.toml
    sed -i "s/^minimum-gas-prices = .*/minimum-gas-prices = \"0.001ahome\"/" $DATA_DIR/config/app.toml
    sed -i "s/^pruning = .*/pruning = \"custom\"/" $DATA_DIR/config/app.toml
    sed -i "s/^pruning-keep-recent = .*/pruning-keep-recent = \"2000\"/" $DATA_DIR/config/app.toml
    sed -i "s/^pruning-keep-every = .*/pruning-keep-every = \"0\"/" $DATA_DIR/config/app.toml
    sed -i "s/^pruning-interval = .*/pruning-interval = \"100\"/" $DATA_DIR/config/app.toml
    
    # EVM configuration
    sed -i "s/^trpc_laddr = .*/trpc_laddr = \"tcp:\/\/127.0.0.1:8545\"/" $DATA_DIR/config/app.toml
    sed -i "s/^ws_addr = .*/ws_addr = \"tcp:\/\/127.0.0.1:8546\"/" $DATA_DIR/config/app.toml
    
    log_info "Application configured"
}

setup_systemd() {
    log_info "Setting up systemd service..."
    
    cat <<EOF > /etc/systemd/system/homechaind.service
[Unit]
Description=HomeChain Node
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=$INSTALL_DIR/$BINARY_NAME start
Restart=always
RestartSec=3
StandardOutput=journal
StandardError=journal
LimitNOFILE=65535
MemoryMax=7G
Environment="HOME=$HOME"
Environment="GOMEMLIMIT=6GiB"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable homechaind
    log_info "Systemd service configured"
}

setup_firewall() {
    log_info "Configuring firewall..."
    
    ufw allow 22/tcp    # SSH
    ufw allow 26656/tcp # P2P
    ufw allow 26657/tcp # RPC (localhost only, but allow for monitoring)
    ufw --force enable
    
    log_info "Firewall configured"
}

show_validator_info() {
    log_info "Validator Information:"
    echo ""
    echo "Validator Address:"
    $BINARY_NAME keys show validator -a --keyring-backend file
    echo ""
    echo "Miner Address:"
    $BINARY_NAME keys show miner -a --keyring-backend file
    echo ""
    echo "Node Runner Address:"
    $BINARY_NAME keys show noderunner -a --keyring-backend file
    echo ""
    echo "Node ID:"
    $BINARY_NAME tendermint show-node-id
    echo ""
    log_warn "SAVE YOUR MNEMONICS AND PRIVATE KEYS SECURELY!"
}

main() {
    log_info "Starting HomeChain V10 deployment..."
    log_info "Moniker: $MONIKER"
    
    check_root
    install_dependencies
    install_go
    build_binary
    init_node
    generate_keys
    configure_p2p
    configure_app
    setup_systemd
    setup_firewall
    show_validator_info
    
    log_info "Deployment complete!"
    log_info "Next steps:"
    log_info "1. Wait for genesis file from genesis node"
    log_info "2. Copy genesis.json to $DATA_DIR/config/"
    log_info "3. Configure persistent_peers in config.toml"
    log_info "4. Start node: sudo systemctl start homechaind"
    log_info "5. Monitor logs: sudo journalctl -u homechaind -f"
}

# Run main function
main "$@"
