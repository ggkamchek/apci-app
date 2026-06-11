# Changelog

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/).
Версионирование — [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

### Добавлено

- gRPC-контракт модуля E2E: `shared/proto/apci/e2e/v1/e2e.proto`
  - `UploadPreKeyBundle` — загрузка X25519 prekeys (identity, signed, one-time)
  - `GetPreKeyBundle` — получение bundle собеседника для X3DH
  - `SendMessage` — отправка ciphertext и ratchet header
  - `FetchMessages` — poll входящих сообщений (MVP)
- Сгенерированный Go-код: `server/pkg/pb/apci/e2e/v1/` (`e2e.pb.go`, `e2e_grpc.pb.go`)
- Документация по `E2EService` и команде `protoc` в `server/README.md`

### Зафиксировано в контракте

- Auth для всех RPC E2E — metadata `session-id` (login session из `UsersService`)
- `chat_id` — ссылка на беседу (создаётся модулем chats)
- `ratchet_session_id` — id crypto-сессии Double Ratchet (отдельно от login session)
- X25519 identity key для E2E — отдельно от Ed25519 ключа аутентификации

---

## [0.2.0] — 2026-06-10

### Добавлено

- Модуль **users**: gRPC `UsersService` (Register, BeginLogin, CompleteLogin)
- Аутентификация через **Ed25519 challenge-response** (без пароля и JWT)
- Миграция `001_users.sql`: users, user_devices, login_challenges, sessions
- Автоматический запуск миграций при старте API (`internal/migrate`)
- **Desktop** (Wails): регистрация и вход через UI
  - `desktop/internal/auth` — gRPC-клиент, Ed25519 ключи на диске, device_hash
- Proto: `shared/proto/apci/users/v1/users.proto` + сгенерированный код в `server/pkg/pb/`
- gRPC на `:50051`, переменные `GRPC_ADDR`, `GRPC_PORT`, `CHALLENGE_TTL`, `SESSION_TTL`
- `docs/getting-started.md` — пошаговый локальный запуск

---

## [0.1.0] — 2026-06-09

### Добавлено

- Каркас Chat API: HTTP `GET /health` с проверкой PostgreSQL
- `server/internal/config`, `server/internal/db` — конфиг из env, пул pgx
- Docker Compose: PostgreSQL (`:5433`) + API (`:8080`)
- `.env.example` с переменными окружения

---

## [0.0.1] — 2026-06-08

### Добавлено

- Инициализация монорепозитория `apci-app`
- Структура каталогов: `server/`, `desktop/`, `shared/`, `docs/`
- Документация платформы: overview, architecture, MVP, integration, модули E2E и Центра Безопасности
- README с описанием стека и модулей

