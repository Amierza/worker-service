package grpcclient

import (
	"context"
	"fmt"
	"log"

	pb "github.com/Amierza/ai-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type SummaryClient struct {
	client pb.SummaryServiceClient
	conn   *grpc.ClientConn
}

func NewSummaryClient(target string) (*SummaryClient, error) {
	// Membuat koneksi ke gRPC server
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server: %w", err)
	}

	client := pb.NewSummaryServiceClient(conn)
	return &SummaryClient{
		client: client,
		conn:   conn,
	}, nil
}

func (c *SummaryClient) Close() {
	if err := c.conn.Close(); err != nil {
		log.Printf("failed to close gRPC connection: %v", err)
	}
}

func (c *SummaryClient) GenerateSummary(ctx context.Context, req *pb.SummaryRequest) (*pb.SummaryResponse, error) {
	return c.client.GenerateSummary(ctx, req)
}
