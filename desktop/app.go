package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/black/apci-app/desktop/internal/auth"
	"github.com/black/apci-app/desktop/internal/crypto"
	"github.com/black/apci-app/desktop/internal/e2e"
	"github.com/black/apci-app/desktop/internal/fingerprint"
	e2ev1 "github.com/black/apci-app/server/pkg/pb/apci/e2e/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type App struct {
	ctx      context.Context
	mu       sync.Mutex
	client   *auth.Client
	session  *auth.Session
	e2e      *e2e.Client
	crypto   *crypto.Store
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

	a.e2e = e2e.NewClient(client.Conn())
	a.crypto = crypto.NewStore(dataDir)
}

func (a *App) shutdown(context.Context) {
	if a.client != nil {
		_ = a.client.Close()
	}
}

func (a *App) SetFingerprintSignals(signals fingerprint.Signals) {
	if a.client != nil {
		a.client.SetFingerprintSignals(signals)
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

type E2EResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type E2ESendResult struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	MessageID string `json:"messageId,omitempty"`
}

type E2EInboxMessage struct {
	MessageID  string `json:"messageId"`
	ChatID     string `json:"chatId"`
	SenderID   string `json:"senderId"`
	Plaintext  string `json:"plaintext"`
	SentAtUnix int64  `json:"sentAtUnix"`
}

type E2EInboxResult struct {
	Success  bool              `json:"success"`
	Message  string            `json:"message"`
	Messages []E2EInboxMessage `json:"messages,omitempty"`
}

func (a *App) E2EInit() E2EResult {
	a.mu.Lock()
	sess := a.session
	client := a.client
	e2eClient := a.e2e
	store := a.crypto
	a.mu.Unlock()

	if sess == nil {
		return E2EResult{Success: false, Message: "нужно войти"}
	}
	if client == nil || e2eClient == nil || store == nil {
		return E2EResult{Success: false, Message: "клиент не готов — перезапустите приложение"}
	}

	ctx, cancel := auth.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	st, err := store.LoadOrCreate()
	if err != nil {
		return E2EResult{Success: false, Message: err.Error()}
	}

	identityPub, spkPub, spkSig, spkID, spkExpUnix, ot, err := crypto.BuildUploadBundle(st)
	if err != nil {
		return E2EResult{Success: false, Message: err.Error()}
	}

	req := &e2ev1.UploadPreKeyBundleRequest{
		IdentityKey: identityPub,
		SignedPrekey: &e2ev1.SignedPreKey{
			KeyId:     spkID,
			PublicKey: spkPub,
			Signature: spkSig,
			ExpiresAt: timestamppb.New(time.Unix(spkExpUnix, 0)),
		},
	}

	for _, k := range ot {
		req.OneTimePrekeys = append(req.OneTimePrekeys, &e2ev1.OneTimePreKey{
			KeyId:     k.ID,
			PublicKey: k.Pub,
		})
	}

	resp, err := e2eClient.UploadPreKeyBundle(ctx, sess.SessionID, req)
	if err != nil {
		return E2EResult{Success: false, Message: err.Error()}
	}
	return E2EResult{Success: true, Message: "E2E bundle загружен, one-time accepted: " + fmt.Sprint(resp.GetOneTimePrekeysAccepted())}
}

func (a *App) E2ESend(recipientID string, chatID string, plaintext string) E2ESendResult {
	recipientID = strings.TrimSpace(recipientID)
	chatID = strings.TrimSpace(chatID)
	plaintext = strings.TrimSpace(plaintext)

	a.mu.Lock()
	sess := a.session
	e2eClient := a.e2e
	store := a.crypto
	a.mu.Unlock()

	if sess == nil {
		return E2ESendResult{Success: false, Message: "нужно войти"}
	}
	if e2eClient == nil || store == nil {
		return E2ESendResult{Success: false, Message: "клиент не готов — перезапустите приложение"}
	}
	if recipientID == "" {
		return E2ESendResult{Success: false, Message: "recipient_id обязателен"}
	}
	if plaintext == "" {
		return E2ESendResult{Success: false, Message: "текст сообщения обязателен"}
	}
	if chatID == "" {
		// V0: пока нет chats модуля — делаем детерминированный chat_id.
		if sess.UserID < recipientID {
			chatID = "mvp:" + sess.UserID + ":" + recipientID
		} else {
			chatID = "mvp:" + recipientID + ":" + sess.UserID
		}
	}

	ctx, cancel := auth.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	st, err := store.LoadOrCreate()
	if err != nil {
		return E2ESendResult{Success: false, Message: err.Error()}
	}

	bundle, err := e2eClient.GetPreKeyBundle(ctx, sess.SessionID, recipientID)
	if err != nil {
		return E2ESendResult{Success: false, Message: err.Error()}
	}

	ct, hdr, err := crypto.EncryptV0(st, bundle.GetIdentityKey(), []byte(plaintext))
	if err != nil {
		return E2ESendResult{Success: false, Message: err.Error()}
	}

	ratchetSessionID := uuid.NewString()
	sendResp, err := e2eClient.SendMessage(ctx, sess.SessionID, &e2ev1.SendMessageRequest{
		RecipientId:      recipientID,
		ChatId:           chatID,
		RatchetSessionId: ratchetSessionID,
		Ciphertext:       ct,
		RatchetHeader:    hdr,
	})
	if err != nil {
		return E2ESendResult{Success: false, Message: err.Error()}
	}

	return E2ESendResult{Success: true, Message: "отправлено", MessageID: sendResp.GetMessageId()}
}

func (a *App) E2EInbox(sinceMessageID string, limit uint32) E2EInboxResult {
	a.mu.Lock()
	sess := a.session
	e2eClient := a.e2e
	store := a.crypto
	a.mu.Unlock()

	if sess == nil {
		return E2EInboxResult{Success: false, Message: "нужно войти"}
	}
	if e2eClient == nil || store == nil {
		return E2EInboxResult{Success: false, Message: "клиент не готов — перезапустите приложение"}
	}

	ctx, cancel := auth.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()

	st, err := store.LoadOrCreate()
	if err != nil {
		return E2EInboxResult{Success: false, Message: err.Error()}
	}

	resp, err := e2eClient.FetchMessages(ctx, sess.SessionID, strings.TrimSpace(sinceMessageID), limit)
	if err != nil {
		return E2EInboxResult{Success: false, Message: err.Error()}
	}

	out := make([]E2EInboxMessage, 0, len(resp.GetMessages()))
	for _, m := range resp.GetMessages() {
		senderPub, err := crypto.ParseSenderIdentityPubFromHeader(m.GetRatchetHeader())
		if err != nil {
			// Skip malformed message but keep going.
			continue
		}
		pt, err := crypto.DecryptV0(st, senderPub, m.GetRatchetHeader(), m.GetCiphertext())
		if err != nil {
			continue
		}
		sentAt := int64(0)
		if ts := m.GetSentAt(); ts != nil {
			sentAt = ts.AsTime().Unix()
		}
		out = append(out, E2EInboxMessage{
			MessageID:  m.GetMessageId(),
			ChatID:     m.GetChatId(),
			SenderID:   m.GetSenderId(),
			Plaintext:  string(pt),
			SentAtUnix: sentAt,
		})
	}

	return E2EInboxResult{Success: true, Message: "ok", Messages: out}
}
