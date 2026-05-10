package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/crypto/argon2"
)

// MinerInfo represents information about a miner
type MinerInfo struct {
	Address      string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty" yaml:"address"`
	DeviceInfo   string `protobuf:"bytes,2,opt,name=device_info,json=deviceInfo,proto3" json:"device_info,omitempty" yaml:"device_info"`
	RegisteredAt uint64 `protobuf:"varint,3,opt,name=registered_at,json=registeredAt,proto3" json:"registered_at,omitempty" yaml:"registered_at"`
	IsActive     bool   `protobuf:"varint,4,opt,name=is_active,json=isActive,proto3" json:"is_active,omitempty" yaml:"is_active"`
	LastShareAt  uint64 `protobuf:"varint,5,opt,name=last_share_at,json=lastShareAt,proto3" json:"last_share_at,omitempty" yaml:"last_share_at"`
	TotalShares  uint64 `protobuf:"varint,6,opt,name=total_shares,json=totalShares,proto3" json:"total_shares,omitempty" yaml:"total_shares"`
	ValidShares  uint64 `protobuf:"varint,7,opt,name=valid_shares,json=validShares,proto3" json:"valid_shares,omitempty" yaml:"valid_shares"`
	Referrer     string `protobuf:"bytes,8,opt,name=referrer,proto3" json:"referrer,omitempty" yaml:"referrer"`
}

// NewMinerInfo creates a new MinerInfo instance
func NewMinerInfo(address, deviceInfo, referrer string, registeredAt uint64) MinerInfo {
	return MinerInfo{
		Address:      address,
		DeviceInfo:   deviceInfo,
		RegisteredAt: registeredAt,
		IsActive:     true,
		LastShareAt:  registeredAt,
		TotalShares:  0,
		ValidShares:  0,
		Referrer:     referrer,
	}
}

// Validate validates the MinerInfo
func (m MinerInfo) Validate() error {
	if len(m.Address) == 0 {
		return fmt.Errorf("address cannot be empty")
	}
	if len(m.DeviceInfo) == 0 {
		return fmt.Errorf("device_info cannot be empty")
	}
	if m.RegisteredAt == 0 {
		return fmt.Errorf("registered_at must be greater than 0")
	}
	return nil
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (m *MinerInfo) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (m *MinerInfo) Reset() {
	*m = MinerInfo{}
}

// String implements proto.Message interface for codec compatibility
func (m *MinerInfo) String() string {
	return fmt.Sprintf("MinerInfo{Address: %s, DeviceInfo: %s, IsActive: %v}", m.Address, m.DeviceInfo, m.IsActive)
}

// ShareData represents mining share data with Argon2id hash
type ShareData struct {
	Miner      string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty" yaml:"miner"`
	PrevHash   string `protobuf:"bytes,2,opt,name=prev_hash,json=prevHash,proto3" json:"prev_hash,omitempty" yaml:"prev_hash"`
	Nonce      string `protobuf:"bytes,3,opt,name=nonce,proto3" json:"nonce,omitempty" yaml:"nonce"`
	Difficulty uint64 `protobuf:"varint,4,opt,name=difficulty,proto3" json:"difficulty,omitempty" yaml:"difficulty"`
	Timestamp  uint64 `protobuf:"varint,5,opt,name=timestamp,proto3" json:"timestamp,omitempty" yaml:"timestamp"`
	Hash       string `protobuf:"bytes,6,opt,name=hash,proto3" json:"hash,omitempty" yaml:"hash"`
}

// NewShareData creates a new ShareData instance
func NewShareData(miner, prevHash, nonce string, difficulty, timestamp uint64) ShareData {
	return ShareData{
		Miner:      miner,
		PrevHash:   prevHash,
		Nonce:      nonce,
		Difficulty: difficulty,
		Timestamp:  timestamp,
	}
}

// Validate validates the ShareData
func (s ShareData) Validate() error {
	if len(s.Miner) == 0 {
		return fmt.Errorf("miner cannot be empty")
	}
	if len(s.PrevHash) != 64 {
		return fmt.Errorf("prev_hash must be 64 characters")
	}
	if len(s.Nonce) == 0 {
		return fmt.Errorf("nonce cannot be empty")
	}
	if s.Difficulty == 0 {
		return fmt.Errorf("difficulty must be greater than 0")
	}
	if s.Timestamp == 0 {
		return fmt.Errorf("timestamp must be greater than 0")
	}
	return nil
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (s *ShareData) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (s *ShareData) Reset() {
	*s = ShareData{}
}

// String implements proto.Message interface for codec compatibility
func (s *ShareData) String() string {
	return fmt.Sprintf("ShareData{Miner: %s, Difficulty: %d, Hash: %s...}", s.Miner, s.Difficulty, s.Hash[:16])
}

// HashShare generates Argon2id hash for the share using C wrapper
func (s ShareData) HashShare() (string, error) {
	// Create input for Argon2id - combine all share data
	input := fmt.Sprintf("%s:%s:%s:%d:%d",
		s.Miner, s.PrevHash, s.Nonce, s.Difficulty, s.Timestamp)

	// Hash with Argon2id C wrapper for deterministic cross-language behavior
	return HashWithArgon2id(input, s.Difficulty)
}

// VerifyShare verifies the share hash meets difficulty requirement
func (s ShareData) VerifyShare() bool {
	if len(s.Hash) == 0 {
		return false
	}

	// Verify hash meets difficulty requirement
	return VerifyDifficulty(s.Hash, s.Difficulty)
}

// HashWithArgon2id generates hash using Go implementation
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

// GenesisState defines the mining module's genesis state
type GenesisState struct {
	Params Params      `protobuf:"bytes,1,opt,name=params,proto3" json:"params"`
	Miners []MinerInfo `protobuf:"bytes,2,rep,name=miners,proto3" json:"miners"`
	Shares []ShareData `protobuf:"bytes,3,rep,name=shares,proto3" json:"shares"`
}

// NewGenesisState creates a new GenesisState instance
func NewGenesisState(params Params, miners []MinerInfo, shares []ShareData) *GenesisState {
	return &GenesisState{
		Params: params,
		Miners: miners,
		Shares: shares,
	}
}

// DefaultGenesisState returns a default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
		Miners: []MinerInfo{},
		Shares: []ShareData{},
	}
}

// Validate performs genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	for _, miner := range gs.Miners {
		if err := miner.Validate(); err != nil {
			return err
		}
	}

	for _, share := range gs.Shares {
		if err := share.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (gs *GenesisState) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (gs *GenesisState) Reset() {
	*gs = GenesisState{}
}

// String implements proto.Message interface for codec compatibility
func (gs *GenesisState) String() string {
	return fmt.Sprintf("GenesisState{Miners: %d, Shares: %d}", len(gs.Miners), len(gs.Shares))
}

// Params defines the parameters for the mining module
type Params struct {
	MaxSharesPerBlock      uint64 `protobuf:"varint,1,opt,name=max_shares_per_block,json=maxSharesPerBlock,proto3" json:"max_shares_per_block,omitempty"`
	RateLimitPerMinute     uint64 `protobuf:"varint,2,opt,name=rate_limit_per_minute,json=rateLimitPerMinute,proto3" json:"rate_limit_per_minute,omitempty"`
	MinerInactiveThreshold uint64 `protobuf:"varint,3,opt,name=miner_inactive_threshold,json=minerInactiveThreshold,proto3" json:"miner_inactive_threshold,omitempty"`
	ShareDifficulty        uint64 `protobuf:"varint,4,opt,name=share_difficulty,json=shareDifficulty,proto3" json:"share_difficulty,omitempty"`
}

// NewParams creates a new Params instance
func NewParams(maxSharesPerBlock, rateLimitPerMinute, minerInactiveThreshold, shareDifficulty uint64) Params {
	return Params{
		MaxSharesPerBlock:      maxSharesPerBlock,
		RateLimitPerMinute:     rateLimitPerMinute,
		MinerInactiveThreshold: minerInactiveThreshold,
		ShareDifficulty:        shareDifficulty,
	}
}

// DefaultParams returns default mining parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMaxSharesPerBlock,
		DefaultRateLimitPerMinute,
		DefaultMinerInactiveThreshold,
		ShareDifficulty,
	)
}

// Validate validates the parameters
func (p Params) Validate() error {
	if p.MaxSharesPerBlock == 0 {
		return fmt.Errorf("max_shares_per_block must be greater than 0")
	}
	if p.RateLimitPerMinute == 0 {
		return fmt.Errorf("rate_limit_per_minute must be greater than 0")
	}
	if p.MinerInactiveThreshold == 0 {
		return fmt.Errorf("miner_inactive_threshold must be greater than 0")
	}
	if p.ShareDifficulty == 0 {
		return fmt.Errorf("share_difficulty must be greater than 0")
	}
	return nil
}

// ProtoMessage implements proto.Message interface for codec compatibility
func (p *Params) ProtoMessage() {}

// Reset implements proto.Message interface for codec compatibility
func (p *Params) Reset() {
	*p = Params{}
}

// String implements proto.Message interface for codec compatibility
func (p *Params) String() string {
	return fmt.Sprintf("Params{MaxSharesPerBlock: %d, Difficulty: %d}", p.MaxSharesPerBlock, p.ShareDifficulty)
}

// MsgRegisterMiner defines a message for registering a miner
type MsgRegisterMiner struct {
	Address    string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
	DeviceInfo string `protobuf:"bytes,2,opt,name=device_info,json=deviceInfo,proto3" json:"device_info,omitempty"`
	Referrer   string `protobuf:"bytes,3,opt,name=referrer,proto3" json:"referrer,omitempty"`
	Signature  string `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
}

// NewMsgRegisterMiner creates a new MsgRegisterMiner instance
func NewMsgRegisterMiner(address, deviceInfo, referrer, signature string) *MsgRegisterMiner {
	return &MsgRegisterMiner{
		Address:    address,
		DeviceInfo: deviceInfo,
		Referrer:   referrer,
		Signature:  signature,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgRegisterMiner) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgRegisterMiner) Type() string { return "register_miner" }

// GetSigners implements the sdk.Msg interface
func (msg MsgRegisterMiner) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgRegisterMiner) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgRegisterMiner) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Address)
	if err != nil {
		return fmt.Errorf("invalid address: %s", err)
	}

	if len(msg.DeviceInfo) == 0 {
		return fmt.Errorf("device_info cannot be empty")
	}

	if len(msg.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	return nil
}

// MsgRegisterMinerResponse defines the response for registering a miner
type MsgRegisterMinerResponse struct{}

// MsgSubmitShare defines a message for submitting a mining share
type MsgSubmitShare struct {
	Miner     string    `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	ShareData ShareData `protobuf:"bytes,2,opt,name=share_data,json=shareData,proto3" json:"share_data"`
	Signature string    `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

// NewMsgSubmitShare creates a new MsgSubmitShare instance
func NewMsgSubmitShare(miner string, shareData ShareData, signature string) *MsgSubmitShare {
	return &MsgSubmitShare{
		Miner:     miner,
		ShareData: shareData,
		Signature: signature,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgSubmitShare) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgSubmitShare) Type() string { return "submit_share" }

// GetSigners implements the sdk.Msg interface
func (msg MsgSubmitShare) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgSubmitShare) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgSubmitShare) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		return fmt.Errorf("invalid miner address: %s", err)
	}

	if err := msg.ShareData.Validate(); err != nil {
		return err
	}

	if len(msg.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	return nil
}

// MsgSubmitShareResponse defines the response for submitting a mining share
type MsgSubmitShareResponse struct {
	Accepted bool   `protobuf:"varint,1,opt,name=accepted,proto3" json:"accepted,omitempty"`
	Reason   string `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
}

// MsgMinerHeartbeat defines a message for miner heartbeat
type MsgMinerHeartbeat struct {
	Miner     string `protobuf:"bytes,1,opt,name=miner,proto3" json:"miner,omitempty"`
	Timestamp uint64 `protobuf:"varint,2,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Nonce     uint64 `protobuf:"varint,3,opt,name=nonce,proto3" json:"nonce,omitempty"`
	Signature string `protobuf:"bytes,4,opt,name=signature,proto3" json:"signature,omitempty"`
}

// NewMsgMinerHeartbeat creates a new MsgMinerHeartbeat instance
func NewMsgMinerHeartbeat(miner string, timestamp, nonce uint64, signature string) *MsgMinerHeartbeat {
	return &MsgMinerHeartbeat{
		Miner:     miner,
		Timestamp: timestamp,
		Nonce:     nonce,
		Signature: signature,
	}
}

// Route implements the sdk.Msg interface
func (msg MsgMinerHeartbeat) Route() string { return RouterKey }

// Type implements the sdk.Msg interface
func (msg MsgMinerHeartbeat) Type() string { return "miner_heartbeat" }

// GetSigners implements the sdk.Msg interface
func (msg MsgMinerHeartbeat) GetSigners() []sdk.AccAddress {
	signer, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{signer}
}

// GetSignBytes implements the sdk.Msg interface
func (msg MsgMinerHeartbeat) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// ValidateBasic implements the sdk.Msg interface
func (msg MsgMinerHeartbeat) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Miner)
	if err != nil {
		return fmt.Errorf("invalid miner address: %s", err)
	}

	if msg.Timestamp == 0 {
		return fmt.Errorf("timestamp must be greater than 0")
	}

	if len(msg.Signature) == 0 {
		return fmt.Errorf("signature cannot be empty")
	}

	return nil
}

// MsgMinerHeartbeatResponse defines the response for miner heartbeat
type MsgMinerHeartbeatResponse struct{}

// MsgServer is the interface for the mining module's message server
type MsgServer interface {
	RegisterMiner(goCtx context.Context, msg *MsgRegisterMiner) (*MsgRegisterMinerResponse, error)
	SubmitShare(goCtx context.Context, msg *MsgSubmitShare) (*MsgSubmitShareResponse, error)
	MinerHeartbeat(goCtx context.Context, msg *MsgMinerHeartbeat) (*MsgMinerHeartbeatResponse, error)
}

// RegisterInterfaces registers the interfaces for the module
// TODO: Uncomment after protobuf generation
/*
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterMiner{},
		&MsgSubmitShare{},
		&MsgMinerHeartbeat{})
}
*/

// RegisterLegacyAminoCodec registers the amino codec for the module
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterMiner{}, "homechain/MsgRegisterMiner", nil)
	cdc.RegisterConcrete(&MsgSubmitShare{}, "homechain/MsgSubmitShare", nil)
	cdc.RegisterConcrete(&MsgMinerHeartbeat{}, "homechain/MsgMinerHeartbeat", nil)
}

var ModuleCdc *codec.LegacyAmino

func init() {
	cdc := codec.NewLegacyAmino()
	RegisterLegacyAminoCodec(cdc)
	ModuleCdc = cdc
}

// ProtoMessage stub implementations for codec compatibility
func (m *MsgRegisterMiner) ProtoMessage() {}
func (m *MsgRegisterMiner) Reset()        { *m = MsgRegisterMiner{} }
func (m MsgRegisterMiner) String() string { return m.Address }

func (m *MsgRegisterMinerResponse) ProtoMessage() {}
func (m *MsgRegisterMinerResponse) Reset()        { *m = MsgRegisterMinerResponse{} }
func (m MsgRegisterMinerResponse) String() string { return "MsgRegisterMinerResponse" }

func (m *MsgSubmitShare) ProtoMessage() {}
func (m *MsgSubmitShare) Reset()        { *m = MsgSubmitShare{} }
func (m MsgSubmitShare) String() string { return m.Miner }

func (m *MsgSubmitShareResponse) ProtoMessage() {}
func (m *MsgSubmitShareResponse) Reset()        { *m = MsgSubmitShareResponse{} }
func (m MsgSubmitShareResponse) String() string { return "MsgSubmitShareResponse" }

func (m *MsgMinerHeartbeat) ProtoMessage() {}
func (m *MsgMinerHeartbeat) Reset()        { *m = MsgMinerHeartbeat{} }
func (m MsgMinerHeartbeat) String() string { return m.Miner }

func (m *MsgMinerHeartbeatResponse) ProtoMessage() {}
func (m *MsgMinerHeartbeatResponse) Reset()        { *m = MsgMinerHeartbeatResponse{} }
func (m MsgMinerHeartbeatResponse) String() string { return "MsgMinerHeartbeatResponse" }
