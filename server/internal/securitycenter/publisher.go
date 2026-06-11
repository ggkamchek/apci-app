package securitycenter

import (
	"context"
	"log"
	"strings"
	"time"

	securityv1 "github.com/black/apci-app/security-center/pkg/pb/apci/security/v1"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	EventUserRegistered = "user.registered"
	EventUserLoggedIn   = "user.logged_in"
)

type Publisher struct {
	client securityv1.SecurityIngestClient
}

func NewPublisher(addr string) (*Publisher, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return nil, nil
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Publisher{
		client: securityv1.NewSecurityIngestClient(conn),
	}, nil
}

func (p *Publisher) PublishUserRegistered(userID, deviceHash string) {
	p.publish(EventUserRegistered, userID, deviceHash)
}

func (p *Publisher) PublishUserLoggedIn(userID, deviceHash string) {
	p.publish(EventUserLoggedIn, userID, deviceHash)
}

func (p *Publisher) publish(eventType, userID, deviceHash string) {
	if p == nil || p.client == nil {
		return
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := p.client.IngestEvent(ctx, &securityv1.IngestEventRequest{
			EventId:    uuid.NewString(),
			EventType:  eventType,
			OccurredAt: timestamppb.Now(),
			UserId:     userID,
			DeviceHash: deviceHash,
		})
		if err != nil {
			log.Printf("security ingest %s: %v", eventType, err)
		}
	}()
}
