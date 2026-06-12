package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SessionValidator проверяет session-id и возвращает user_id.
type SessionValidator interface {
	ValidateSession(ctx context.Context, sessionID string) (userID string, err error)
}

const sessionMetadataKey = "session-id"

// UnarySessionInterceptor требует session-id для RPC E2EService.
func UnarySessionInterceptor(validator SessionValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !strings.HasPrefix(info.FullMethod, "/apci.e2e.v1.E2EService/") {
			return handler(ctx, req)
		}

		sessionID := sessionIDFromContext(ctx)
		if sessionID == "" {
			return nil, status.Error(codes.Unauthenticated, "session-id metadata is required")
		}

		userID, err := validator.ValidateSession(ctx, sessionID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
		}

		ctx = WithUserSession(ctx, UserSession{
			UserID:    userID,
			SessionID: sessionID,
		})
		return handler(ctx, req)
	}
}

func sessionIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(sessionMetadataKey)
	if len(values) == 0 {
		return ""
	}
	return strings.TrimSpace(values[0])
}
