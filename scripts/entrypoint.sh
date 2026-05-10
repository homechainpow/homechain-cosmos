#!/bin/sh
# HomeChain Docker Entrypoint Script
# Handles node initialization and P2P peer configuration

set -e

# Default chain ID
CHAIN_ID=${CHAIN_ID:-homechain_9000-1}
MONIKER=${MONIKER:-homechain-node}
HOME_DIR=${HOME:-/root/.homechain}

# Check if node is initialized
if [ ! -f "$HOME_DIR/config/config.toml" ]; then
    echo "======================================"
    echo "Initializing HomeChain node: $MONIKER"
    echo "======================================"
    
    # Initialize the node
    homechaind init "$MONIKER" --chain-id "$CHAIN_ID" --home "$HOME_DIR"
    
    echo "Node initialized successfully!"
    echo "Node ID: $(homechaind tendermint show-node-id --home "$HOME_DIR")"
fi

# Configure P2P settings if environment variables are set
if [ -n "$P2P_PERSISTENT_PEERS" ]; then
    echo "Configuring persistent peers: $P2P_PERSISTENT_PEERS"
    sed -i "s/^persistent_peers = .*/persistent_peers = \"$P2P_PERSISTENT_PEERS\"/" "$HOME_DIR/config/config.toml"
fi

if [ -n "$P2P_SEEDS" ]; then
    echo "Configuring seeds: $P2P_SEEDS"
    sed -i "s/^seeds = .*/seeds = \"$P2P_SEEDS\"/" "$HOME_DIR/config/config.toml"
fi

# Update moniker in config
sed -i "s/^moniker = .*/moniker = \"$MONIKER\"/" "$HOME_DIR/config/config.toml"

# Set proper permissions
chmod -R 755 "$HOME_DIR"

echo "======================================"
echo "Starting HomeChain node: $MONIKER"
echo "======================================"
echo "Node ID: $(homechaind tendermint show-node-id --home "$HOME_DIR")"
echo ""

# Execute the main command
exec homechaind "$@"
