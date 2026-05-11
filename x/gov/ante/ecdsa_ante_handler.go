package ante

import (
	"encoding/hex"
	"fmt"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
)

// ECDSASigVerificationDecorator verifies Ethereum-style ECDSA signatures
// on HomeChain messages. It implements sdk.AnteDecorator so it can be
// chained with the standard SDK ante decorators.
//
// HomeChain uses Ethereum-style (secp256k1) addresses and EIP-191
// personal_sign signatures instead of the default Ed25519/Amino scheme.
type ECDSASigVerificationDecorator struct{}

// NewECDSASigVerificationDecorator creates a new ECDSA signature verification decorator
func NewECDSASigVerificationDecorator() ECDSASigVerificationDecorator {
	return ECDSASigVerificationDecorator{}
}

// AnteHandle implements sdk.AnteDecorator.
// For each message in the transaction, it checks if the message implements
// ECDSASignedMessage. If so, it verifies the ECDSA signature against the
// declared signer using EIP-191 personal_sign.
//
// If simulate=true (dry-run), signature verification is skipped.
// Standard SDK messages (bank.Send, etc.) that don't implement ECDSASignedMessage
// are passed through to the next decorator.
func (esvd ECDSASigVerificationDecorator) AnteHandle(
	ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler,
) (sdk.Context, error) {
	// Skip verification during simulation (e.g. query simulation)
	if simulate {
		return next(ctx, tx, simulate)
	}

	// Extract signer addresses from the transaction using SDK v0.50 SigVerifiableTx
	sigVerifiableTx, ok := tx.(signing.SigVerifiableTx)
	if !ok {
		return next(ctx, tx, simulate)
	}

	signerAddrs, err := sigVerifiableTx.GetSigners()
	if err != nil {
		return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized, fmt.Sprintf("failed to get signers: %v", err))
	}

	// Check each message for ECDSA signature
	msgs := tx.GetMsgs()
	for i, msg := range msgs {
		// Try to extract ECDSA signature from message if it implements ECDSASignedMessage
		type ecdsaSignedMsg interface {
			GetEcdsaSignature() string
			GetSigner() string
		}

		if signedMsg, ok := msg.(ecdsaSignedMsg); ok {
			sig := signedMsg.GetEcdsaSignature()
			msgSigner := signedMsg.GetSigner()

			if sig == "" {
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized,
					fmt.Sprintf("empty ECDSA signature for signer %s", msgSigner))
			}

			// Build the sign bytes from the message
			signBytes := getSignBytes(msg)

			if err := ValidateEthereumSignature(msgSigner, sig, string(signBytes)); err != nil {
				return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized,
					fmt.Sprintf("ECDSA signature verification failed: %v", err))
			}

			// Verify signer address matches the tx-level signer
			if i < len(signerAddrs) {
				signerEthAddr := ConvertCosmosToEthereumAddress(sdk.AccAddress(signerAddrs[i]))
				if !strings.EqualFold(msgSigner, signerEthAddr) {
					return ctx, errorsmod.Wrap(sdkerrors.ErrUnauthorized,
						fmt.Sprintf("signer mismatch: msg signer %s != tx signer %s", msgSigner, signerEthAddr))
				}
			}
		}
		// Messages that don't implement ECDSASignedMessage are passed through
		// (e.g. standard SDK messages like bank.Send)
	}

	return next(ctx, tx, simulate)
}

// getSignBytes produces the message bytes that should be signed.
// Uses the protobuf string representation as the signing payload.
func getSignBytes(msg sdk.Msg) []byte {
	return []byte(msg.String())
}

// ValidateEthereumSignature validates an Ethereum EIP-191 personal_sign signature
func ValidateEthereumSignature(address, signature, message string) error {
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}

	if len(sigBytes) != 65 {
		return fmt.Errorf("invalid signature length: expected 65, got %d", len(sigBytes))
	}

	// EIP-191 personal_sign format
	eip191Message := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	messageHash := ethcrypto.Keccak256Hash([]byte(eip191Message))

	// Recover public key from signature
	pubKey, err := ethcrypto.SigToPub(messageHash[:], sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	recoveredAddr := ethcrypto.PubkeyToAddress(*pubKey)
	providedAddr := common.HexToAddress(address)

	if recoveredAddr != providedAddr {
		return fmt.Errorf("recovered %s != provided %s", recoveredAddr.Hex(), providedAddr.Hex())
	}

	return nil
}

// GetEthereumAddressFromPubKey converts a secp256k1 public key (65 bytes, uncompressed) to Ethereum address
func GetEthereumAddressFromPubKey(pubKey []byte) (string, error) {
	if len(pubKey) != 65 {
		return "", fmt.Errorf("invalid public key length: expected 65, got %d", len(pubKey))
	}

	pubKeyBytes := pubKey[1:] // Remove 0x04 prefix
	hash := ethcrypto.Keccak256Hash(pubKeyBytes)
	address := hash.Bytes()[12:]

	return fmt.Sprintf("0x%x", address), nil
}

// ConvertCosmosToEthereumAddress converts a Cosmos address to Ethereum address format
func ConvertCosmosToEthereumAddress(cosmosAddr sdk.AccAddress) string {
	if len(cosmosAddr) >= 20 {
		ethAddr := cosmosAddr[len(cosmosAddr)-20:]
		return fmt.Sprintf("0x%x", ethAddr)
	}

	padded := make([]byte, 20)
	copy(padded[20-len(cosmosAddr):], cosmosAddr)
	return fmt.Sprintf("0x%x", padded)
}

// ConvertEthereumToCosmosAddress converts an Ethereum address to Cosmos address format
func ConvertEthereumToCosmosAddress(ethAddr string) (sdk.AccAddress, error) {
	cleanAddr := strings.TrimPrefix(ethAddr, "0x")

	if len(cleanAddr) != 40 {
		return nil, fmt.Errorf("invalid Ethereum address length: expected 40 hex chars, got %d", len(cleanAddr))
	}

	addrBytes, err := hex.DecodeString(cleanAddr)
	if err != nil {
		return nil, fmt.Errorf("invalid Ethereum address hex: %w", err)
	}

	if len(addrBytes) < 20 {
		padded := make([]byte, 20)
		copy(padded[20-len(addrBytes):], addrBytes)
		addrBytes = padded
	}

	return sdk.AccAddress(addrBytes), nil
}
