package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/argon2"
)

// Epoch constants for rate calculation
// CRITICAL: Using big.Int to prevent overflow at Epoch 62
// Epoch 61 rate = 10 * 2^60 ≈ 1.15 * 10^19 (safe, ~62% of uint64 max)
// Epoch 62 rate = 10 * 2^61 ≈ 2.30 * 10^19 (OVERFLOW! > 1.84 * 10^19 uint64 max)
// uint64 max = 18,446,744,073,709,551,615
const (
	// EpochRateBase is the base rate for Epoch 1 (10 hashes per coin)
	EpochRateBase = 10
	// MaxEpochBeforeOverflow is the last safe epoch before uint64 overflow (Epoch 61)
	MaxEpochBeforeOverflow = 61
)

// PoHData represents Proof of History data with Argon2id hash
type PoHData struct {
	PrevHash   string `protobuf:"bytes,1,opt,name=prev_hash,json=prevHash,proto3" json:"prev_hash,omitempty" yaml:"prev_hash"`
	NewHash    string `protobuf:"bytes,2,opt,name=new_hash,json=newHash,proto3" json:"new_hash,omitempty" yaml:"new_hash"`
	Difficulty uint64 `protobuf:"varint,3,opt,name=difficulty,proto3" json:"difficulty,omitempty" yaml:"difficulty"`
	Timestamp  uint64 `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty" yaml:"timestamp"`
	Nonce      []byte `protobuf:"bytes,5,opt,name=nonce,proto3" json:"nonce,omitempty" yaml:"nonce"`
	Salt       []byte `protobuf:"bytes,6,opt,name=salt,proto3" json:"salt,omitempty" yaml:"salt"`
	Version    uint32 `protobuf:"varint,7,opt,name=version,proto3" json:"version,omitempty" yaml:"version"`
}

// NewPoHData creates a new PoHData instance
func NewPoHData(prevHash, newHash string, difficulty, timestamp uint64) PoHData {
	return PoHData{
		PrevHash:   prevHash,
		NewHash:    newHash,
		Difficulty: difficulty,
		Timestamp:  timestamp,
		Nonce:      make([]byte, 8),  // Default 8-byte nonce
		Salt:       make([]byte, 16), // Default 16-byte salt (Argon2id standard)
		Version:    1,                // Default version 1
	}
}

// NewPoHDataWithNonce creates a new PoHData instance with custom nonce and salt
func NewPoHDataWithNonce(prevHash, newHash string, difficulty, timestamp uint64, nonce, salt []byte) PoHData {
	return PoHData{
		PrevHash:   prevHash,
		NewHash:    newHash,
		Difficulty: difficulty,
		Timestamp:  timestamp,
		Nonce:      nonce,
		Salt:       salt,
		Version:    1,
	}
}

// Validate validates the PoHData
func (p PoHData) Validate() error {
	if len(p.PrevHash) != 64 {
		return fmt.Errorf("prev_hash must be 64 characters")
	}
	if len(p.NewHash) != 64 {
		return fmt.Errorf("new_hash must be 64 characters")
	}
	if p.Difficulty == 0 {
		return fmt.Errorf("difficulty must be greater than 0")
	}
	if p.Timestamp == 0 {
		return fmt.Errorf("timestamp must be greater than 0")
	}
	// Validate Nonce - minimum 8 bytes for security
	if len(p.Nonce) < 8 {
		return fmt.Errorf("nonce must be at least 8 bytes, got %d", len(p.Nonce))
	}
	// Validate Salt - minimum 8 bytes (16 recommended for Argon2id)
	if len(p.Salt) < 8 {
		return fmt.Errorf("salt must be at least 8 bytes (16 recommended), got %d", len(p.Salt))
	}
	return nil
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (p *PoHData) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (p *PoHData) Reset() {
	*p = PoHData{}
}

// String implements proto.Message interface for codec compatibility
func (p *PoHData) String() string {
	return fmt.Sprintf("PoHData{PrevHash: %s..., NewHash: %s..., Difficulty: %d, Timestamp: %d, Nonce: %x..., Salt: %x..., Version: %d}",
		p.PrevHash[:8], p.NewHash[:8], p.Difficulty, p.Timestamp, p.Nonce[:4], p.Salt[:4], p.Version)
}

// Argon2idConfig defines the configuration for Argon2id hashing
type Argon2idConfig struct {
	Time    uint32 // iterations
	Memory  uint32 // memory cost (KB)
	Threads uint32 // parallelism
	SaltLen uint32 // salt length
	HashLen uint32 // hash length
}

// DefaultArgon2idConfig returns the default Argon2id configuration
func DefaultArgon2idConfig() Argon2idConfig {
	return Argon2idConfig{
		Time:    3,     // iterations
		Memory:  65536, // 64 MB
		Threads: 4,     // parallelism
		SaltLen: 16,    // 16 bytes salt
		HashLen: 32,    // 32 bytes hash
	}
}

// HashWithArgon2id generates Argon2id hash using pure Go implementation
func HashWithArgon2id(input string, difficulty uint64) (string, error) {
	config := DefaultArgon2idConfig()

	password := []byte(input)
	salt := []byte(input)

	hash := argon2.IDKey(password, salt, config.Time, config.Memory, uint8(config.Threads), config.HashLen)
	return fmt.Sprintf("%x", hash), nil
}

// VerifyDifficulty checks if the hash meets the difficulty requirement
func VerifyDifficulty(hash string, difficulty uint64) bool {
	// Convert hash to big integer
	hashInt := new(big.Int)
	hashInt, ok := hashInt.SetString(hash, 16)
	if !ok {
		return false
	}

	// Calculate target: 2^(256 - difficulty)
	target := new(big.Int).Lsh(big.NewInt(1), uint(256-difficulty))

	// Check if hash is less than target
	return hashInt.Cmp(target) < 0
}

// ComputeHash generates a new PoH hash by chaining prevHash with nonce
func ComputeHash(prevHash string, nonce uint64, difficulty uint64) (string, error) {
	// Create deterministic input: prevHash + nonce
	input := fmt.Sprintf("%s:%d", prevHash, nonce)

	// Generate Argon2id hash
	hash, err := HashWithArgon2id(input, difficulty)
	if err != nil {
		return "", fmt.Errorf("failed to compute hash: %w", err)
	}

	return hash, nil
}

// FindValidNonce mines for a valid nonce that meets the difficulty requirement
func FindValidNonce(prevHash string, difficulty uint64, maxAttempts uint64) (uint64, string, error) {
	var nonce uint64 = 0

	for nonce < maxAttempts {
		hash, err := ComputeHash(prevHash, nonce, difficulty)
		if err != nil {
			return 0, "", err
		}

		if VerifyDifficulty(hash, difficulty) {
			return nonce, hash, nil
		}

		nonce++
	}

	return 0, "", fmt.Errorf("failed to find valid nonce after %d attempts", maxAttempts)
}

// VerifyPoHSequence verifies the PoH sequence is valid
func VerifyPoHSequence(prevHash, newHash string, difficulty uint64) bool {
	// Verify hash format
	if len(prevHash) != 64 || len(newHash) != 64 {
		return false
	}

	// Verify difficulty
	if !VerifyDifficulty(newHash, difficulty) {
		return false
	}

	return true
}

// VerifyPoHChain verifies that newHash is correctly derived from prevHash with given nonce
func VerifyPoHChain(prevHash, newHash string, nonce uint64, difficulty uint64) bool {
	// First verify the sequence
	if !VerifyPoHSequence(prevHash, newHash, difficulty) {
		return false
	}

	// Compute expected hash
	expectedHash, err := ComputeHash(prevHash, nonce, difficulty)
	if err != nil {
		return false
	}

	// Verify hash matches
	return expectedHash == newHash
}

// GenesisState defines the poh module's genesis state
type GenesisState struct {
	Params     Params    `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	PoHData    []PoHData `protobuf:"bytes,2,rep,name=poh_data,json=pohData,proto3" json:"poh_data"`
	Difficulty uint64    `protobuf:"varint,3,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (gs *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (gs *GenesisState) Reset() {
	*gs = GenesisState{}
}

// String implements proto.Message interface for codec compatibility
func (gs *GenesisState) String() string {
	return fmt.Sprintf("GenesisState{Difficulty: %d, PoHDataCount: %d}", gs.Difficulty, len(gs.PoHData))
}

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, pohData []PoHData, difficulty uint64) *GenesisState {
	return &GenesisState{
		Params:     params,
		PoHData:    pohData,
		Difficulty: difficulty,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:     DefaultParams(),
		PoHData:    []PoHData{},
		Difficulty: DefaultDifficulty,
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, poh := range gs.PoHData {
		if err := poh.Validate(); err != nil {
			return err
		}
	}

	if gs.Difficulty == 0 {
		return fmt.Errorf("difficulty must be greater than 0")
	}

	return nil
}

// Params defines the parameters for the poh module
type Params struct {
	HeartbeatPeriod uint64 `protobuf:"varint,1,opt,name=heartbeat_period,json=heartbeatPeriod,proto3" json:"heartbeat_period,omitempty"`
	MaxDifficulty   uint64 `protobuf:"varint,2,opt,name=max_difficulty,json=maxDifficulty,proto3" json:"max_difficulty,omitempty"`
	MinDifficulty   uint64 `protobuf:"varint,3,opt,name=min_difficulty,json=minDifficulty,proto3" json:"min_difficulty,omitempty"`
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (p *Params) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (p *Params) Reset() {
	*p = Params{}
}

// String implements proto.Message interface for codec compatibility
func (p *Params) String() string {
	return fmt.Sprintf("Params{HeartbeatPeriod: %d, MaxDifficulty: %d, MinDifficulty: %d}",
		p.HeartbeatPeriod, p.MaxDifficulty, p.MinDifficulty)
}

// NewParams creates a new Params instance
func NewParams(heartbeatPeriod, maxDifficulty, minDifficulty uint64) Params {
	return Params{
		HeartbeatPeriod: heartbeatPeriod,
		MaxDifficulty:   maxDifficulty,
		MinDifficulty:   minDifficulty,
	}
}

// DefaultParams returns default poh parameters
func DefaultParams() Params {
	return NewParams(
		HeartbeatPeriod,
		50, // max difficulty
		10, // min difficulty
	)
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.HeartbeatPeriod == 0 {
		return fmt.Errorf("heartbeat_period must be greater than 0")
	}
	if p.MaxDifficulty == 0 {
		return fmt.Errorf("max_difficulty must be greater than 0")
	}
	if p.MinDifficulty == 0 {
		return fmt.Errorf("min_difficulty must be greater than 0")
	}
	if p.MinDifficulty > p.MaxDifficulty {
		return fmt.Errorf("min_difficulty cannot be greater than max_difficulty")
	}
	return nil
}

// MsgSubmitPoH defines a message for submitting PoH data
type MsgSubmitPoH struct {
	Signer    string  `protobuf:"bytes,1,opt,name=signer,proto3" json:"signer,omitempty"`
	PohData   PoHData `protobuf:"bytes,2,opt,name=poh_data,json=pohData,proto3" json:"poh_data"`
	Signature string  `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

// NewMsgSubmitPoH creates a new MsgSubmitPoH instance
func NewMsgSubmitPoH(signer string, pohData PoHData, signature string) *MsgSubmitPoH {
	return &MsgSubmitPoH{
		Signer:    signer,
		PohData:   pohData,
		Signature: signature,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgSubmitPoH) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgSubmitPoH) Type() string { return "submit_poh" }

// GetSigners implements the sdk.Msg interface
func (msg MsgSubmitPoH) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgSubmitPoH) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgSubmitPoH) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Signer)
	if err != nil {
		return fmt.Errorf("invalid signer address: %s", err)
	}

	if err := msg.PohData.Validate(); err != nil {
		return err
	}

	if len(msg.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	return nil
}

// MsgSubmitPoHResponse defines the response for submitting PoH data
type MsgSubmitPoHResponse struct{}

// MsgVerifyPoHResponse defines the response for verifying PoH data
type MsgVerifyPoHResponse struct {
	Valid bool `protobuf:"varint,1,opt,name=valid,proto3" json:"valid,omitempty"`
}

// MsgServer is the interface for the PoH module's message server
type MsgServer interface {
	SubmitPoH(goCtx context.Context, msg *MsgSubmitPoH) (*MsgSubmitPoHResponse, error)
}

// QueryServer is the interface for the PoH module's query server
type QueryServer interface {
	CurrentPoH(goCtx context.Context, req *QueryCurrentPoHRequest) (*QueryCurrentPoHResponse, error)
	PoHSequence(goCtx context.Context, req *QueryPoHSequenceRequest) (*QueryPoHSequenceResponse, error)
}

// QueryCurrentPoHRequest is the request type for CurrentPoH query
type QueryCurrentPoHRequest struct{}

// QueryCurrentPoHResponse is the response type for CurrentPoH query
type QueryCurrentPoHResponse struct {
	PohData *PoHData `protobuf:"bytes,1,opt,name=poh_data,json=pohData,proto3" json:"poh_data,omitempty"`
}

// QueryPoHSequenceRequest is the request type for PoHSequence query
type QueryPoHSequenceRequest struct {
	Height uint64 `protobuf:"varint,1,opt,name=height,proto3" json:"height,omitempty"`
}

// QueryPoHSequenceResponse is the response type for PoHSequence query
type QueryPoHSequenceResponse struct {
	PohData *PoHData `protobuf:"bytes,1,opt,name=poh_data,json=pohData,proto3" json:"poh_data,omitempty"`
}

// RegisterInterfaces registers the interfaces for the module
// TODO: Uncomment after protobuf generation
/*
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitPoH{})
}
*/

// GetEpochRate calculates the hash-per-coin rate for a given epoch
// Uses big.Int to prevent overflow at Epoch 62+
// Formula: rate = 10 * 2^(epoch-1) for epoch >= 1
// Returns big.Int to handle arbitrarily large numbers
func GetEpochRate(epoch uint64) *big.Int {
	// Base rate: 10 hashes per coin at Epoch 1
	baseRate := big.NewInt(EpochRateBase)

	// Epoch 0 or 1: return base rate
	if epoch <= 1 {
		return baseRate
	}

	// Calculate 10 * 2^(epoch-1)
	// Using big.Int.Exp for safe exponentiation
	exponent := new(big.Int).Exp(
		big.NewInt(2),              // base: 2
		big.NewInt(int64(epoch-1)), // exponent: epoch-1
		nil,                        // modulus: nil (no modular arithmetic)
	)

	// Multiply base rate by 2^(epoch-1)
	rate := new(big.Int).Mul(baseRate, exponent)
	return rate
}

// GetEpochRateSafe returns the epoch rate as uint64 if safe, or an error if overflow would occur
// Use this when you need uint64 for storage/comparison, but check for overflow first
func GetEpochRateSafe(epoch uint64) (uint64, error) {
	rate := GetEpochRate(epoch)

	// Check if rate fits in uint64
	maxUint64 := new(big.Int).SetUint64(^uint64(0)) // 18446744073709551615
	if rate.Cmp(maxUint64) > 0 {
		return 0, fmt.Errorf("epoch rate overflow for epoch %d: exceeds uint64 max", epoch)
	}

	return rate.Uint64(), nil
}

// RegisterLegacyAminoCodec registers the amino codec for the module
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSubmitPoH{}, "homechain/MsgSubmitPoH", nil)
}

var ModuleCdc *codec.LegacyAmino

func init() {
	cdc := codec.NewLegacyAmino()
	RegisterLegacyAminoCodec(cdc)
	ModuleCdc = cdc
}

// ProtoMessage stub implementations for codec compatibility
func (m *MsgSubmitPoH) ProtoMessage() {}
func (m *MsgSubmitPoH) Reset()        { *m = MsgSubmitPoH{} }
func (m MsgSubmitPoH) String() string { return "MsgSubmitPoH" }

func (m *MsgSubmitPoHResponse) ProtoMessage() {}
func (m *MsgSubmitPoHResponse) Reset()        { *m = MsgSubmitPoHResponse{} }
func (m MsgSubmitPoHResponse) String() string { return "MsgSubmitPoHResponse" }

func (m *MsgVerifyPoHResponse) ProtoMessage() {}
func (m *MsgVerifyPoHResponse) Reset()        { *m = MsgVerifyPoHResponse{} }
func (m MsgVerifyPoHResponse) String() string { return "MsgVerifyPoHResponse" }
