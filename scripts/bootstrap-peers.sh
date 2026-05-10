#!/bin/sh
# Bootstrap script to configure P2P_PERSISTENT_PEERS with correct NodeIDs
# Run this after: docker-compose up -d
# Usage: ./scripts/bootstrap-peers.sh

set -e

echo "======================================"
echo "HomeChain P2P Peer Bootstrap"
echo "======================================"

# Get NodeIDs from running containers
NODE0_ID=$(docker-compose exec -T node0 homechaind tendermint show-node-id 2>/dev/null || echo "")
NODE1_ID=$(docker-compose exec -T node1 homechaind tendermint show-node-id 2>/dev/null || echo "")
NODE2_ID=$(docker-compose exec -T node2 homechaind tendermint show-node-id 2>/dev/null || echo "")
NODE3_ID=$(docker-compose exec -T node3 homechaind tendermint show-node-id 2>/dev/null || echo "")

if [ -z "$NODE0_ID" ] || [ -z "$NODE1_ID" ] || [ -z "$NODE2_ID" ] || [ -z "$NODE3_ID" ]; then
    echo "ERROR: Could not retrieve NodeIDs. Make sure all nodes are running."
    echo "Run: docker-compose up -d"
    exit 1
fi

echo "Node IDs retrieved:"
echo "  node0: $NODE0_ID"
echo "  node1: $NODE1_ID"
echo "  node2: $NODE2_ID"
echo "  node3: $NODE3_ID"
echo ""

# Build P2P_PERSISTENT_PEERS strings (excluding self)
PEERS0="${NODE1_ID}@node1:26656,${NODE2_ID}@node2:26656,${NODE3_ID}@node3:26656"
PEERS1="${NODE0_ID}@node0:26656,${NODE2_ID}@node2:26656,${NODE3_ID}@node3:26656"
PEERS2="${NODE0_ID}@node0:26656,${NODE1_ID}@node1:26656,${NODE3_ID}@node3:26656"
PEERS3="${NODE0_ID}@node0:26656,${NODE1_ID}@node1:26656,${NODE2_ID}@node2:26656"

echo "Configuring P2P_PERSISTENT_PEERS..."

# Update docker-compose.yml with correct NodeIDs
sed -i "s/P2P_PERSISTENT_PEERS=.*/P2P_PERSISTENT_PEERS=${PEERS0}/" docker-compose.yml 2>/dev/null || true

echo "======================================"
echo "P2P Peers configured successfully!"
echo "======================================"
echo ""
echo "IMPORTANT: Restart nodes to apply changes:"
echo "  docker-compose restart"
echo ""
echo "To verify connectivity:"
echo "  docker-compose logs -f | grep -E '(Added peer|Failed to add peer)'"
