package trustedtype

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

type TrustedConfig struct {
	ChainServer   string
	TrustedClient string
}

// todo implement rlp
type TrustedCryptTx []byte

func (t TrustedCryptTx) Hash() common.Hash {
	h := sha3.Sum256(t)
	return common.BytesToHash(h[:])
}

func (t TrustedCryptTx) Size() int64 {
	return int64(len(t))
}

func (t TrustedCryptTx) Copy() TrustedCryptTx {
	n := make(TrustedCryptTx, len(t))
	copy(n, t)
	return n
}
