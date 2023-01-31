package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trusted/engine"
	"math/big"
	"sync"
)

type TxPool struct {
	client *engine.TrustedEngineClient
	txFeed event.Feed
	scope  event.SubscriptionScope
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

func NewTxPool(config TxPoolConfig, chainconfig *params.ChainConfig, chain blockChain) *TxPool {
	pool := &TxPool{
		client: engine.NewTrustedEngineClient(),
	}
	return pool
}

// Stop terminates the transaction pool.
func (pool *TxPool) Stop() {
	// Unsubscribe all subscriptions registered from txpool
	pool.scope.Close()

	pool.wg.Wait()

	log.Info("Transaction pool stopped")
}

// SubscribeNewTxsEvent registers a subscription of NewTxsEvent and
// starts sending event to the given channel.
func (pool *TxPool) SubscribeNewTxsEvent(ch chan<- NewTxsEvent) event.Subscription {
	return pool.scope.Track(pool.txFeed.Subscribe(ch))
}

// GasPrice returns the current gas price enforced by the transaction pool.
func (pool *TxPool) GasPrice() *big.Int {
	return pool.client.GasPrice()
}

// SetGasPrice updates the minimum price required by the transaction pool for a
// new transaction, and drops all transactions below this threshold.
func (pool *TxPool) SetGasPrice(price *big.Int) {
	pool.client.SetPrice(price)

	log.Info("Transaction pool price threshold updated", "price", price)
}

// Nonce returns the next nonce of an account, with all transactions executable
// by the pool already applied on top.
func (pool *TxPool) Nonce(addr common.Address) uint64 {
	return pool.client.Nonce(addr)
}

// Stats retrieves the current pool stats, namely the number of pending and the
// number of queued (non-executable) transactions.
func (pool *TxPool) Stats() (int, int) {
	return pool.client.Stat()
}

// Content retrieves the data content of the transaction pool, returning all the
// pending as well as queued transactions, grouped by account and sorted by nonce.
func (pool *TxPool) Content() (map[common.Address]types.Transactions, map[common.Address]types.Transactions) {
	pending, queued := pool.client.Content()
	return pending, queued
}

// ContentFrom retrieves the data content of the transaction pool, returning the
// pending as well as queued transactions of this address, grouped by nonce.
func (pool *TxPool) ContentFrom(addr common.Address) (types.Transactions, types.Transactions) {
	pending, queued := pool.client.ContentFrom(addr)
	return pending, queued
}

// Pending retrieves all currently processable transactions, grouped by origin
// account and sorted by nonce. The returned transaction set is a copy and can be
// freely modified by calling code.
//
// The enforceTips parameter can be used to do an extra filtering on the pending
// transactions and only return those whose **effective** tip is large enough in
// the next pending execution environment.
func (pool *TxPool) Pending(enforceTips bool) map[common.Address]types.Transactions {
	return pool.client.Pending()
}

// Locals retrieves the accounts currently considered local by the pool.
func (pool *TxPool) Locals() []common.Address {
	return pool.client.Locals()
}

// AddLocals enqueues a batch of transactions into the pool if they are valid, marking the
// senders as a local ones, ensuring they go around the local pricing constraints.
//
// This method is used to add transactions from the RPC API and performs synchronous pool
// reorganization and event propagation.
func (pool *TxPool) AddLocals(txs []*types.Transaction) []error {
	return pool.client.AddLocals(txs)
}

// AddLocal enqueues a single local transaction into the pool if it is valid. This is
// a convenience wrapper aroundd AddLocals.
func (pool *TxPool) AddLocal(tx *types.Transaction) error {
	errs := pool.AddLocals([]*types.Transaction{tx})
	return errs[0]
}

// AddRemotes enqueues a batch of transactions into the pool if they are valid. If the
// senders are not among the locally tracked ones, full pricing constraints will apply.
//
// This method is used to add transactions from the p2p network and does not wait for pool
// reorganization and internal event propagation.
func (pool *TxPool) AddRemotes(txs []*types.Transaction) []error {
	return pool.client.AddRemotes(txs)
}

// This is like AddRemotes, but waits for pool reorganization. Tests use this method.
func (pool *TxPool) AddRemotesSync(txs []*types.Transaction) []error {
	return pool.client.AddRemotes(txs)
}

// AddRemote enqueues a single transaction into the pool if it is valid. This is a convenience
// wrapper around AddRemotes.
//
// Deprecated: use AddRemotes
func (pool *TxPool) AddRemote(tx *types.Transaction) error {
	errs := pool.AddRemotes([]*types.Transaction{tx})
	return errs[0]
}

// Status returns the status (unknown/pending/queued) of a batch of transactions
// identified by their hashes.
func (pool *TxPool) Status(hashes []common.Hash) []TxStatus {

	status := make([]TxStatus, len(hashes))
	pstatus := pool.client.Status(hashes)

	for i, s := range pstatus {
		status[i] = TxStatus(s)
	}
	return status
}

// Get returns a transaction if it is contained in the pool and nil otherwise.
func (pool *TxPool) Get(hash common.Hash) *types.Transaction {
	return pool.client.Get(hash)
}

// Has returns an indicator whether txpool has a transaction cached with the
// given hash.
func (pool *TxPool) Has(hash common.Hash) bool {
	return pool.client.Has(hash)
}
