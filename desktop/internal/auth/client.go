package auth

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"time"

	usersv1 "github.com/black/apci-app/server/pkg/pb/apci/users/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Session struct {
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	SessionID string `json:"sessionId"`
}

type Client struct {
	keysDir    string
	dataDir    string
	grpcClient usersv1.UsersServiceClient
	conn       *grpc.ClientConn
}

func NewClient(addr, keysDir, dataDir string) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к серверу: %w", err)
	}

	return &Client{
		keysDir:    keysDir,
		dataDir:    dataDir,
		grpcClient: usersv1.NewUsersServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) Register(ctx context.Context, username string) (Session, error) {
	if HasKeyPair(c.keysDir, username) {
		return Session{}, fmt.Errorf("на этом устройстве уже есть ключи для %q — войдите", username)
	}

	pair, err := GenerateKeyPair()
	if err != nil {
		return Session{}, err
	}

	deviceHash, err := DeviceHash(c.dataDir)
	if err != nil {
		return Session{}, err
	}

	_, err = c.grpcClient.Register(ctx, &usersv1.RegisterRequest{
		Username:         username,
		Ed25519PublicKey: pair.PublicKey,
		DeviceHash:       deviceHash,
	})
	if err != nil {
		return Session{}, mapGRPCError(err)
	}

	if err := SaveKeyPair(c.keysDir, username, pair); err != nil {
		return Session{}, err
	}

	return c.loginWithPair(ctx, username, pair, deviceHash)
}

func (c *Client) Login(ctx context.Context, username string) (Session, error) {
	pair, err := LoadKeyPair(c.keysDir, username)
	if err != nil {
		return Session{}, err
	}

	deviceHash, err := DeviceHash(c.dataDir)
	if err != nil {
		return Session{}, err
	}

	return c.loginWithPair(ctx, username, pair, deviceHash)
}

func (c *Client) loginWithPair(ctx context.Context, username string, pair KeyPair, deviceHash string) (Session, error) {
	beginResp, err := c.grpcClient.BeginLogin(ctx, &usersv1.BeginLoginRequest{Username: username})
	if err != nil {
		return Session{}, mapGRPCError(err)
	}

	signature := ed25519.Sign(pair.PrivateKey, beginResp.GetChallenge())

	completeResp, err := c.grpcClient.CompleteLogin(ctx, &usersv1.CompleteLoginRequest{
		Username:    username,
		ChallengeId: beginResp.GetChallengeId(),
		Signature:   signature,
		DeviceHash:  deviceHash,
	})
	if err != nil {
		return Session{}, mapGRPCError(err)
	}

	return Session{
		UserID:    completeResp.GetUserId(),
		Username:  completeResp.GetUsername(),
		SessionID: completeResp.GetSessionId(),
	}, nil
}

func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}

func mapGRPCError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Message() {
		case "username already exists":
			return fmt.Errorf("имя пользователя уже занято")
		case "user not found":
			return fmt.Errorf("пользователь не найден")
		case "invalid signature":
			return fmt.Errorf("неверная подпись — возможно, ключи на устройстве не совпадают с аккаунтом")
		case "challenge not found or expired":
			return fmt.Errorf("сессия входа истекла — попробуйте снова")
		default:
			return fmt.Errorf("%s", st.Message())
		}
	}
	return err
}
