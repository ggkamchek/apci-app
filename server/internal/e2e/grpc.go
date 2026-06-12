package e2e

import (
	"context"
	"errors"
	"strings"

	e2ev1 "github.com/black/apci-app/server/pkg/pb/apci/e2e/v1"
	"github.com/black/apci-app/server/internal/users"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCServer struct {
	e2ev1.UnimplementedE2EServiceServer
	service *Service
}

func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) UploadPreKeyBundle(ctx context.Context, req *e2ev1.UploadPreKeyBundleRequest) (*e2ev1.UploadPreKeyBundleResponse, error) {
	signed := req.GetSignedPrekey()
	if signed == nil {
		return nil, status.Error(codes.InvalidArgument, "signed_prekey is required")
	}
	if signed.GetExpiresAt() == nil {
		return nil, status.Error(codes.InvalidArgument, "signed_prekey.expires_at is required")
	}

	var oneTime []OneTimePreKeyRecord
	for _, key := range req.GetOneTimePrekeys() {
		oneTime = append(oneTime, OneTimePreKeyRecord{
			KeyID:     key.GetKeyId(),
			PublicKey: key.GetPublicKey(),
		})
	}

	accepted, err := s.service.UploadPreKeyBundle(ctx, UploadBundleInput{
		IdentityKey: req.GetIdentityKey(),
		SignedPreKey: SignedPreKeyRecord{
			KeyID:     signed.GetKeyId(),
			PublicKey: signed.GetPublicKey(),
			Signature: signed.GetSignature(),
			ExpiresAt: signed.GetExpiresAt().AsTime(),
		},
		OneTimePrekeys: oneTime,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &e2ev1.UploadPreKeyBundleResponse{
		OneTimePrekeysAccepted: accepted,
	}, nil
}

func (s *GRPCServer) GetPreKeyBundle(ctx context.Context, req *e2ev1.GetPreKeyBundleRequest) (*e2ev1.GetPreKeyBundleResponse, error) {
	bundle, err := s.service.GetPreKeyBundle(ctx, req.GetUserId())
	if err != nil {
		return nil, mapError(err)
	}

	return &e2ev1.GetPreKeyBundleResponse{
		UserId:                bundle.UserID,
		IdentityKey:           bundle.IdentityKey,
		SignedPrekeyId:        bundle.SignedPreKeyID,
		SignedPrekey:          bundle.SignedPreKey,
		SignedPrekeySignature: bundle.SignedPreKeySignature,
		OneTimePrekeyId:       bundle.OneTimePreKeyID,
		OneTimePrekey:         bundle.OneTimePreKey,
	}, nil
}

func (s *GRPCServer) SendMessage(ctx context.Context, req *e2ev1.SendMessageRequest) (*e2ev1.SendMessageResponse, error) {
	msg, err := s.service.SendMessage(ctx, SendMessageInput{
		RecipientID:      req.GetRecipientId(),
		ChatID:           req.GetChatId(),
		RatchetSessionID: req.GetRatchetSessionId(),
		Ciphertext:       req.GetCiphertext(),
		RatchetHeader:    req.GetRatchetHeader(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &e2ev1.SendMessageResponse{
		MessageId: msg.ID,
		SentAt:    timestamppb.New(msg.SentAt),
	}, nil
}

func (s *GRPCServer) FetchMessages(ctx context.Context, req *e2ev1.FetchMessagesRequest) (*e2ev1.FetchMessagesResponse, error) {
	messages, err := s.service.FetchMessages(ctx, req.GetSinceMessageId(), req.GetLimit())
	if err != nil {
		return nil, mapError(err)
	}

	resp := &e2ev1.FetchMessagesResponse{
		Messages: make([]*e2ev1.EncryptedMessage, 0, len(messages)),
	}
	for _, msg := range messages {
		resp.Messages = append(resp.Messages, &e2ev1.EncryptedMessage{
			MessageId:        msg.ID,
			ChatId:           msg.ChatID,
			RatchetSessionId: msg.RatchetSessionID,
			SenderId:         msg.SenderID,
			RecipientId:      msg.RecipientID,
			Ciphertext:         msg.Ciphertext,
			RatchetHeader:      msg.RatchetHeader,
			SentAt:             timestamppb.New(msg.SentAt),
		})
	}
	return resp, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrBundleNotFound):
		return status.Error(codes.NotFound, "prekey bundle not found")
	case errors.Is(err, users.ErrNotFound):
		return status.Error(codes.NotFound, "user not found")
	}

	message := err.Error()
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "session"):
		return status.Error(codes.Unauthenticated, message)
	case strings.Contains(lower, "required"),
		strings.Contains(lower, "must be"),
		strings.Contains(lower, "too large"),
		strings.Contains(lower, "too many"),
		strings.Contains(lower, "duplicate"),
		strings.Contains(lower, "cannot send"):
		return status.Error(codes.InvalidArgument, message)
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
