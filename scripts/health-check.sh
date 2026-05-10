#!/bin/bash

################################################################################
# HomeChain V10 Health Check Script
# Monitors node health and alerts on issues
################################################################################

BINARY_NAME="homechaind"
RPC_URL="http://localhost:26657"

check_sync() {
    SYNC_STATUS=$(curl -s $RPC_URL/status | jq -r '.result.sync_info.catching_up')
    BLOCK_HEIGHT=$(curl -s $RPC_URL/status | jq -r '.result.sync_info.latest_block_height')
    
    if [ "$SYNC_STATUS" = "true" ]; then
        echo "⚠️  WARNING: Node is catching up (block: $BLOCK_HEIGHT)"
        return 1
    else
        echo "✅ OK: Node synced (block: $BLOCK_HEIGHT)"
        return 0
    fi
}

check_peers() {
    PEERS=$(curl -s $RPC_URL/net_info | jq -r '.result.n_peers')
    
    if [ "$PEERS" -lt 3 ]; then
        echo "⚠️  WARNING: Low peer count ($PEERS)"
        return 1
    else
        echo "✅ OK: Peers connected ($PEERS)"
        return 0
    fi
}

check_validator() {
    VALIDATOR_STATUS=$(curl -s $RPC_URL/status | jq -r '.result.validator_info.voting_power')
    
    if [ -z "$VALIDATOR_STATUS" ] || [ "$VALIDATOR_STATUS" = "0" ]; then
        echo "⚠️  WARNING: Not a validator or jailed"
        return 1
    else
        echo "✅ OK: Validator active (power: $VALIDATOR_STATUS)"
        return 0
    fi
}

check_disk() {
    DISK_USAGE=$(df -h ~/.homechain | tail -1 | awk '{print $5}' | sed 's/%//')
    
    if [ "$DISK_USAGE" -gt 80 ]; then
        echo "⚠️  WARNING: Disk usage high ($DISK_USAGE%)"
        return 1
    else
        echo "✅ OK: Disk usage ($DISK_USAGE%)"
        return 0
    fi
}

check_memory() {
    MEM_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100}')
    
    if [ "$MEM_USAGE" -gt 80 ]; then
        echo "⚠️  WARNING: Memory usage high ($MEM_USAGE%)"
        return 1
    else
        echo "✅ OK: Memory usage ($MEM_USAGE%)"
        return 0
    fi
}

main() {
    echo "HomeChain V10 Health Check"
    echo "=========================="
    echo ""
    
    check_sync
    check_peers
    check_validator
    check_disk
    check_memory
    
    echo ""
    echo "Health check complete"
}

main
