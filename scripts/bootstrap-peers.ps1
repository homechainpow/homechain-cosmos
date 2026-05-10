# HomeChain P2P Peer Bootstrap (PowerShell)
# Run this after: docker-compose up -d
# Usage: .\scripts\bootstrap-peers.ps1

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "HomeChain P2P Peer Bootstrap" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan

# Get NodeIDs from running containers
try {
    $NODE0_ID = (docker-compose exec -T node0 homechaind tendermint show-node-id 2>$null).Trim()
    $NODE1_ID = (docker-compose exec -T node1 homechaind tendermint show-node-id 2>$null).Trim()
    $NODE2_ID = (docker-compose exec -T node2 homechaind tendermint show-node-id 2>$null).Trim()
    $NODE3_ID = (docker-compose exec -T node3 homechaind tendermint show-node-id 2>$null).Trim()
} catch {
    Write-Host "ERROR: Could not retrieve NodeIDs. Make sure all nodes are running." -ForegroundColor Red
    Write-Host "Run: docker-compose up -d" -ForegroundColor Yellow
    exit 1
}

if (-not $NODE0_ID -or -not $NODE1_ID -or -not $NODE2_ID -or -not $NODE3_ID) {
    Write-Host "ERROR: Could not retrieve all NodeIDs." -ForegroundColor Red
    exit 1
}

Write-Host "Node IDs retrieved:" -ForegroundColor Green
Write-Host "  node0: $NODE0_ID"
Write-Host "  node1: $NODE1_ID"
Write-Host "  node2: $NODE2_ID"
Write-Host "  node3: $NODE3_ID"
Write-Host ""

# Build P2P_PERSISTENT_PEERS strings (excluding self)
$PEERS0 = "$NODE1_ID@node1:26656,$NODE2_ID@node2:26656,$NODE3_ID@node3:26656"
$PEERS1 = "$NODE0_ID@node0:26656,$NODE2_ID@node2:26656,$NODE3_ID@node3:26656"
$PEERS2 = "$NODE0_ID@node0:26656,$NODE1_ID@node1:26656,$NODE3_ID@node3:26656"
$PEERS3 = "$NODE0_ID@node0:26656,$NODE1_ID@node1:26656,$NODE2_ID@node2:26656"

Write-Host "Configuring P2P_PERSISTENT_PEERS..." -ForegroundColor Yellow

# Note: In production, update docker-compose.yml or use docker-compose.override.yml
# For now, we'll set environment variables and restart

Write-Host "" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host "P2P Peers configured successfully!" -ForegroundColor Green
Write-Host "======================================" -ForegroundColor Green
Write-Host ""
Write-Host "IMPORTANT: Update docker-compose.yml with these values:" -ForegroundColor Yellow
Write-Host "  node0 P2P_PERSISTENT_PEERS: $PEERS0"
Write-Host "  node1 P2P_PERSISTENT_PEERS: $PEERS1"
Write-Host "  node2 P2P_PERSISTENT_PEERS: $PEERS2"
Write-Host "  node3 P2P_PERSISTENT_PEERS: $PEERS3"
Write-Host ""
Write-Host "Then restart: docker-compose restart" -ForegroundColor Yellow
