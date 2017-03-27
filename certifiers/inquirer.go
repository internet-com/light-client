package certifiers

import (
	lc "github.com/tendermint/light-client"
	"github.com/tendermint/tendermint/types"
)

type InquiringCertifier struct {
	Cert *DynamicCertifier
	Provider
}

func NewInquiring(chainID string, vals []*types.Validator, provider Provider) *InquiringCertifier {
	return &InquiringCertifier{
		Cert:     NewDynamic(chainID, vals),
		Provider: provider,
	}
}

func (c *InquiringCertifier) Certify(check lc.Checkpoint) error {
	err := c.Cert.Certify(check)
	if !ValidatorsChanged(err) {
		return err
	}
	err = c.updateToHash(check.Header.ValidatorsHash)
	if err != nil {
		return err
	}
	return c.Cert.Certify(check)
}

func (c *InquiringCertifier) Update(check lc.Checkpoint, vals []*types.Validator) error {
	err := c.Cert.Update(check, vals)
	if err == nil {
		c.StoreSeed(Seed{Checkpoint: check, Validators: vals})
	}
	return err
}

// updateToHash gets the validator hash we want to update to
// if TooMuchChange, we try to find a path by binary search over height
func (c *InquiringCertifier) updateToHash(vhash []byte) error {
	// try to get the match, and update
	seed, err := c.GetByHash(vhash)
	if err != nil {
		return err
	}
	err = c.Cert.Update(seed.Checkpoint, seed.Validators)
	// handle TooMuchChange by using divide and conquer
	if TooMuchChange(err) {
		err = c.updateToHeight(seed.Height())
	}
	return err
}

// updateToHeight will use divide-and-conquer to find a path to h
func (c *InquiringCertifier) updateToHeight(h int) error {
	// try to update to this height (with checks)
	seed, err := c.GetByHeight(h)
	if err != nil {
		return err
	}
	start, end := c.Cert.LastHeight, seed.Height()
	if end <= start {
		return ErrNoPathFound()
	}
	err = c.Update(seed.Checkpoint, seed.Validators)

	// we can handle TooMuchChange specially
	if !TooMuchChange(err) {
		return err
	}

	// try to update to mid
	mid := (start + end) / 2
	err = c.updateToHeight(mid)
	if err != nil {
		return err
	}

	// if we made it to mid, we recurse
	return c.updateToHeight(h)
}
