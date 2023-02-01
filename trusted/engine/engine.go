package engine

import (
	"context"
	"errors"
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
	req := new(trustedv1.SetPriceRequest)
	req.Price = price.Bytes()
	t.client.PoolSetPrice(context.Background(), req)
}

func (t *TrustedEngineClient) GasPrice() *big.Int {
	res, err := t.client.PoolGasPrice(context.Background(), nil)
	if err != nil {
		gas, _ := new(big.Int).SetString("1000000000", 10)
		return gas
	}
	return new(big.Int).SetBytes(res.Price)
}

func (t *TrustedEngineClient) Nonce(addr common.Address) uint64 {
	req := new(trustedv1.PendingNonceRequest)
	req.Address = addr.Bytes()
	res, err := t.client.PendingNonce(context.Background(), req)
	if err != nil {
		return 0
	}
	return res.Nonce
}

func (t *TrustedEngineClient) Stat() (int, int) {
	res, err := t.client.PoolStat(context.Background(), nil)
	if err != nil {
		return 0, 0
	}
	return int(res.Pending), int(res.Queue)
}

func (t *TrustedEngineClient) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pending := make(map[common.Address]types.Transactions)
	queue := make(map[common.Address]types.Transactions)

	req := new(trustedv1.PoolContentRequest)
	res, err := t.client.PoolContent(context.Background(), req)
	if err != nil {
		return pending, queue
	}
	pending = parseAccountTransactionsToMap(res.PendingList)
	queue = parseAccountTransactionsToMap(res.QueueList)
	return pending, queue
}

func (t *TrustedEngineClient) ContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	pending := make(types.Transactions, 0)
	queue := make(types.Transactions, 0)
	req := new(trustedv1.PoolContentRequest)
	req.Address = addr.Bytes()
	res, err := t.client.PoolContentFrom(context.Background(), req)
	if err != nil {
		return pending, queue
	}
	pending = parseAccountTransactionsToList(res.PendingList)
	queue = parseAccountTransactionsToList(res.QueueList)
	return pending, queue
}

func (t *TrustedEngineClient) Pending() map[common.Address]types.Transactions {
	pending := make(map[common.Address]types.Transactions)
	res, err := t.client.PoolPending(context.Background(), nil)
	if err != nil {
		return pending
	}
	pending = parseAccountTransactionsToMap(res.PendingList)
	return pending
}

func (t *TrustedEngineClient) Locals() []common.Address {
	l := make([]common.Address, 0)
	res, err := t.client.PoolLocals(context.Background(), nil)
	if err != nil {
		return l
	}
	for _, addr := range res.AddressList {
		a := common.BytesToAddress(addr)
		l = append(l, a)
	}
	return l
}

func (t *TrustedEngineClient) AddLocals(txs []*types.Transaction) []error {
	errs := make([]error, len(txs))
	req := new(trustedv1.AddTxsRequest)
	req.TxList = parseTxsToList(txs)
	res, err := t.client.AddLocalsTx(context.Background(), req)
	if err != nil {
		for i := 0; i < len(errs); i++ {
			errs[i] = err
		}
		return errs
	}
	for i := 0; i < len(errs); i++ {
		errs[i] = errors.New(res.Errors[i])
	}
	return errs
}

func (t *TrustedEngineClient) AddRemotes(txs []*types.Transaction) []error {
	errs := make([]error, len(txs))
	req := new(trustedv1.AddTxsRequest)
	req.TxList = parseTxsToList(txs)
	res, err := t.client.AddRemoteTx(context.Background(), req)
	if err != nil {
		for i := 0; i < len(errs); i++ {
			errs[i] = err
		}
		return errs
	}
	for i := 0; i < len(errs); i++ {
		errs[i] = errors.New(res.Errors[i])
	}
	return errs
}

func (t *TrustedEngineClient) Status(hashes []common.Hash) []uint {
	req := new(trustedv1.TxStatusRequest)
	req.TxHashs = parseHashesToBytes(hashes)
	status := make([]uint, len(hashes))
	res, err := t.client.TxStatus(context.Background(), req)
	if err != nil {
		return status
	}
	for i, s := range res.TxStatus {
		status[i] = uint(s)
	}
	return status
}

func (t *TrustedEngineClient) Get(hash common.Hash) *types.Transaction {
	req := new(trustedv1.TxGetRequest)
	req.TxHash = hash.Bytes()
	res, err := t.client.TxGet(context.Background(), req)
	if err != nil {
		return nil
	}
	return parseToTransaction(res.Tx)
}

func (t *TrustedEngineClient) Has(hash common.Hash) bool {
	req := new(trustedv1.TxHasRequest)
	req.TxHash = hash.Bytes()
	res, err := t.client.TxHas(context.Background(), req)
	if err != nil {
		return false
	}
	return res.Has
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
