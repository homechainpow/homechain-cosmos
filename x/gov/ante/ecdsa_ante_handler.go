package ante

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ECDSAAnteHandler is a placeholder ante handler for SDK v0.50 compatibility
// TODO: Implement proper ECDSA signature verification for v0.50
// Note: sdk.Tx interface changed in v0.50 - GetSigners/GetSignatures removed
// Need to use tx.GetMsgs() + signing.GetSignBytes() approach

type ECDSAAnteHandler struct{}

// NewECDSAAnteHandler creates a new ECDSAAnteHandler (stub)
func NewECDSAAnteHandler() ECDSAAnteHandler {
	return ECDSAAnteHandler{}
}

// AnteHandle implements the AnteHandler interface - pass-through for now
func (eah ECDSAAnteHandler) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool) (sdk.Context, error) {
	return ctx, nil
}

// GetEthereumAddressFromPubKey converts a public key to Ethereum address
func GetEthereumAddressFromPubKey(pubKey []byte) (string, error) {
	if len(pubKey) != 65 {
		return "", fmt.Errorf("invalid public key length: expected 65, got %d", len(pubKey))
	}

	// Remove the first byte (0x04 for uncompressed public key)
	pubKeyBytes := pubKey[1:]

	// Hash with Keccak256 and take last 20 bytes
	hash := ethcrypto.Keccak256Hash(pubKeyBytes)
	address := hash.Bytes()[12:]

	return fmt.Sprintf("0x%x", address), nil
}

// ValidateEthereumSignature validates an Ethereum signature against a message
func ValidateEthereumSignature(address, signature, message string) error {
	// Convert hex signature to bytes
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(sigBytes) != 65 {
		return fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	// Create EIP-191 message
	eip191Message := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)

	// Hash with Keccak256
	messageHash := ethcrypto.Keccak256Hash([]byte(eip191Message))

	// Recover public key from signature
	pubKey, err := ethcrypto.SigToPub(messageHash[:], sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get recovered address
	recoveredAddr := ethcrypto.PubkeyToAddress(*pubKey)

	// Compare with provided address
	providedAddr := common.HexToAddress(address)

	if recoveredAddr != providedAddr {
		return fmt.Errorf("signature verification failed: recovered address %s != provided address %s",
			recoveredAddr.Hex(), providedAddr.Hex())
	}

	return nil
}

// ConvertCosmosToEthereumAddress converts a Cosmos address to Ethereum address format
func ConvertCosmosToEthereumAddress(cosmosAddr sdk.AccAddress) string {
	// Take the last 20 bytes of the Cosmos address
	if len(cosmosAddr) >= 20 {
		ethAddr := cosmosAddr[len(cosmosAddr)-20:]
		return fmt.Sprintf("0x%x", ethAddr)
	}

	// If address is shorter than 20 bytes, pad with zeros
	padded := make([]byte, 20)
	copy(padded[20-len(cosmosAddr):], cosmosAddr)
	return fmt.Sprintf("0x%x", padded)
}

// ConvertEthereumToCosmosAddress converts an Ethereum address to Cosmos address format
func ConvertEthereumToCosmosAddress(ethAddr string) (sdk.AccAddress, error) {
	// Remove 0x prefix
	cleanAddr := strings.TrimPrefix(ethAddr, "0x")

	// Validate length
	if len(cleanAddr) != 40 {
		return nil, fmt.Errorf("invalid Ethereum address length: expected 40 hex chars, got %d", len(cleanAddr))
	}

	// Decode hex
	addrBytes, err := hex.DecodeString(cleanAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid Ethereum address hex: %w", err)
	}

	// Convert to Cosmos address (pad to 20 bytes if needed)
	if len(addrBytes) < 20 {
		padded := make([]byte, 20)
		copy(padded[20-len(addrBytes):], addrBytes)
		addrBytes = padded
	}

	return sdk.AccAddress(addrBytes), nil
}
