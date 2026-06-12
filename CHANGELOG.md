# Changelog

Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/).
Версионирование — [Semantic Versioning](https://semver.org/lang/ru/).

## [Unreleased]

### Добавлено

- Реализация **E2EService** на Chat API (`server/internal/e2e`)
  - `UploadPreKeyBundle`, `GetPreKeyBundle`, `SendMessage`, `FetchMessages`
  - Миграция `002_e2e.sql`: prekeys и ciphertext
  - gRPC interceptor: metadata `session-id` для всех RPC E2E
  - `users`: проверка сессии (`ValidateSession`, `GetByID`)

---

## [0.4.0] — 2026-06-12

### Добавлено

- **Центр Безопасности** — отдельный сервис `security-center/` (gRPC ingest + граф + admin UI)
  - `SecurityIngest.IngestEvent` — приём событий от Chat API (`shared/proto/apci/security/v1/ingest.proto`)
  - События MVP: `user.registered`, `user.logged_in`
  - Граф связей: `same_device`, `shared_contact`
  - Web-админка (русский UI): http://localhost:8081/admin/
  - Миграция `001_graph.sql`, отдельная PostgreSQL (`:5434`)
- **Chat API:** fire-and-forget публикация событий в Центр (`server/internal/securitycenter/publisher.go`)
- **Desktop:** сбор fingerprint-сигналов (`desktop/internal/fingerprint`), передача в Register/Login
- **Docker Compose:** сервисы `postgres-security`, `security-center`
- **Скрипт** `scripts/dev.ps1` — единый локальный запуск (Chat API + Центр + Postgres)
- Обновлены `docs/getting-started.md`, `README.md`, документация интеграции

### Изменено

- `desktop/internal/auth/fingerprint.go` — device hash упрощён; сигналы вынесены в модуль fingerprint
- `server/internal/users/grpc.go` — передача fingerprint в события безопасности

---

## [0.3.0] — 2026-06-11

### Добавлено

- gRPC-контракт модуля E2E: `shared/proto/apci/e2e/v1/e2e.proto`
  - `UploadPreKeyBundle` — загрузка X25519 prekeys (identity, signed, one-time)
  - `GetPreKeyBundle` — получение bundle собеседника для X3DH
  - `SendMessage` — отправка ciphertext и ratchet header
  - `FetchMessages` — poll входящих сообщений (MVP)
- Сгенерированный Go-код: `server/pkg/pb/apci/e2e/v1/` (`e2e.pb.go`, `e2e_grpc.pb.go`)
- Документация по `E2EService` и команде `protoc` в `server/README.md`
- `CHANGELOG.md`

### Зафиксировано в контракте

- Auth для всех RPC E2E — metadata `session-id` (сессия входа из `UsersService`)
- `chat_id` — ссылка на беседу (создаётся модулем chats)
- `ratchet_session_id` — id crypto-сессии Double Ratchet (отдельно от сессии входа)
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

[Unreleased]: https://github.com/ggkamchek/apci-app/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/ggkamchek/apci-app/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/ggkamchek/apci-app/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/ggkamchek/apci-app/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/ggkamchek/apci-app/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/ggkamchek/apci-app/releases/tag/v0.0.1
