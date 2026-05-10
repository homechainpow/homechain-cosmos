package types

import (
	"fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Referral tree constants - Aligned with ARCHITECTURE.md Section 8.3
const (
	// MaxReferralDepth is the maximum depth of the referral tree (21 levels per ARCHITECTURE.md)
	MaxReferralDepth = 21
)

// Percentage constants as sdk.Dec for deterministic arithmetic
// Using NewDecWithPrec(value, precision) where precision is number of decimal places
var (
	// REFERRAL_L1_POOL_PCT = 20% of pool = 10% of base (per ARCHITECTURE.md)
	// L1 gets 20% of the 50% referral pool = 10% of total base reward
	ReferralL1PoolPct = math.LegacyNewDecWithPrec(20, 2) // 0.20

	// REFERRAL_LN_POOL_PCT = 4% of pool each = 2% of base each (per ARCHITECTURE.md)
	// L2-L21 get 4% each of the 50% referral pool = 2% of total base reward each
	ReferralLnPoolPct = math.LegacyNewDecWithPrec(4, 2) // 0.04

	// ReferralPoolPercentage is the percentage of base reward going to referral pool
	// Per ARCHITECTURE.md: 50% to miner directly, 50% to referral pool
	ReferralPoolPercentage = math.LegacyNewDecWithPrec(50, 2) // 0.50

	// BranchCapPct is the maximum voting power per referral tree branch (anti-Sybil)
	// Per ARCHITECTURE.md Section 2.2: Max 5% per Referral-Branch
	BranchCapPct = math.LegacyNewDecWithPrec(5, 2) // 0.05

	// MinerDirectPercentage = 50% (1 - ReferralPoolPercentage)
	MinerDirectPercentage = math.LegacyNewDecWithPrec(50, 2) // 0.50
)

// ReferralBonuses returns the bonus percentage for each level (21 levels)
// Per ARCHITECTURE.md: L1 gets 20% of pool, L2-L21 get 4% each of pool
// Pool is 50% of base reward, so:
// - L1 = 20% of 50% = 10% of base
// - L2-L21 = 4% of 50% = 2% of base each
var ReferralBonuses = []math.LegacyDec{
	ReferralL1PoolPct, // Level 1 (direct referrer): 20% of pool = 10% of base
	ReferralLnPoolPct, // Level 2: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 3: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 4: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 5: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 6: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 7: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 8: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 9: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 10: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 11: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 12: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 13: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 14: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 15: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 16: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 17: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 18: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 19: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 20: 4% of pool = 2% of base
	ReferralLnPoolPct, // Level 21: 4% of pool = 2% of base
}

// OLD CONSTANTS - DEPRECATED, use ReferralL1PoolPct and ReferralLnPoolPct instead
// Kept for backward compatibility during migration
const (
	_ = 0 // placeholder to prevent empty const block
)

// ReferralNode represents a node in the referral tree
type ReferralNode struct {
	Address        string    `json:"address"`
	Referrer       string    `json:"referrer"`
	Referrals      []string  `json:"referrals"`       // Direct referrals
	TotalReferrals uint64    `json:"total_referrals"` // Including indirect
	Level          uint32    `json:"level"`           // Depth in tree
	EarnedRewards  sdk.Coins `json:"earned_rewards"`
	JoinedAt       uint64    `json:"joined_at"`
}

// NewReferralNode creates a new referral node
func NewReferralNode(address, referrer string, joinedAt uint64) ReferralNode {
	return ReferralNode{
		Address:        address,
		Referrer:       referrer,
		Referrals:      []string{},
		TotalReferrals: 0,
		Level:          0,
		EarnedRewards:  sdk.NewCoins(),
		JoinedAt:       joinedAt,
	}
}

// AddReferral adds a new referral to this node
func (rn *ReferralNode) AddReferral(address string) error {
	if address == "" {
		return fmt.Errorf("referral address cannot be empty")
	}
	if address == rn.Address {
		return fmt.Errorf("cannot refer yourself")
	}

	rn.Referrals = append(rn.Referrals, address)
	rn.TotalReferrals++
	return nil
}

// RemoveReferral removes a referral from this node
func (rn *ReferralNode) RemoveReferral(address string) error {
	for i, ref := range rn.Referrals {
		if ref == address {
			rn.Referrals = append(rn.Referrals[:i], rn.Referrals[i+1:]...)
			rn.TotalReferrals--
			return nil
		}
	}
	return fmt.Errorf("referral not found")
}

// IsValidReferralChain checks if a referral chain is valid (no cycles, within depth limit)
func IsValidReferralChain(tree map[string]ReferralNode, newAddress, referrer string, depth uint32) (bool, error) {
	// Check for empty addresses
	if newAddress == "" || referrer == "" {
		return false, fmt.Errorf("addresses cannot be empty")
	}

	// Check for self-referral
	if newAddress == referrer {
		return false, fmt.Errorf("self-referral not allowed")
	}

	// Check depth limit
	if depth > MaxReferralDepth {
		return false, fmt.Errorf("referral depth exceeds maximum of %d", MaxReferralDepth)
	}

	// Check for cycle: trace up the tree to ensure newAddress doesn't exist
	current := referrer
	visited := make(map[string]bool)

	for current != "" {
		if visited[current] {
			return false, fmt.Errorf("cycle detected in referral chain")
		}
		visited[current] = true

		// Check if we reached the new address (would create cycle)
		if current == newAddress {
			return false, fmt.Errorf("referral would create a cycle")
		}

		// Move up the tree
		if node, exists := tree[current]; exists {
			current = node.Referrer
		} else {
			break
		}
	}

	return true, nil
}

// ReferralTree manages the referral tree structure
type ReferralTree struct {
	Nodes map[string]ReferralNode
}

// NewReferralTree creates a new referral tree
func NewReferralTree() ReferralTree {
	return ReferralTree{
		Nodes: make(map[string]ReferralNode),
	}
}

// AddNode adds a new node to the referral tree
func (rt *ReferralTree) AddNode(address, referrer string, joinedAt uint64) error {
	// Check if address already exists
	if _, exists := rt.Nodes[address]; exists {
		return fmt.Errorf("address %s already exists in referral tree", address)
	}

	// If referrer provided, validate the chain
	if referrer != "" {
		if node, exists := rt.Nodes[referrer]; exists {
			valid, err := IsValidReferralChain(rt.Nodes, address, referrer, node.Level+1)
			if err != nil {
				return err
			}
			if !valid {
				return fmt.Errorf("invalid referral chain")
			}

			// Update referrers total referrals
			node.TotalReferrals++
			rt.Nodes[referrer] = node
		} else {
			return fmt.Errorf("referrer %s not found", referrer)
		}
	}

	// Create new node
	node := NewReferralNode(address, referrer, joinedAt)
	if referrer != "" {
		if refNode, exists := rt.Nodes[referrer]; exists {
			node.Level = refNode.Level + 1
			refNode.AddReferral(address)
			rt.Nodes[referrer] = refNode
		}
	}

	rt.Nodes[address] = node
	return nil
}

// GetReferralChain returns the chain of referrers up to max depth
func (rt *ReferralTree) GetReferralChain(address string, maxDepth int) ([]string, error) {
	chain := []string{}
	current := address
	depth := 0

	for current != "" && depth < maxDepth {
		node, exists := rt.Nodes[current]
		if !exists {
			break
		}

		if node.Referrer != "" {
			chain = append(chain, node.Referrer)
			current = node.Referrer
			depth++
		} else {
			break
		}
	}

	return chain, nil
}

// CalculateReferralBonus calculates the bonus for each level in the referral chain
// Per ARCHITECTURE.md:
// - 50% of base reward goes to miner directly
// - 50% goes to referral pool
// - Pool distribution: L1 (20%), L2-L21 (4% each) = 100% of pool
func CalculateReferralBonus(baseAmount sdk.Coins, chain []string) map[string]sdk.Coins {
	bonuses := make(map[string]sdk.Coins)

	if baseAmount.IsZero() || len(chain) == 0 {
		return bonuses
	}

	baseValue := baseAmount.AmountOf("uhome")

	// Calculate referral pool (50% of base per ARCHITECTURE.md)
	poolValue := math.LegacyNewDecFromInt(baseValue).Mul(ReferralPoolPercentage)

	for i, referrer := range chain {
		if i >= len(ReferralBonuses) || i >= MaxReferralDepth {
			break
		}

		// Calculate bonus as percentage of pool (not base)
		// L1: 20% of pool = 10% of base
		// L2-L21: 4% of pool each = 2% of base each
		poolPercent := ReferralBonuses[i]
		bonusAmount := poolValue.Mul(poolPercent).TruncateInt()

		bonuses[referrer] = sdk.NewCoins(sdk.NewCoin("uhome", bonusAmount))
	}

	return bonuses
}

// CalculateReferralDistribution breaks down base reward into miner direct + referral pool
// Per ARCHITECTURE.md Section 5.1 and 5.2:
// - 50% to miner directly (credited immediately)
// - 50% to referral pool (distributed daily if miner active)
func CalculateReferralDistribution(baseAmount sdk.Coins) (minerDirect sdk.Coins, referralPool sdk.Coins) {
	if baseAmount.IsZero() {
		return sdk.NewCoins(), sdk.NewCoins()
	}

	baseValue := baseAmount.AmountOf("uhome")
	baseDec := math.LegacyNewDecFromInt(baseValue)

	// 50% to miner directly
	minerValue := baseDec.Mul(MinerDirectPercentage).TruncateInt()
	minerDirect = sdk.NewCoins(sdk.NewCoin("uhome", minerValue))

	// 50% to referral pool
	poolValue := baseDec.Mul(ReferralPoolPercentage).TruncateInt()
	referralPool = sdk.NewCoins(sdk.NewCoin("uhome", poolValue))

	return minerDirect, referralPool
}

// VerifyReferralDistribution verifies that bonus distribution matches ARCHITECTURE.md specs
// Returns true if L1 gets 20% of pool and L2-L21 get 4% each
func VerifyReferralDistribution(bonuses map[string]sdk.Coins, poolAmount sdk.Coins) bool {
	if poolAmount.IsZero() {
		return true
	}

	poolValue := poolAmount.AmountOf("uhome")
	if poolValue.IsZero() {
		return true
	}

	poolDec := math.LegacyNewDecFromInt(poolValue)

	// Check each level's percentage
	level := 0
	for _, bonus := range bonuses {
		bonusValue := bonus.AmountOf("uhome")
		bonusDec := math.LegacyNewDecFromInt(bonusValue)
		percent := bonusDec.Quo(poolDec)

		if level == 0 {
			// L1 should be 20%
			l1Min := math.LegacyNewDecWithPrec(19, 2) // 0.19
			l1Max := math.LegacyNewDecWithPrec(21, 2) // 0.21
			if percent.LT(l1Min) || percent.GT(l1Max) {
				return false
			}
		} else {
			// L2-L21 should be 4% each
			lnMin := math.LegacyNewDecWithPrec(39, 3) // 0.039
			lnMax := math.LegacyNewDecWithPrec(41, 3) // 0.041
			if percent.LT(lnMin) || percent.GT(lnMax) {
				return false
			}
		}
		level++
	}

	return true
}

// GetReferralStats returns statistics for a referrer
func (rt *ReferralTree) GetReferralStats(address string) (direct, indirect uint64, earned sdk.Coins, err error) {
	node, exists := rt.Nodes[address]
	if !exists {
		return 0, 0, sdk.NewCoins(), fmt.Errorf("address not found")
	}

	direct = uint64(len(node.Referrals))
	indirect = node.TotalReferrals - direct
	earned = node.EarnedRewards

	return direct, indirect, earned, nil
}

// AddRewardToReferrer adds reward to a referrer's earned rewards
func (rt *ReferralTree) AddRewardToReferrer(address string, reward sdk.Coins) error {
	node, exists := rt.Nodes[address]
	if !exists {
		return fmt.Errorf("address not found")
	}

	node.EarnedRewards = node.EarnedRewards.Add(reward...)
	rt.Nodes[address] = node
	return nil
}

// IsReferrer checks if an address is a referrer of another
func (rt *ReferralTree) IsReferrer(potentialReferrer, referred string) bool {
	node, exists := rt.Nodes[referred]
	if !exists {
		return false
	}

	return node.Referrer == potentialReferrer
}

// GetTopReferrers returns the top N referrers by total referrals
func (rt *ReferralTree) GetTopReferrers(n int) []ReferralNode {
	var allNodes []ReferralNode
	for _, node := range rt.Nodes {
		allNodes = append(allNodes, node)
	}

	// Simple bubble sort by total referrals
	for i := 0; i < len(allNodes); i++ {
		for j := i + 1; j < len(allNodes); j++ {
			if allNodes[j].TotalReferrals > allNodes[i].TotalReferrals {
				allNodes[i], allNodes[j] = allNodes[j], allNodes[i]
			}
		}
	}

	if n > len(allNodes) {
		n = len(allNodes)
	}

	return allNodes[:n]
}

// ValidateReferralCode validates a referral code format
func ValidateReferralCode(code string) error {
	if code == "" {
		return fmt.Errorf("referral code cannot be empty")
	}

	if len(code) < 3 || len(code) > 20 {
		return fmt.Errorf("referral code must be between 3 and 20 characters")
	}

	// Only allow alphanumeric characters
	for _, char := range code {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return fmt.Errorf("referral code can only contain alphanumeric characters")
		}
	}

	return nil
}

// GenerateReferralCode generates a referral code from an address
func GenerateReferralCode(address string) string {
	if len(address) < 8 {
		return ""
	}

	// Use first 6 characters of address (excluding prefix)
	cleanAddr := strings.TrimPrefix(address, "home1")
	if len(cleanAddr) < 6 {
		cleanAddr = address
	}

	code := strings.ToUpper(cleanAddr[:6])
	return code
}

// GetReferralPath returns the full path from root to the given address
func (rt *ReferralTree) GetReferralPath(address string) ([]string, error) {
	path := []string{}
	current := address

	for current != "" {
		path = append([]string{current}, path...)
		node, exists := rt.Nodes[current]
		if !exists {
			break
		}
		current = node.Referrer
	}

	return path, nil
}
