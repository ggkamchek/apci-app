package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/black/apci-app/desktop/internal/auth"
)

type App struct {
	ctx      context.Context
	mu       sync.Mutex
	client   *auth.Client
	session  *auth.Session
	grpcAddr string
}

func NewApp() *App {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}
	return &App{grpcAddr: grpcAddr}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	configDir, err := os.UserConfigDir()
	if err != nil {
		return
	}

	dataDir := filepath.Join(configDir, "apci-desktop")
	keysDir := filepath.Join(dataDir, "keys")

	client, err := auth.NewClient(a.grpcAddr, keysDir, dataDir)
	if err != nil {
		return
	}
	a.client = client
}

func (a *App) shutdown(context.Context) {
	if a.client != nil {
		_ = a.client.Close()
	}
}

type AuthResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	UserID    string `json:"userId,omitempty"`
	Username  string `json:"username,omitempty"`
	SessionID string `json:"sessionId,omitempty"`
}

func (a *App) Register(username string) AuthResult {
	username = strings.TrimSpace(username)
	if username == "" {
		return AuthResult{Success: false, Message: "введите имя пользователя"}
	}
	if a.client == nil {
		return AuthResult{Success: false, Message: "нет подключения к серверу — запустите API"}
	}

	ctx, cancel := auth.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	session, err := a.client.Register(ctx, username)
	if err != nil {
		return AuthResult{Success: false, Message: err.Error()}
	}

	a.mu.Lock()
	a.session = &session
	a.mu.Unlock()

	return AuthResult{
		Success:   true,
		Message:   "Регистрация успешна",
		UserID:    session.UserID,
		Username:  session.Username,
		SessionID: session.SessionID,
	}
}

func (a *App) Login(username string) AuthResult {
	username = strings.TrimSpace(username)
	if username == "" {
		return AuthResult{Success: false, Message: "введите имя пользователя"}
	}
	if a.client == nil {
		return AuthResult{Success: false, Message: "нет подключения к серверу — запустите API"}
	}

	ctx, cancel := auth.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	session, err := a.client.Login(ctx, username)
	if err != nil {
		return AuthResult{Success: false, Message: err.Error()}
	}

	a.mu.Lock()
	a.session = &session
	a.mu.Unlock()

	return AuthResult{
		Success:   true,
		Message:   "Вход выполнен",
		UserID:    session.UserID,
		Username:  session.Username,
		SessionID: session.SessionID,
	}
}

func (a *App) Logout() {
	a.mu.Lock()
	a.session = nil
	a.mu.Unlock()
}

func (a *App) CurrentUser() AuthResult {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.session == nil {
		return AuthResult{Success: false}
	}

	return AuthResult{
		Success:   true,
		Username:  a.session.Username,
		UserID:    a.session.UserID,
		SessionID: a.session.SessionID,
	}
}
