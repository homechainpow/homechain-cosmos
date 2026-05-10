#!/bin/bash

################################################################################
# HomeChain V10 State-Sync Verification Script
# Verifies hash from 3 independent sources before state-sync
################################################################################

set -e

TRUST_HEIGHT=$1

if [ -z "$TRUST_HEIGHT" ]; then
    echo "Usage: $0 <trust_height>"
    exit 1
fi

BOOTNODE_URL="https://bootnode.homechain.io:26657"
EXPLORER_URL="https://explorer.homechain.io/api"
VALIDATOR3_URL="https://validator3.homechain.io:26657"

echo "Verifying state-sync hash for height: $TRUST_HEIGHT"
echo "========================================"

# Source 1: Official Bootnode
echo "Source 1: Official Bootnode"
HASH1=$(curl -s $BOOTNODE_URL/commit?height=$TRUST_HEIGHT | jq -r '.result.signed_header.commit.block_id.hash' 2>/dev/null)
if [ -z "$HASH1" ] || [ "$HASH1" = "null" ]; then
    echo "ERROR: Failed to get hash from bootnode"
    exit 1
fi
echo "Hash: $HASH1"
echo ""

# Source 2: Community Explorer
echo "Source 2: Community Explorer"
HASH2=$(curl -s $EXPLORER_URL/block/$TRUST_HEIGHT | jq -r '.hash' 2>/dev/null)
if [ -z "$HASH2" ] || [ "$HASH2" = "null" ]; then
    echo "ERROR: Failed to get hash from explorer"
    exit 1
fi
echo "Hash: $HASH2"
echo ""

# Source 3: Independent Validator
echo "Source 3: Independent Validator"
HASH3=$(curl -s $VALIDATOR3_URL/commit?height=$TRUST_HEIGHT | jq -r '.result.signed_header.commit.block_id.hash' 2>/dev/null)
if [ -z "$HASH3" ] || [ "$HASH3" = "null" ]; then
    echo "ERROR: Failed to get hash from validator"
    exit 1
fi
echo "Hash: $HASH3"
echo ""

# Verification
echo "========================================"
if [ "$HASH1" = "$HASH2" ] && [ "$HASH2" = "$HASH3" ]; then
    echo "✅ VERIFIED: All 3 sources match"
    echo "TRUST_HASH=$HASH1"
    echo ""
    echo "Add to ~/.homechain/config/config.toml:"
    echo "[statesync]"
    echo "enable = true"
    echo "rpc_servers = \"$BOOTNODE_URL,$VALIDATOR3_URL\""
    echo "trust_height = $TRUST_HEIGHT"
    echo "trust_hash = \"$HASH1\""
    echo "trust_period = \"168h0m0s\""
else
    echo "❌ WARNING: Hash mismatch detected!"
    echo "DO NOT proceed with state-sync from untrusted sources"
    echo ""
    echo "Bootnode:  $HASH1"
    echo "Explorer:  $HASH2"
    echo "Validator: $HASH3"
    exit 1
fi
