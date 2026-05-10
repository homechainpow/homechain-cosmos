package app

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/homechain/homechain/x/poh/keeper"
	pohtypes "github.com/homechain/homechain/x/poh/types"
)

// ProposalHandler handles ABCI++ PrepareProposal and ProcessProposal for PoH verification
type ProposalHandler struct {
	pohKeeper keeper.Keeper
	baseApp   *baseapp.BaseApp
}

// NewProposalHandler creates a new ProposalHandler instance
func NewProposalHandler(pohKeeper keeper.Keeper, baseApp *baseapp.BaseApp) *ProposalHandler {
	return &ProposalHandler{
		pohKeeper: pohKeeper,
		baseApp:   baseApp,
	}
}

// PrepareProposal generates PoH data for block proposal
func (h *ProposalHandler) PrepareProposal(ctx sdk.Context, req *abci.RequestPrepareProposal) (*abci.ResponsePrepareProposal, error) {
	// Get current PoH state from keeper
	currentPoH, err := h.pohKeeper.GetCurrentPoH(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current PoH: %w", err)
	}

	// Generate new PoH hash using Argon2id
	newPoH, err := h.generatePoHHash(ctx, currentPoH, req.Time.Unix())
	if err != nil {
		return nil, fmt.Errorf("failed to generate PoH hash: %w", err)
	}

	// Create PoH transaction
	pohTx, err := h.createPoHTx(ctx, currentPoH, newPoH)
	if err != nil {
		return nil, fmt.Errorf("failed to create PoH transaction: %w", err)
	}

	// Add PoH transaction to the beginning of the proposal
	allTxs := make([][]byte, 0, len(req.Txs)+1)
	allTxs = append(allTxs, pohTx)
	allTxs = append(allTxs, req.Txs...)

	return &abci.ResponsePrepareProposal{
		Txs: allTxs,
	}, nil
}

// ProcessProposal verifies PoH data in block proposal
func (h *ProposalHandler) ProcessProposal(ctx sdk.Context, req *abci.RequestProcessProposal) (*abci.ResponseProcessProposal, error) {
	// Check if there are any transactions
	if len(req.Txs) == 0 {
		return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
	}

	// Extract PoH transaction from the first transaction
	pohData, err := h.extractPoHData(ctx, req.Txs[0])
	if err != nil {
		// If first transaction is not a PoH transaction, reject
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_REJECT,
		}, fmt.Errorf("failed to extract PoH data: %w", err)
	}

	// Verify PoH sequence
	if !h.pohKeeper.VerifyPoHSequence(ctx, pohData.PrevHash, pohData.NewHash, pohData.Difficulty) {
		return &abci.ResponseProcessProposal{
				Status: abci.ResponseProcessProposal_REJECT,
			}, fmt.Errorf("invalid PoH sequence: prev=%s, new=%s, difficulty=%d",
				pohData.PrevHash, pohData.NewHash, pohData.Difficulty)
	}

	// Verify timestamp is within acceptable range
	if !h.verifyTimestamp(pohData.Timestamp, ctx.BlockTime()) {
		return &abci.ResponseProcessProposal{
			Status: abci.ResponseProcessProposal_REJECT,
		}, fmt.Errorf("invalid PoH timestamp: %d", pohData.Timestamp)
	}

	// PoH verification passed, accept the proposal
	return &abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}, nil
}

// generatePoHHash generates new PoH hash using Argon2id
func (h *ProposalHandler) generatePoHHash(ctx sdk.Context, currentPoH *pohtypes.PoHData, timestamp int64) (*pohtypes.PoHData, error) {
	// Get current difficulty from keeper
	difficulty := h.pohKeeper.GetCurrentDifficulty(ctx)

	// Use FindValidNonce to find a valid hash
	maxAttempts := uint64(1000000) // 1 million attempts max
	nonce, newHash, err := pohtypes.FindValidNonce(currentPoH.PrevHash, difficulty, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("failed to find valid nonce: %w", err)
	}

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, nonce)
	return &pohtypes.PoHData{
		PrevHash:   currentPoH.NewHash,
		NewHash:    newHash,
		Difficulty: difficulty,
		Timestamp:  uint64(timestamp),
		Nonce:      nonceBytes,
		Version:    1,
	}, nil
}

// createPoHTx creates a PoH transaction for the proposal
func (h *ProposalHandler) createPoHTx(ctx sdk.Context, prevPoH, newPoH *pohtypes.PoHData) ([]byte, error) {
	// Create PoH submission message
	msg := &pohtypes.MsgSubmitPoH{
		Signer:    sdk.ConsAddress(ctx.BlockHeader().ProposerAddress).String(),
		PohData:   *newPoH,
		Signature: "", // Will be signed by proposer
	}

	// Create and sign transaction
	// Note: In a real implementation, this would involve proper transaction creation
	// and signing with the proposer's private key
	// For now, just serialize the message directly as a placeholder
	txBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// extractPoHData extracts PoH data from a transaction
func (h *ProposalHandler) extractPoHData(ctx sdk.Context, txBytes []byte) (*pohtypes.PoHData, error) {
	// Decode transaction - placeholder implementation
	// In SDK v0.50+, TxDecoder is unexported, so we use a simplified approach
	var msg pohtypes.MsgSubmitPoH
	if err := json.Unmarshal(txBytes, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PoH message: %w", err)
	}

	return &msg.PohData, nil
}

// verifyTimestamp checks if the PoH timestamp is within acceptable range
func (h *ProposalHandler) verifyTimestamp(pohTimestamp uint64, blockTime time.Time) bool {
	// Allow 30 seconds tolerance for clock skew
	tolerance := 30 * time.Second

	pohTime := time.Unix(int64(pohTimestamp), 0)

	// Check if PoH timestamp is within tolerance of block time
	diff := pohTime.Sub(blockTime)
	if diff < 0 {
		diff = -diff
	}

	return diff <= tolerance
}

// createTransaction creates a signed transaction (simplified)
func (h *ProposalHandler) createTransaction(ctx sdk.Context, msg sdk.Msg) ([]byte, error) {
	// In a real implementation, this would:
	// 1. Create a transaction builder
	// 2. Add the message
	// 3. Set fee and gas
	// 4. Sign with proposer's private key
	// 5. Encode to bytes

	// For now, return a mock transaction
	// This would be implemented with proper Cosmos SDK transaction creation
	return []byte("mock_poh_transaction"), nil
}

// VerifyPoHDifficulty checks if the PoH hash meets the difficulty requirement
func (h *ProposalHandler) VerifyPoHDifficulty(hash string, difficulty uint64) bool {
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
