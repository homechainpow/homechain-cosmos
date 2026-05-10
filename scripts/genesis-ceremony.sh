#!/bin/bash

################################################################################
# HomeChain V10 Genesis Ceremony Script
# For genesis node only - generates genesis and distributes to validators
################################################################################

set -e

CHAIN_ID="homechain_9000-1"
BINARY_NAME="homechaind"
DATA_DIR="$HOME/.homechain"
GENESIS_TIME="2026-05-10T00:00:00Z"
BOOTSTRAP_ADDRESS="homechain1bootstrap000000000000000000000"
BOOTSTRAP_BALANCE="10000000000000000000" # 10 HOME

################################################################################
# Functions
################################################################################

log_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

log_warn() {
    echo -e "\033[1;33m[WARN]\033[0m $1"
}

log_error() {
    echo -e "\033[0;31m[ERROR]\033[0m $1"
}

collect_validator_pubkeys() {
    log_info "Collecting validator pubkeys from all nodes..."
    
    mkdir -p validators
    
    # List of validator nodes (update with actual IPs)
    VALIDATORS=(
        "vps5:54.87.46.165"
        "vps6:54.221.188.16"
        "vps7:63.176.149.75"
        "vps8:63.182.165.111"
    )
    
    for validator in "${VALIDATORS[@]}"; do
        NAME=$(echo $validator | cut -d: -f1)
        IP=$(echo $validator | cut -d: -f2)
        log_info "Collecting from $NAME ($IP)..."
        scp ubuntu@$IP:/home/ubuntu/.homechain/config/priv_validator_key.json ./validators/${NAME}.json
    done
    
    log_info "Validator pubkeys collected"
}

build_genesis() {
    log_info "Building genesis file..."
    
    cd $DATA_DIR/config
    
    # Initialize with custom genesis
    $BINARY_NAME init genesis-node --chain-id $CHAIN_ID
    
    # Update genesis time
    jq '.genesis_time = "'$GENESIS_TIME'"' genesis.json > genesis.tmp.json
    mv genesis.tmp.json genesis.json
    
    # Update consensus params
    jq '.consensus_params.block.max_bytes = 1048576' genesis.json > genesis.tmp.json
    jq '.consensus_params.evidence.max_age_num_blocks = 1500' genesis.tmp.json > genesis.json
    rm genesis.tmp.json
    
    # Add bootstrap account
    jq '.app_state.bank.balances += [{"address": "'$BOOTSTRAP_ADDRESS'", "coins": [{"denom": "ahome", "amount": "'$BOOTSTRAP_BALANCE'"}]}]' genesis.json > genesis.tmp.json
    mv genesis.tmp.json genesis.json
    
    # Add validators
    for validator in validators/*.json; do
        PUBKEY=$(jq -r '.pub_key.value' $validator)
        ADDRESS=$(jq -r '.address' $validator)
        $BINARY_NAME genesis add-genesis-account $ADDRESS 0aHOME
    done
    
    # Collect gentxs
    $BINARY_NAME genesis collect-gentxs
    
    # Validate genesis
    $BINARY_NAME genesis validate-genesis
    
    log_info "Genesis file built and validated"
}

distribute_genesis() {
    log_info "Distributing genesis file to all validators..."
    
    VALIDATORS=(
        "vps5:54.87.46.165"
        "vps6:54.221.188.16"
        "vps7:63.176.149.75"
        "vps8:63.182.165.111"
    )
    
    GENESIS_HASH=$(sha256sum $DATA_DIR/config/genesis.json | awk '{print $1}')
    log_info "Genesis hash: $GENESIS_HASH"
    
    for validator in "${VALIDATORS[@]}"; do
        NAME=$(echo $validator | cut -d: -f1)
        IP=$(echo $validator | cut -d: -f2)
        log_info "Distributing to $NAME ($IP)..."
        scp $DATA_DIR/config/genesis.json ubuntu@$IP:/home/ubuntu/.homechain/config/
    done
    
    log_info "Genesis file distributed"
}

verify_genesis() {
    log_info "Verifying genesis on all nodes..."
    
    VALIDATORS=(
        "vps5:54.87.46.165"
        "vps6:54.221.188.16"
        "vps7:63.176.149.75"
        "vps8:63.182.165.111"
    )
    
    GENESIS_HASH=$(sha256sum $DATA_DIR/config/genesis.json | awk '{print $1}')
    
    for validator in "${VALIDATORS[@]}"; do
        NAME=$(echo $validator | cut -d: -f1)
        IP=$(echo $validator | cut -d: -f2)
        REMOTE_HASH=$(ssh ubuntu@$IP "sha256sum /home/ubuntu/.homechain/config/genesis.json | awk '{print \$1}'")
        
        if [ "$GENESIS_HASH" = "$REMOTE_HASH" ]; then
            log_info "$NAME: OK"
        else
            log_error "$NAME: HASH MISMATCH!"
            exit 1
        fi
    done
    
    log_info "All genesis files verified"
}

distribute_bootstrap_gas() {
    log_info "Distributing bootstrap gas to first 100 validators..."
    
    # Import bootstrap key
    log_warn "Enter bootstrap key mnemonic:"
    $BINARY_NAME keys add bootstrap --recover --keyring-backend file
    
    # Distribute 0.1 HOME each for gas
    for i in {1..100}; do
        VALIDATOR_ADDR=$(cat validator_${i}_address.txt 2>/dev/null || echo "")
        if [ -n "$VALIDATOR_ADDR" ]; then
            $BINARY_NAME tx bank send bootstrap $VALIDATOR_ADDR 100000000000000000 \
                --chain-id $CHAIN_ID \
                --fees 1000aHOME \
                --yes
            log_info "Gas distributed to validator $i"
        fi
    done
    
    # Send remaining to community pool
    $BINARY_NAME tx bank send bootstrap homechain1communitypool0000000 9000000000000000000 \
        --chain-id $CHAIN_ID \
        --fees 1000aHOME \
        --yes
    
    log_info "Bootstrap gas distribution complete"
}

main() {
    log_info "Starting Genesis Ceremony..."
    
    collect_validator_pubkeys
    build_genesis
    distribute_genesis
    verify_genesis
    
    log_warn "IMPORTANT: Genesis file is ready!"
    log_warn "DO NOT start nodes until genesis_time: $GENESIS_TIME"
    log_warn "After genesis, run distribute_bootstrap_gas.sh to distribute gas"
}

main "$@"
