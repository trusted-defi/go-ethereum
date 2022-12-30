package txpoolclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	trustedv1 "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math/big"
)

type TxPoolClient struct {
	client trustedv1.TrustedServiceClient
}

func NewTxPoolClient() *TxPoolClient {
	c := new(TxPoolClient)
	client, err := grpc.Dial("127.0.0.1:38000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("netserver connect failed", "err", err)
	}

	c.client = trustedv1.NewTrustedServiceClient(client)
	return c
}

func (t *TxPoolClient) SetPrice(price *big.Int) {
	return
}

func (t *TxPoolClient) GasPrice() *big.Int {
	gas, _ := new(big.Int).SetString("1000000000", 10)
	return gas
}

func (t *TxPoolClient) Nonce(addr common.Address) uint64 {

	return 0
}

func (t *TxPoolClient) Stat() (int, int) {
	return 0, 0
}

func (t *TxPoolClient) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pending := make(map[common.Address]types.Transactions)
	queue := make(map[common.Address]types.Transactions)
	return pending, queue
}

func (t *TxPoolClient) ContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	pending := make(types.Transactions, 0)
	queue := make(types.Transactions, 0)
	return pending, queue
}

func (t *TxPoolClient) Pending() map[common.Address]types.Transactions {
	pending := make(map[common.Address]types.Transactions)
	return pending
}

func (t *TxPoolClient) Locals() []common.Address {
	l := make([]common.Address, 0)
	return l
}

func (t *TxPoolClient) AddLocals(txs []*types.Transaction) []error {
	err := make([]error, 0)
	return err
}

func (t *TxPoolClient) AddRemotes(txs []*types.Transaction) []error {
	err := make([]error, 0)
	return err
}

func (t *TxPoolClient) Status(hashes []common.Hash) []uint {
	status := make([]uint, len(hashes))
	return status
}

func (t *TxPoolClient) Get(hash common.Hash) *types.Transaction {
	tx := new(types.Transaction)
	return tx
}

func (t *TxPoolClient) Has(hash common.Hash) bool {
	return false
}
