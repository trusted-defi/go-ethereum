package engine

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	trustedv1 "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/big"
)

type TrustedEngineClient struct {
	client trustedv1.TrustedServiceClient
}

func NewTrustedEngineClient() *TrustedEngineClient {
	c := new(TrustedEngineClient)
	client, err := grpc.Dial("127.0.0.1:38000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("netserver connect failed", "err", err)
	}

	c.client = trustedv1.NewTrustedServiceClient(client)
	return c
}

func (t *TrustedEngineClient) SetPrice(price *big.Int) {
	return
}

func (t *TrustedEngineClient) GasPrice() *big.Int {
	gas, _ := new(big.Int).SetString("1000000000", 10)
	return gas
}

func (t *TrustedEngineClient) Nonce(addr common.Address) uint64 {

	return 0
}

func (t *TrustedEngineClient) Stat() (int, int) {
	return 0, 0
}

func (t *TrustedEngineClient) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pending := make(map[common.Address]types.Transactions)
	queue := make(map[common.Address]types.Transactions)
	return pending, queue
}

func (t *TrustedEngineClient) ContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	pending := make(types.Transactions, 0)
	queue := make(types.Transactions, 0)
	return pending, queue
}

func (t *TrustedEngineClient) Pending() map[common.Address]types.Transactions {
	pending := make(map[common.Address]types.Transactions)
	return pending
}

func (t *TrustedEngineClient) Locals() []common.Address {
	l := make([]common.Address, 0)
	return l
}

func (t *TrustedEngineClient) AddLocals(txs []*types.Transaction) []error {
	err := make([]error, 0)
	return err
}

func (t *TrustedEngineClient) AddRemotes(txs []*types.Transaction) []error {
	err := make([]error, 0)
	return err
}

func (t *TrustedEngineClient) Status(hashes []common.Hash) []uint {
	status := make([]uint, len(hashes))
	return status
}

func (t *TrustedEngineClient) Get(hash common.Hash) *types.Transaction {
	tx := new(types.Transaction)
	return tx
}

func (t *TrustedEngineClient) Has(hash common.Hash) bool {
	return false
}

func (t *TrustedEngineClient) Crypt(data []byte) ([]byte, error) {
	req := new(trustedv1.CryptRequest)
	req.Data = common.CopyBytes(data)
	req.Method = 1
	res, err := t.client.Crypt(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.Crypted, nil
}

type SendTrustedTransacionResult struct {
	Hash  common.Hash   `json:"hash"`  // transaction hash
	Asset hexutil.Bytes `json:"asset"` // verification for tx add by trusted engine
}

func (t *TrustedEngineClient) AddLocalTrustedTx(txdata []byte) (*SendTrustedTransacionResult, error) {
	req := new(trustedv1.AddTrustedTxRequest)
	req.CtyptedTx = common.CopyBytes(txdata)

	res, err := t.client.AddLocalTrustedTx(context.Background(), req)
	if err != nil {
		return nil, err
	}
	result := new(SendTrustedTransacionResult)
	result.Asset = common.CopyBytes(res.Asset)
	result.Hash = common.BytesToHash(res.Hash)
	return result, nil
}

func (t *TrustedEngineClient) AddRemoteTrustedTx(txdata []byte) (*SendTrustedTransacionResult, error) {
	req := new(trustedv1.AddTrustedTxRequest)
	req.CtyptedTx = common.CopyBytes(txdata)

	res, err := t.client.AddLocalTrustedTx(context.Background(), req)
	if err != nil {
		return nil, err
	}
	result := new(SendTrustedTransacionResult)
	result.Asset = common.CopyBytes(res.Asset)
	result.Hash = common.BytesToHash(res.Hash)
	return result, nil
}

type TrustedPool interface {
	SetPrice(price *big.Int)
	GasPrice() *big.Int
	Nonce(addr common.Address) uint64
	Stat() (pendingCount int, queueCount int)
	Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions)
	ContentFrom(addr common.Address) (types.Transactions, types.Transactions)
	Pending() map[common.Address]types.Transactions
	Locals() []common.Address
	AddLocals(txs []*types.Transaction) []error
	AddRemotes(txs []*types.Transaction) []error
	Status(hashes []common.Hash) []uint
	Get(hash common.Hash) *types.Transaction
	Has(hash common.Hash) bool
}
