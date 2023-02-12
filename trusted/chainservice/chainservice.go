package chainservice

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	trusted "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/big"
	"net"
)

type ChainService struct {
	chain *core.BlockChain
	trusted.UnimplementedChainServiceServer
}

func (s *ChainService) GetBlock(ctx context.Context, req *trusted.BlockRequest) (*trusted.BlockResponse, error) {
	hash := common.BytesToHash(req.BlockHash)
	block := s.chain.GetBlock(hash, req.BlockNum)
	buffer := bytes.NewBuffer(make([]byte, 0))
	if block == nil {
		return nil, status.Errorf(codes.NotFound, "block not found")
	}
	err := block.EncodeRLP(buffer)
	if err != nil {
		log.Error("chain service ", "encode rlp failed with err", err)
		return nil, status.Errorf(codes.Unknown, "encode rlp failed")
	}
	res := new(trusted.BlockResponse)
	res.BlockData = buffer.Bytes()
	return res, nil
}
func (s *ChainService) GetBalance(ctx context.Context, req *trusted.BalanceRequest) (*trusted.BalanceResponse, error) {
	var db *state.StateDB
	var blockheight *big.Int
	var err error
	var res = new(trusted.BalanceResponse)
	addr := common.BytesToAddress(req.Address)
	if len(req.BlockNum) > 0 {
		blockheight = new(big.Int).SetBytes(req.BlockNum)
	}
	if blockheight == nil {
		// get latest state
		db, err = s.chain.State()
		if err != nil {
			return nil, err
		}
	} else {
		header := s.chain.GetHeaderByNumber(blockheight.Uint64())
		if header == nil {
			return nil, errors.New("block not found")
		}
		db, err = s.chain.StateAt(header.Root)
		if err != nil {
			return nil, err
		}
	}
	balance := db.GetBalance(addr)
	res.Balance = balance.Bytes()

	return res, nil
}
func (s *ChainService) GetNonce(ctx context.Context, req *trusted.NonceRequest) (*trusted.NonceResponse, error) {
	var db *state.StateDB
	var blockheight *big.Int
	var err error
	var res = new(trusted.NonceResponse)
	addr := common.BytesToAddress(req.Address)
	if len(req.BlockNum) > 0 {
		blockheight = new(big.Int).SetBytes(req.BlockNum)
	}
	if blockheight == nil {
		// get latest state
		db, err = s.chain.State()
		if err != nil {
			return nil, err
		}
	} else {
		header := s.chain.GetHeaderByNumber(blockheight.Uint64())
		if header == nil {
			return nil, errors.New("block not found")
		}
		db, err = s.chain.StateAt(header.Root)
		if err != nil {
			return nil, err
		}
	}
	res.Nonce = db.GetNonce(addr)

	return res, nil
}
func (s *ChainService) CurrentBlock(ctx context.Context, req *trusted.CurrentBlockRequest) (*trusted.CurrentBlockResponse, error) {
	res := new(trusted.CurrentBlockResponse)
	block := s.chain.CurrentBlock()
	buffer := bytes.NewBuffer(make([]byte, 0))
	if block == nil {
		return nil, status.Errorf(codes.NotFound, "block not found")
	}
	err := block.EncodeRLP(buffer)
	if err != nil {
		log.Error("chain service ", "encode rlp failed with err", err)
		return nil, status.Errorf(codes.Unknown, "encode rlp failed")
	}
	res.BlockData = buffer.Bytes()
	return res, nil
}
func (s *ChainService) LatestHeader(ctx context.Context, req *trusted.LatestHeaderRequest) (*trusted.LatestHeaderResponse, error) {
	header := s.chain.CurrentHeader()
	if header == nil {
		return nil, status.Errorf(codes.NotFound, "header not found")
	}
	res := new(trusted.LatestHeaderResponse)
	res.HeaderJson, _ = json.Marshal(header)
	res.BlockNum = header.Number.Bytes()
	return res, nil
}

func (s *ChainService) ChainHeadEvent(req *trusted.ChainHeadEventRequest, res trusted.ChainService_ChainHeadEventServer) error {
	ch := make(chan core.ChainEvent, 10)
	sub := s.chain.SubscribeChainEvent(ch)
	bcontinue := true
	var err error
	for bcontinue {
		select {
		case e := <-sub.Err():
			err = e
			bcontinue = false
		case newchain, ok := <-ch:
			if !ok {
				err = errors.New("chain event channel closed")
				bcontinue = false
			}
			msg := new(trusted.ChainHeadEventResponse)
			msg.BlockData, _ = rlp.EncodeToBytes(newchain.Block)
			if err = res.SendMsg(msg); err != nil {
				log.Error("chain head event send msg failed", err)
				return err
			}
		}
	}
	return err
}

func RegisterService(server *grpc.Server, chain *core.BlockChain) {
	s := new(ChainService)
	s.chain = chain
	trusted.RegisterChainServiceServer(server, s)
}

func StartChainService(chain *core.BlockChain) {
	lis, err := net.Listen("tcp", ":3801")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	RegisterService(s, chain)
	log.Info("chain service registered")
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
