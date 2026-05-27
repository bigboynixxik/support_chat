package clients

import (
	"context"
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "support_chat/pkg/api/v1/support_auth_v1"
)

type Client struct {
	api pb.SupportAuthServiceClient
	log *slog.Logger
}

func New(target string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("auth.New grpc.NewClient: %w", err)
	}

	return &Client{
		api: pb.NewSupportAuthServiceClient(conn),
		log: log.With(slog.String("component", "grpc_auth_client")),
	}, nil
}

func (c *Client) Validate(ctx context.Context, token string) (*pb.ValidateResponse, error) {
	req := &pb.ValidateRequest{
		JwtToken: token,
	}

	resp, err := c.api.ValidateWSSession(ctx, req)
	if err != nil {
		c.log.Error("failed to validate token via grpc", slog.String("error", err.Error()))
		return nil, fmt.Errorf("auth.Validate: %w", err)
	}

	return resp, nil
}
