package proofs

import (
	"github.com/pkg/errors"
	data "github.com/tendermint/go-data"
	merkle "github.com/tendermint/go-merkle"
	wire "github.com/tendermint/go-wire"
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/tendermint/rpc/client"
)

var _ lc.Prover = AppProver{}
var _ lc.Proof = AppProof{}

// we limit proofs to 1MB to stop overflow attacks
const appLimit = 1000 * 1000

// AppProver provides positive proofs of key-value pairs in the abciapp.
//
// TODO: also support negative proofs (this key is not set)
type AppProver struct {
	node client.Client
}

func NewAppProver(node client.Client) AppProver {
	return AppProver{node: node}
}

// Get tries to download a merkle hash for app state on this key from
// the tendermint node.
func (a AppProver) Get(key []byte, h uint64) (lc.Proof, error) {
	res, err := a.node.ABCIQuery("/key", key, true)
	if err != nil {
		return nil, err
	}

	// make sure the proof is the proper height
	resp := res.Response
	if !resp.Code.IsOK() {
		return nil, errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
	}
	if len(resp.Key) == 0 || len(resp.Value) == 0 || len(resp.Proof) == 0 {
		return nil, errors.New("Missing information in query response")
	}
	if h != 0 && h != resp.Height {
		return nil, errors.Errorf("Requested height %d, received %d", h, resp.Height)
	}
	proof := AppProof{
		Height: resp.Height,
		Key:    resp.Key,
		Value:  resp.Value,
		Proof:  resp.Proof,
	}
	return proof, nil
}

func (a AppProver) Unmarshal(data []byte) (pr lc.Proof, err error) {
	// to handle go-wire panics... ugh
	defer func() {
		if rec := recover(); rec != nil {
			if e, ok := rec.(error); ok {
				err = errors.WithStack(e)
			} else {
				err = errors.Errorf("Panic: %v", rec)
			}
		}
	}()
	var proof AppProof
	err = errors.WithStack(wire.ReadBinaryBytes(data, &proof))
	return proof, err
}

type AppProof struct {
	Height uint64
	Key    data.Bytes
	Value  data.Bytes
	Proof  data.Bytes
}

func (p AppProof) BlockHeight() uint64 {
	return p.Height
}

func (p AppProof) Validate(check lc.Checkpoint) error {
	if uint64(check.Height()) != p.Height {
		return errors.Errorf("Trying to validate proof for block %d with header for block %d",
			p.Height, check.Height())
	}

	proof, err := merkle.ReadProof(p.Proof)
	if err != nil {
		return errors.WithStack(err)
	}

	if !proof.Verify(p.Key, p.Value, check.Header.AppHash) {
		return errors.Errorf("Didn't validate against hash %X", check.Header.AppHash)
	}

	// LGTM!
	return nil
}

func (p AppProof) Marshal() ([]byte, error) {
	data := wire.BinaryBytes(p)
	return data, nil
}