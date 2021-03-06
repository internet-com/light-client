package certifiers_test

import (
	"fmt"
	"testing"

	"github.com/tendermint/light-client/certifiers"
)

func BenchmarkGenCheckpoint20(b *testing.B) {
	keys := certifiers.GenValKeys(20)
	benchmarkGenCheckpoint(b, keys)
}

func BenchmarkGenCheckpoint100(b *testing.B) {
	keys := certifiers.GenValKeys(100)
	benchmarkGenCheckpoint(b, keys)
}

func BenchmarkGenCheckpointSec20(b *testing.B) {
	keys := certifiers.GenSecValKeys(20)
	benchmarkGenCheckpoint(b, keys)
}

func BenchmarkGenCheckpointSec100(b *testing.B) {
	keys := certifiers.GenSecValKeys(100)
	benchmarkGenCheckpoint(b, keys)
}

func benchmarkGenCheckpoint(b *testing.B, keys certifiers.ValKeys) {
	chainID := fmt.Sprintf("bench-%d", len(keys))
	vals := keys.ToValidators(20, 10)
	for i := 0; i < b.N; i++ {
		h := 1 + i
		appHash := []byte(fmt.Sprintf("h=%d", h))
		keys.GenCheckpoint(chainID, h, nil, vals, appHash, 0, len(keys))
	}
}

// this benchmarks generating one key
func BenchmarkGenValKeys(b *testing.B) {
	keys := certifiers.GenValKeys(20)
	for i := 0; i < b.N; i++ {
		keys = keys.Extend(1)
	}
}

// this benchmarks generating one key
func BenchmarkGenSecValKeys(b *testing.B) {
	keys := certifiers.GenSecValKeys(20)
	for i := 0; i < b.N; i++ {
		keys = keys.Extend(1)
	}
}

func BenchmarkToValidators20(b *testing.B) {
	benchmarkToValidators(b, 20)
}

func BenchmarkToValidators100(b *testing.B) {
	benchmarkToValidators(b, 100)
}

// this benchmarks constructing the validator set (.PubKey() * nodes)
func benchmarkToValidators(b *testing.B, nodes int) {
	keys := certifiers.GenValKeys(nodes)
	for i := 1; i <= b.N; i++ {
		keys.ToValidators(int64(2*i), int64(i))
	}
}

func BenchmarkToValidatorsSec100(b *testing.B) {
	benchmarkToValidatorsSec(b, 100)
}

// this benchmarks constructing the validator set (.PubKey() * nodes)
func benchmarkToValidatorsSec(b *testing.B, nodes int) {
	keys := certifiers.GenSecValKeys(nodes)
	for i := 1; i <= b.N; i++ {
		keys.ToValidators(int64(2*i), int64(i))
	}
}

func BenchmarkCertifyCheckpoint20(b *testing.B) {
	keys := certifiers.GenValKeys(20)
	benchmarkCertifyCheckpoint(b, keys)
}

func BenchmarkCertifyCheckpoint100(b *testing.B) {
	keys := certifiers.GenValKeys(100)
	benchmarkCertifyCheckpoint(b, keys)
}

func BenchmarkCertifyCheckpointSec20(b *testing.B) {
	keys := certifiers.GenSecValKeys(20)
	benchmarkCertifyCheckpoint(b, keys)
}

func BenchmarkCertifyCheckpointSec100(b *testing.B) {
	keys := certifiers.GenSecValKeys(100)
	benchmarkCertifyCheckpoint(b, keys)
}

func benchmarkCertifyCheckpoint(b *testing.B, keys certifiers.ValKeys) {
	chainID := "bench-certify"
	vals := keys.ToValidators(20, 10)
	cert := certifiers.NewStatic(chainID, vals)
	check := keys.GenCheckpoint(chainID, 123, nil, vals, []byte("foo"), 0, len(keys))
	for i := 0; i < b.N; i++ {
		err := cert.Certify(check)
		if err != nil {
			panic(err)
		}
	}

}
