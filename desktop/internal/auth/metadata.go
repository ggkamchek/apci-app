package auth

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const SessionIDMetadataKey = "session-id"

func ContextWithSessionID(ctx context.Context, sessionID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return metadata.AppendToOutgoingContext(ctx, SessionIDMetadataKey, sessionID)
}

