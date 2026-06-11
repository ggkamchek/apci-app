package ingest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/black/apci-app/security-center/internal/graph"
	"github.com/google/uuid"
)

const (
	EventUserRegistered = "user.registered"
	EventUserLoggedIn   = "user.logged_in"
	EventContactAdded   = "contact.added"
	EventMessageSent    = "message.sent"
)

type Input struct {
	EventID    string
	EventType  string
	OccurredAt time.Time
	UserID     string
	DeviceHash string
	ContactID  string
	PeerUserID string
	MessageID  string
}

type Result struct {
	Accepted  bool
	Duplicate bool
}

type Service struct {
	repo *graph.Repository
	now  func() time.Time
}

func NewService(repo *graph.Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) Ingest(ctx context.Context, input Input) (Result, error) {
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" {
		return Result{}, fmt.Errorf("event_id is required")
	}
	if _, err := uuid.Parse(eventID); err != nil {
		return Result{}, fmt.Errorf("event_id must be a valid uuid")
	}

	eventType := strings.TrimSpace(input.EventType)
	if eventType == "" {
		return Result{}, fmt.Errorf("event_type is required")
	}

	duplicate, err := s.repo.IsEventProcessed(ctx, eventID)
	if err != nil {
		return Result{}, err
	}
	if duplicate {
		return Result{Accepted: true, Duplicate: true}, nil
	}

	seenAt := input.OccurredAt
	if seenAt.IsZero() {
		seenAt = s.now().UTC()
	}

	switch eventType {
	case EventUserRegistered, EventUserLoggedIn:
		if err := s.handleUserActivity(ctx, input.UserID, input.DeviceHash, seenAt); err != nil {
			return Result{}, err
		}
	case EventContactAdded:
		if err := s.handleContactAdded(ctx, input.UserID, input.ContactID, seenAt); err != nil {
			return Result{}, err
		}
	case EventMessageSent:
		if err := s.handleMessageSent(ctx, input.UserID, input.PeerUserID, seenAt); err != nil {
			return Result{}, err
		}
	default:
		return Result{}, fmt.Errorf("unsupported event_type %q", eventType)
	}

	if err := s.repo.MarkEventProcessed(ctx, eventID, eventType); err != nil {
		return Result{}, err
	}

	return Result{Accepted: true}, nil
}

func (s *Service) handleUserActivity(ctx context.Context, userID, deviceHash string, seenAt time.Time) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("user_id must be a valid uuid")
	}

	if err := s.repo.TouchAccount(ctx, userID, seenAt); err != nil {
		return err
	}

	deviceHash = strings.TrimSpace(deviceHash)
	if deviceHash != "" {
		if err := s.repo.UpsertDevice(ctx, userID, deviceHash, seenAt); err != nil {
			return err
		}
		others, err := s.repo.FindUsersByDevice(ctx, deviceHash, userID)
		if err != nil {
			return err
		}
		for _, other := range others {
			if err := s.repo.UpsertLink(ctx, userID, other, graph.LinkSameDevice, graph.WeightSameDevice, seenAt); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) handleContactAdded(ctx context.Context, userID, contactID string, seenAt time.Time) error {
	userID = strings.TrimSpace(userID)
	contactID = strings.TrimSpace(contactID)
	if userID == "" || contactID == "" {
		return fmt.Errorf("user_id and contact_id are required")
	}
	if _, err := uuid.Parse(userID); err != nil {
		return fmt.Errorf("user_id must be a valid uuid")
	}
	if _, err := uuid.Parse(contactID); err != nil {
		return fmt.Errorf("contact_id must be a valid uuid")
	}

	if err := s.repo.TouchAccount(ctx, userID, seenAt); err != nil {
		return err
	}
	if err := s.repo.TouchAccount(ctx, contactID, seenAt); err != nil {
		return err
	}

	return s.repo.UpsertLink(ctx, userID, contactID, graph.LinkSharedContact, graph.WeightSharedContact, seenAt)
}

func (s *Service) handleMessageSent(ctx context.Context, fromUserID, toUserID string, seenAt time.Time) error {
	fromUserID = strings.TrimSpace(fromUserID)
	toUserID = strings.TrimSpace(toUserID)
	if fromUserID == "" || toUserID == "" {
		return fmt.Errorf("user_id and peer_user_id are required")
	}
	if _, err := uuid.Parse(fromUserID); err != nil {
		return fmt.Errorf("user_id must be a valid uuid")
	}
	if _, err := uuid.Parse(toUserID); err != nil {
		return fmt.Errorf("peer_user_id must be a valid uuid")
	}

	if err := s.repo.TouchAccount(ctx, fromUserID, seenAt); err != nil {
		return err
	}
	if err := s.repo.TouchAccount(ctx, toUserID, seenAt); err != nil {
		return err
	}

	return s.repo.UpsertLink(ctx, fromUserID, toUserID, graph.LinkCommunicatesWith, graph.WeightCommunicatesWith, seenAt)
}

func (s *Service) ListLinks(ctx context.Context) ([]graph.Link, error) {
	return s.repo.ListLinks(ctx)
}

func (s *Service) GetStats(ctx context.Context) (graph.Stats, error) {
	return s.repo.GetStats(ctx)
}

func (s *Service) TopDeviceClones(ctx context.Context, limit int) ([]graph.DeviceClone, error) {
	return s.repo.TopDeviceClones(ctx, limit)
}

func (s *Service) ActivitySeries(ctx context.Context, hours int) ([]graph.ActivityPoint, error) {
	return s.repo.ActivitySeries(ctx, hours)
}

func (s *Service) ListDeviceRegistry(ctx context.Context, filter string, limit int) ([]graph.DeviceRegistryItem, error) {
	return s.repo.ListDeviceRegistry(ctx, filter, limit)
}

func (s *Service) GetAccount(ctx context.Context, userID string) (*graph.AccountProfile, error) {
	return s.repo.GetAccount(ctx, userID)
}

func (s *Service) FindAccountByDevice(ctx context.Context, deviceHash string) (*graph.AccountProfile, error) {
	return s.repo.FindAccountByDevice(ctx, deviceHash)
}

func (s *Service) GetRelatedAccounts(ctx context.Context, userID string) (graph.RelatedResult, error) {
	return s.repo.GetRelatedAccounts(ctx, userID)
}

func (s *Service) FindAccountsByDeviceQuery(ctx context.Context, query string) (*graph.DeviceLookupResult, error) {
	return s.repo.FindAccountsByDeviceQuery(ctx, query)
}

func (s *Service) DiffAccounts(ctx context.Context, userA, userB string) (*graph.DiffResult, error) {
	return s.repo.DiffAccounts(ctx, userA, userB)
}
