package ingest

import (
	"context"

	securityv1 "github.com/black/apci-app/security-center/pkg/pb/apci/security/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	securityv1.UnimplementedSecurityIngestServer
	service *Service
}

func NewGRPCServer(service *Service) *GRPCServer {
	return &GRPCServer{service: service}
}

func (s *GRPCServer) IngestEvent(ctx context.Context, req *securityv1.IngestEventRequest) (*securityv1.IngestEventResponse, error) {
	var occurredAt = req.GetOccurredAt().AsTime()

	result, err := s.service.Ingest(ctx, Input{
		EventID:    req.GetEventId(),
		EventType:  req.GetEventType(),
		OccurredAt: occurredAt,
		UserID:     req.GetUserId(),
		DeviceHash: req.GetDeviceHash(),
		ContactID:  req.GetContactId(),
		PeerUserID: req.GetPeerUserId(),
		MessageID:  req.GetMessageId(),
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &securityv1.IngestEventResponse{
		Accepted:  result.Accepted,
		Duplicate: result.Duplicate,
	}, nil
}

func mapError(err error) error {
	if err == nil {
		return nil
	}
	return status.Error(codes.InvalidArgument, err.Error())
}
