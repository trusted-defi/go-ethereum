package chainservice

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core"
	trusted "github.com/ethereum/go-ethereum/trusted/protocol/generate/trusted/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type ChainService struct {
	chain *core.BlockChain
	trusted.UnimplementedChainServiceServer
}

func (s *ChainService) GetBlock(ctx context.Context, req *trusted.BlockRequest) (*trusted.BlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (s *ChainService) GetBalance(ctx context.Context, req *trusted.BalanceRequest) (*trusted.BalanceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (s *ChainService) GetNonce(ctx context.Context, req *trusted.NonceRequest) (*trusted.NonceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetNonce not implemented")
}
func (s *ChainService) CurrentBlock(ctx context.Context, req *trusted.CurrentBlockRequest) (*trusted.CurrentBlockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CurrentBlock not implemented")
}
func (s *ChainService) LatestHeader(ctx context.Context, req *trusted.LatestHeaderRequest) (*trusted.LatestHeaderResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LatestHeader not implemented")
}
func (s *ChainService) ChainHeadEvent(req *trusted.ChainHeadEventRequest, res trusted.ChainService_ChainHeadEventServer) error {
	return status.Errorf(codes.Unimplemented, "method ChainHeadEvent not implemented")
}

func RegisterService(server *grpc.Server, chain *core.BlockChain) {
	s := new(ChainService)
	s.chain = chain
	trusted.RegisterChainServiceServer(server, s)
}

func StartChainService(chain *core.BlockChain) {
	lis, err := net.Listen("tcp", ":38000")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	RegisterService(s, chain)

	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}
