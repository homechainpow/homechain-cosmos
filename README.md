# HomeChain V10

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org)
[![Cosmos SDK](https://img.shields.io/badge/Cosmos%20SDK-v0.50-3E3C3D?logo=cosmos)](https://cosmos.network)

A novel blockchain combining Proof of Hashrate (PoH) mining with a permissionless Bitcoin-style mesh network. Built on Cosmos SDK with Ethermint for EVM compatibility.

## 🌟 Key Features

- **Mobile-First Mining**: Argon2id PoH accessible on smartphones
- **Permissionless Mesh**: Bitcoin-style network with dynamic Top 100 validators
- **21-Level Referral System**: Multi-level rewards with economic self-defense
- **Node Runner Rewards**: Incentivized network infrastructure
- **Hybrid Governance**: 70% token stake + 30% hashrate voting power
- **EVM Compatible**: Full Ethereum Virtual Machine integration
- **No Pre-mine**: Fair launch with 0 tokens at genesis

## 📋 Documentation

- [API Specification](API_SPEC.yaml) - REST API documentation (OpenAPI)
- Architecture, Whitepaper, and Test Plan available upon request

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Ubuntu 22.04 LTS (recommended)
- 2 vCPU / 4GB RAM / 200GB SSD (minimum)

> **Windows Users**: Use WSL2 (Ubuntu) for best results. Native Windows builds may require `CGO_ENABLED=0` due to `go-ethereum/btcec` crypto dependencies.

### Installation

```bash
# Clone repository
git clone https://github.com/homechainpow/homechain-cosmos.git
cd homechain-cosmos

# Install dependencies
sudo apt install -y build-essential git wget jq curl gcc make \
    libssl-dev pkg-config protobuf-compiler
# Note: libargon2-dev not required — HomeChain uses golang.org/x/crypto/argon2 (pure Go)

# Build binary
make build

# Install
sudo cp build/homechaind /usr/local/bin/
sudo chmod +x /usr/local/bin/homechaind

# Verify
homechaind version
```

### Initialize Node

```bash
# Initialize
homechaind init <moniker> --chain-id homechain_9000-1

# Generate keys
homechaind keys add validator --keyring-backend file
homechaind keys add miner --keyring-backend file

# Configure P2P
# Edit ~/.homechain/config/config.toml
# Set external_address to your public IP

# Start
homechaind start
```

### Automated Deployment

Use the deployment script for automated setup:

```bash
# Make executable
chmod +x scripts/deploy-node.sh

# Run with sudo
sudo ./scripts/deploy-node.sh <moniker>
```

## 🔗 Network

| Network | Chain ID | Status |
|---------|----------|--------|
| Mainnet | homechain_9000-1 | Coming Soon |
| Testnet | homechain_9000-2 | Planned |

## 💻 Development

### Project Structure

```
homechain/
├── app/                 # Application entry point
├── cmd/                 # CLI commands
├── x/                   # Custom modules
│   ├── poh/            # Proof of History
│   ├── mining/         # Mining operations
│   ├── referral/       # Referral system
│   ├── nodestake/      # Node staking
│   ├── gov/            # Governance
│   └── evm/            # EVM integration (coming soon)
├── scripts/            # Deployment scripts
├── docs/               # Documentation
└── proto/              # Protocol buffers
```

### Build

```bash
# Build for local development
make build

# Build for Linux (deterministic)
make build-linux

# Build with coverage
make test-coverage

# Generate protobuf types (requires Docker or Linux)
make proto-gen
```

### Test

```bash
# Run unit tests
make test-unit

# Run integration tests
make test-integration

# Run all tests
make test-all
```

## 🔒 Security

For security vulnerabilities, please email security@homechain.io

## 🤝 Contributing

We welcome contributions! Please open an issue or pull request.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

## 🌐 Community

- **Website**: https://homechain.io
- **Twitter**: [@HomeChain](https://twitter.com/HomeChain)
- **Discord**: [Join Discord](https://discord.gg/homechain)
- **Forum**: [Community Forum](https://forum.homechain.io)

## 🙏 Acknowledgments

- Cosmos SDK - Application layer framework
- Ethermint - EVM integration
- CometBFT - Consensus engine
- Argon2 - Memory-hard hashing algorithm

## 📊 Tokenomics

- **Max Supply**: 21,000,000,000 HOME
- **No Pre-mine**: 0 tokens at genesis
- **Distribution**: 100% through mining
- **Epoch Size**: 168,000,000 HOME
- **Halving Period**: Every epoch (supply-driven, no fixed time)

---

**Anytime,Anywhere,Anyone Mining**
