package certifiers

import (
	"time"

	crypto "github.com/tendermint/go-crypto"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/tendermint/types"
)

// Test Helper: ValKeys lets us simulate signing with many keys
type ValKeys []crypto.PrivKey

// GenValKeys produces an array of private keys to generate commits
func GenValKeys(n int) ValKeys {
	res := make(ValKeys, n)
	for i := range res {
		res[i] = crypto.GenPrivKeyEd25519().Wrap()
	}
	return res
}

// Change replaces the key at index i
func (v ValKeys) Change(i int) ValKeys {
	res := make(ValKeys, len(v))
	copy(res, v)
	res[i] = crypto.GenPrivKeyEd25519().Wrap()
	return res
}

// Extend adds n more keys (to remove, just take a slice)
func (v ValKeys) Extend(n int) ValKeys {
	extra := GenValKeys(n)
	return append(v, extra...)
}

// GenSecValKeys produces an array of secp256k1 private keys to generate commits
func GenSecValKeys(n int) ValKeys {
	res := make(ValKeys, n)
	for i := range res {
		res[i] = crypto.GenPrivKeySecp256k1().Wrap()
	}
	return res
}

// Extend adds n more secp256k1 keys (to remove, just take a slice)
func (v ValKeys) ExtendSec(n int) ValKeys {
	extra := GenSecValKeys(n)
	return append(v, extra...)
}

// ToValidators produces a list of validators from the set of keys
// The first key has weight `init` and it increases by `inc` every step
// so we can have all the same weight, or a simple linear distribution
// (should be enough for testing)
func (v ValKeys) ToValidators(init, inc int64) *types.ValidatorSet {
	res := make([]*types.Validator, len(v))
	for i, k := range v {
		res[i] = types.NewValidator(k.PubKey(), init+int64(i)*inc)
	}
	return types.NewValidatorSet(res)
}

// SignHeader properly signs the header with all keys from first to last exclusive
func (v ValKeys) SignHeader(header *types.Header, first, last int) *types.Commit {
	votes := make([]*types.Vote, len(v))

	// we need this list to keep the ordering...
	vset := v.ToValidators(1, 0)

	// fill in the votes we want
	for i := first; i < last; i++ {
		vote := makeVote(header, vset, v[i])
		votes[vote.ValidatorIndex] = vote
	}

	res := &types.Commit{
		BlockID:    types.BlockID{Hash: header.Hash()},
		Precommits: votes,
	}
	return res
}

func makeVote(header *types.Header, vals *types.ValidatorSet, key crypto.PrivKey) *types.Vote {
	addr := key.PubKey().Address()
	idx, _ := vals.GetByAddress(addr)
	vote := &types.Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   idx,
		Height:           header.Height,
		Round:            1,
		Type:             types.VoteTypePrecommit,
		BlockID:          types.BlockID{Hash: header.Hash()},
	}
	// Sign it
	signBytes := types.SignBytes(header.ChainID, vote)
	vote.Signature = key.Sign(signBytes)
	return vote
}

func GenHeader(chainID string, height int, txs types.Txs,
	vals *types.ValidatorSet, appHash []byte) *types.Header {

	return &types.Header{
		ChainID: chainID,
		Height:  height,
		Time:    time.Now(),
		NumTxs:  len(txs),
		// LastBlockID
		// LastCommitHash
		ValidatorsHash: vals.Hash(),
		DataHash:       txs.Hash(),
		AppHash:        appHash,
	}
}

// GenCheckpoint calls GenHeader and SignHeader and combines them into a Checkpoint
func (v ValKeys) GenCheckpoint(chainID string, height int, txs types.Txs,
	vals *types.ValidatorSet, appHash []byte, first, last int) lc.Checkpoint {

	header := GenHeader(chainID, height, txs, vals, appHash)
	check := lc.Checkpoint{
		Header: header,
		Commit: v.SignHeader(header, first, last),
	}
	return check
}
