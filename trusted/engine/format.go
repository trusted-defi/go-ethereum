package engine

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	trustedv1 "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
)

func parseToTransaction(txdata []byte) *types.Transaction {
	tx := new(types.Transaction)
	tx.UnmarshalBinary(txdata)
	return tx
}

func parseHashesToBytes(hashs []common.Hash) [][]byte {
	data := make([][]byte, 0, len(hashs))
	for _, hash := range hashs {
		data = append(data, hash.Bytes())
	}
	return data
}

func parseTxsToList(txs []*types.Transaction) *trustedv1.TransactionList {
	list := new(trustedv1.TransactionList)
	list.Txs = make([][]byte, len(txs))
	for i, tx := range txs {
		list.Txs[i], _ = tx.MarshalBinary()
	}
	return list
}

func parseAccountTransactionsToMap(accountsTx []*trustedv1.AccountTransactionList) map[common.Address]types.Transactions {
	maptx := make(map[common.Address]types.Transactions)
	for _, accountTxlist := range accountsTx {
		address := common.BytesToAddress(accountTxlist.Address)
		tlist := make(types.Transactions, 0, len(accountTxlist.TxList.Txs))
		for _, tx := range accountTxlist.TxList.Txs {
			ntx := new(types.Transaction)
			err := ntx.UnmarshalBinary(tx)
			if err != nil {
				continue
			}
			tlist = append(tlist, ntx)
		}
		maptx[address] = tlist
	}
	return maptx
}

func parseAccountTransactionsToList(accountsTx []*trustedv1.AccountTransactionList) types.Transactions {
	txs := make(types.Transactions, 0)
	if len(accountsTx) > 0 {
		accountTxlist := accountsTx[0]
		txs = make(types.Transactions, 0, len(accountTxlist.TxList.Txs))
		for _, tx := range accountTxlist.TxList.Txs {
			ntx := new(types.Transaction)
			err := ntx.UnmarshalBinary(tx)
			if err != nil {
				continue
			}
			txs = append(txs, ntx)
		}
	}
	return txs
}
