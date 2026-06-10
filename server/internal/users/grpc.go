package users

import (
	"context"
	"errors"
	"strings"

	usersv1 "github.com/black/apci-app/server/pkg/pb/apci/users/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	usersv1.UnimplementedUsersServiceServer
	service *Service
}

func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) Register(ctx context.Context, req *usersv1.RegisterRequest) (*usersv1.RegisterResponse, error) {
	result, err := s.service.Register(ctx, RegisterInput{
		Username:         req.GetUsername(),
		Ed25519PublicKey: req.GetEd25519PublicKey(),
		DeviceHash:       req.GetDeviceHash(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &usersv1.RegisterResponse{
		UserId:   result.UserID,
		Username: result.Username,
	}, nil
}

func (s *GRPCServer) BeginLogin(ctx context.Context, req *usersv1.BeginLoginRequest) (*usersv1.BeginLoginResponse, error) {
	result, err := s.service.BeginLogin(ctx, req.GetUsername())
	if err != nil {
		return nil, mapError(err)
	}

	return &usersv1.BeginLoginResponse{
		ChallengeId: result.ChallengeID,
		Challenge:   result.Challenge,
	}, nil
}

func (s *GRPCServer) CompleteLogin(ctx context.Context, req *usersv1.CompleteLoginRequest) (*usersv1.CompleteLoginResponse, error) {
	result, err := s.service.CompleteLogin(ctx, CompleteLoginInput{
		Username:    req.GetUsername(),
		ChallengeID: req.GetChallengeId(),
		Signature:   req.GetSignature(),
		DeviceHash:  req.GetDeviceHash(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &usersv1.CompleteLoginResponse{
		SessionId: result.SessionID,
		UserId:    result.UserID,
		Username:  result.Username,
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, "username already exists")
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, "user not found")
	case errors.Is(err, ErrChallenge):
		return status.Error(codes.FailedPrecondition, "challenge not found or expired")
	}

	message := err.Error()
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "invalid signature"):
		return status.Error(codes.Unauthenticated, "invalid signature")
	case strings.Contains(lower, "must be"):
		return status.Error(codes.InvalidArgument, message)
	case strings.Contains(lower, "is required"):
		return status.Error(codes.InvalidArgument, message)
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
