package engine

import (
	"context"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/trusted/config"
	trustedv1 "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
	"github.com/ethereum/go-ethereum/trusted/trustedtype"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"math/big"
)

var (
	ErrInvalidRemoteReport = errors.New("invalid remote report")
	ErrGetRemoteReport     = errors.New("get remote report failed")
	ErrVerifyFailed        = errors.New("verify failed")
	ErrClientNotReady      = errors.New("trust engine not ready")
)

type TrustedEngineClient struct {
	client trustedv1.TrustedServiceClient
}

func NewTrustedEngineClient() *TrustedEngineClient {
	teconfig := config.GetConfig()
	c := new(TrustedEngineClient)
	//client, err := grpc.Dial(":3802", grpc.WithTransportCredentials(insecure.NewCredentials()))
	client, err := grpc.Dial(teconfig.TrustedClient, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("netserver connect failed", "err", err)
	}

	c.client = trustedv1.NewTrustedServiceClient(client)
	return c
}

func (t *TrustedEngineClient) Ready() bool {
	_, err := t.client.PoolGasPrice(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Error("gasprice failed", "err", err)
		return false
	}
	return true
}

func (t *TrustedEngineClient) SetPrice(price *big.Int) {
	req := new(trustedv1.SetPriceRequest)
	req.Price = price.Bytes()
	t.client.PoolSetPrice(context.Background(), req)
}

func (t *TrustedEngineClient) GasPrice() *big.Int {
	res, err := t.client.PoolGasPrice(context.Background(), &emptypb.Empty{})
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
	res, err := t.client.PoolStat(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Info("trusted stat", "err", err)
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
	res, err := t.client.PoolPending(context.Background(), &emptypb.Empty{})
	if err != nil {
		return pending
	}
	pending = parseAccountTransactionsToMap(res.PendingList)
	return pending
}

func (t *TrustedEngineClient) Locals() []common.Address {
	l := make([]common.Address, 0)
	res, err := t.client.PoolLocals(context.Background(), &emptypb.Empty{})
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
	for i, err := range res.Errors {
		if len(err) > 0 {
			errs[i] = errors.New(err)
		}
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

func (t *TrustedEngineClient) SubscribeNewTx(ch chan []trustedtype.TrustedCryptTx) error {
	server, err := t.client.SubscribeNewTransaction(context.Background(), new(trustedv1.SubscribeNewTxRequest))
	if err != nil {
		return err
	}
	var msg *trustedv1.SubscribeNewTxResponse
	bcontinue := true
	for bcontinue {
		msg, err = server.Recv()
		if err != nil {
			bcontinue = false
			break
		}
		ntxs := make([]trustedtype.TrustedCryptTx, len(msg.CryptedNewTx))
		for i, tx := range msg.CryptedNewTx {
			ntxs[i] = common.CopyBytes(tx)
		}
		ch <- ntxs
	}
	return err
}

func (t *TrustedEngineClient) Crypt(data []byte) ([]byte, error) {
	req := new(trustedv1.CryptRequest)
	req.Data = common.CopyBytes(data)
	req.Method = 1
	log.Debug("before crypt", "tx", common.Bytes2Hex(data))
	res, err := t.client.Crypt(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Debug("after crypt", "tx", common.Bytes2Hex(res.Crypted))
	return res.Crypted, nil
}

type SendTrustedTransacionResult struct {
	Hash   common.Hash   `json:"hash"`   // transaction hash
	Report hexutil.Bytes `json:"report"` // verification for tx add by trusted engine
	Error  error         `json:"-"`
}

func (t *TrustedEngineClient) AddLocalTrustedTx(tx trustedtype.TrustedCryptTx) (*SendTrustedTransacionResult, error) {
	req := new(trustedv1.AddTrustedTxsRequest)
	req.CtyptedTxs = parseTrustedTxsToList([]trustedtype.TrustedCryptTx{tx})

	//log.Debug("add local trusted tx", "tx", common.Bytes2Hex(txdata))
	resp, err := t.client.AddLocalTrustedTxs(context.Background(), req)
	if err != nil {
		return nil, err
	}
	res := new(SendTrustedTransacionResult)
	result := resp.Results[0]
	if len(result.Error) != 0 {
		res.Error = errors.New(result.Error)
		return res, errors.New(result.Error)
	}
	res.Report = common.CopyBytes(result.Asset)
	res.Hash = common.BytesToHash(result.Hash)
	res.Error = nil
	log.Debug("add local trusted tx", "txhash", res.Hash)
	return res, nil
}

func (t *TrustedEngineClient) AddRemoteTrustedTx(txs []trustedtype.TrustedCryptTx) ([]*SendTrustedTransacionResult, error) {
	req := new(trustedv1.AddTrustedTxsRequest)
	req.CtyptedTxs = parseTrustedTxsToList(txs)
	res := make([]*SendTrustedTransacionResult, len(txs))

	resp, err := t.client.AddRemoteTrustedTxs(context.Background(), req)
	if err != nil {
		return res, err
	}
	for i, response := range resp.Results {
		result := new(SendTrustedTransacionResult)
		if len(response.Error) > 0 {
			result.Error = errors.New(response.Error)
		} else {
			result.Report = common.CopyBytes(response.Asset)
			result.Hash = common.BytesToHash(response.Hash)
			result.Error = nil
		}
		res[i] = result
	}
	return res, nil
}

// CheckSecretKey check secretkey already exist or not.
func (t *TrustedEngineClient) CheckSecretKey() (bool, error) {
	res, err := t.client.CheckSecretKey(context.Background(), &emptypb.Empty{})
	if err != nil {
		return false, err
	}
	return res.Exist, nil
}

// GetAuthData generate a remote report at begin of a auth-verify process.
func (t *TrustedEngineClient) GetAuthData(peerId string) ([]byte, error) {
	req := new(trustedv1.GetAuthDataRequest)
	req.PeerId = peerId
	res, err := t.client.GetAuthData(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.AuthData, err
}

// VerifyAuth verify auth data received from remote peer
func (t *TrustedEngineClient) VerifyAuth(authData []byte, peerId string) error {
	req := new(trustedv1.VerifyAuthRequest)
	req.PeerId = peerId
	req.AuthData = authData
	res, err := t.client.VerifyAuth(context.Background(), req)
	if err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}
	return nil
}

// GetVerifyData generate a remote report used to verify remote peer..
func (t *TrustedEngineClient) GetVerifyData(peerId string) ([]byte, error) {
	req := new(trustedv1.GetVerifyDataRequest)
	req.PeerId = peerId
	res, err := t.client.GetVerifyData(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.VerifyData, err
}

// VerifyRemoteVerify verify remote verify-data received from remote peer..
func (t *TrustedEngineClient) VerifyRemoteVerify(verifyData []byte, peerId string) error {
	req := new(trustedv1.VerifyRemoteVerifyRequest)
	req.PeerId = peerId
	req.VerifyData = verifyData
	res, err := t.client.VerifyRemoteVerify(context.Background(), req)
	if err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}
	return nil
}

// GetRequestKeyData generate a remote report used to request secret key.
func (t *TrustedEngineClient) GetRequestKeyData(peerId string) ([]byte, error) {
	req := new(trustedv1.GetRequestKeyDataRequest)
	req.PeerId = peerId
	res, err := t.client.GetRequestKeyData(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.RequestKeyData, err
}

// VerifyRequestKeyData verify remote verify-data received from remote peer..
func (t *TrustedEngineClient) VerifyRequestKeyData(request []byte, peerId string) error {
	req := new(trustedv1.VerifyRequestKeyDataRequest)
	req.PeerId = peerId
	req.RequestKeyData = request
	res, err := t.client.VerifyRequestKeyData(context.Background(), req)
	if err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}
	return nil
}

// GetResponseKeyData generate a remote report used to request secret key.
func (t *TrustedEngineClient) GetResponseKeyData(peerId string) ([]byte, error) {
	req := new(trustedv1.GetResponseKeyDataRequest)
	req.PeerId = peerId
	res, err := t.client.GetResponseKeyData(context.Background(), req)
	if err != nil {
		return nil, err
	}
	return res.ResponseKeyData, err
}

// VerifyResponseKey verify remote verify-data received from remote peer..
func (t *TrustedEngineClient) VerifyResponseKey(response []byte, peerId string) error {
	req := new(trustedv1.VerifyResponseKeyRequest)
	req.PeerId = peerId
	req.ResponseKeyData = response
	res, err := t.client.VerifyResponseKey(context.Background(), req)
	if err != nil {
		return err
	}
	if len(res.Error) > 0 {
		return errors.New(res.Error)
	}
	return nil
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
