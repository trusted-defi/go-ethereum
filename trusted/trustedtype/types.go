package trustedtype

import (
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

type TrustedCryptTx []byte

func (t TrustedCryptTx) Hash() common.Hash {
	h := sha3.Sum256(t)
	return common.BytesToHash(h[:])
}
