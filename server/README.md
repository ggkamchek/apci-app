# Chat API

Go-сервер: HTTP health + gRPC UsersService (регистрация и вход) + E2EService (prekeys и relay ciphertext).

**Запуск и решение проблем:** [docs/getting-started.md](../docs/getting-started.md)

---

## Быстрый старт

```powershell
# из корня репо
docker compose up -d postgres

Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') { Set-Item -Path "env:$($matches[1].Trim())" -Value $matches[2].Trim() }
}
cd server
go run ./cmd/api
```

Сервисы:
- HTTP health: `http://localhost:8080/health`
- gRPC: `localhost:50051`

Тестирование auth — через **desktop** (`cd desktop && wails dev`), см. getting-started.

---

## gRPC: UsersService

Proto: `shared/proto/apci/users/v1/users.proto`

| RPC | Описание |
|-----|----------|
| `Register` | username + ed25519_public_key (32 bytes) + device_hash |
| `BeginLogin` | username → challenge_id + challenge |
| `CompleteLogin` | username + challenge_id + signature (64 bytes) + device_hash |

Вход — **Ed25519 challenge-response**, без пароля и JWT.

---

## gRPC: E2EService

Proto: `shared/proto/apci/e2e/v1/e2e.proto`

| RPC | Описание |
|-----|----------|
| `UploadPreKeyBundle` | X25519 identity + signed/one-time prekeys |
| `GetPreKeyBundle` | prekey bundle другого пользователя (X3DH) |
| `SendMessage` | ciphertext + ratchet header |
| `FetchMessages` | poll входящих (MVP) |

**Metadata (все RPC):** `session-id` — сессия входа из `CompleteLogin`.

**Поля:** `chat_id` — из модуля chats; `ratchet_session_id` — id crypto-сессии Double Ratchet (не сессия входа).

**Auth E2E:** все RPC требуют metadata `session-id` (UUID из `CompleteLogin`).

Проверка списка сервисов:

```powershell
grpcurl -plaintext localhost:50051 list
```

---

## Генерация protobuf

Из **корня репозитория** (нужен `protoc`):

```bash
protoc \
  --proto_path=shared/proto \
  --go_out=server/pkg/pb --go_opt=paths=source_relative \
  --go-grpc_out=server/pkg/pb --go-grpc_opt=paths=source_relative \
  shared/proto/apci/users/v1/users.proto \
  shared/proto/apci/e2e/v1/e2e.proto
```

Сгенерированный код в `server/pkg/pb/` — импортируется и из `desktop`.
