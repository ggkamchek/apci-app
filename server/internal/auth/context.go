package auth

import "context"

type contextKey struct{}

// UserSession — данные сессии входа из metadata session-id.
type UserSession struct {
	UserID    string
	SessionID string
}

func WithUserSession(ctx context.Context, session UserSession) context.Context {
	return context.WithValue(ctx, contextKey{}, session)
}

func UserSessionFromContext(ctx context.Context) (UserSession, bool) {
	session, ok := ctx.Value(contextKey{}).(UserSession)
	return session, ok
}
