package e2e

import (
	"context"
	"fmt"

	"github.com/black/apci-app/desktop/internal/auth"
	e2ev1 "github.com/black/apci-app/server/pkg/pb/apci/e2e/v1"
	"google.golang.org/grpc"
)

type Client struct {
	grpcClient e2ev1.E2EServiceClient
}

func NewClient(conn grpc.ClientConnInterface) *Client {
	return &Client{grpcClient: e2ev1.NewE2EServiceClient(conn)}
}

func (c *Client) UploadPreKeyBundle(ctx context.Context, sessionID string, req *e2ev1.UploadPreKeyBundleRequest) (*e2ev1.UploadPreKeyBundleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	return c.grpcClient.UploadPreKeyBundle(auth.ContextWithSessionID(ctx, sessionID), req)
}

func (c *Client) GetPreKeyBundle(ctx context.Context, sessionID string, userID string) (*e2ev1.GetPreKeyBundleResponse, error) {
	return c.grpcClient.GetPreKeyBundle(auth.ContextWithSessionID(ctx, sessionID), &e2ev1.GetPreKeyBundleRequest{
		UserId: userID,
	})
}

func (c *Client) SendMessage(ctx context.Context, sessionID string, req *e2ev1.SendMessageRequest) (*e2ev1.SendMessageResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	return c.grpcClient.SendMessage(auth.ContextWithSessionID(ctx, sessionID), req)
}

func (c *Client) FetchMessages(ctx context.Context, sessionID string, sinceMessageID string, limit uint32) (*e2ev1.FetchMessagesResponse, error) {
	return c.grpcClient.FetchMessages(auth.ContextWithSessionID(ctx, sessionID), &e2ev1.FetchMessagesRequest{
		SinceMessageId: sinceMessageID,
		Limit:          limit,
	})
}

