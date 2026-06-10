# Коммуникационная платформа

Desktop-приложение и backend для платформы real-time коммуникаций: мессенджер с E2E-шифрованием, модуль antifraud, далее — голос, стримы и экономика.

**Стек:** Go (Chat API + Центр Безопасности), Wails (desktop), PostgreSQL, Redis (позже).

---

## Структура репозитория

```
.
├── server/              # Chat API (модульный монолит)
│   ├── cmd/api/
│   ├── internal/        # users, e2e, gateway, chats
│   ├── pkg/pb/          # сгенерированный protobuf (общий для server и desktop)
│   └── migrations/
├── security-center/     # Центр Безопасности (отдельный бинарник, позже)
├── desktop/             # Wails desktop client
│   ├── internal/auth/   # ключи, gRPC-клиент
│   └── frontend/        # UI входа и главного экрана
├── shared/proto/        # proto-контракты
└── docs/                # документация проекта
```

---

## Требования

- Go 1.22+
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- Docker Desktop (PostgreSQL)

---

## Быстрый старт

**Полная инструкция:** [`docs/getting-started.md`](docs/getting-started.md) — Docker, API, desktop, типичные ошибки.

### Кратко (Windows, PowerShell)

**Терминал 1 — БД и API:**

```powershell
cd D:\Goalang\apci-app
Copy-Item .env.example .env -ErrorAction SilentlyContinue
docker compose up -d postgres
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') { Set-Item -Path "env:$($matches[1].Trim())" -Value $matches[2].Trim() }
}
cd server
go run ./cmd/api
```

Проверка: `curl http://localhost:8080/health` → `{"status":"ok","database":"ok"}`

**Терминал 2 — приложение:**

```powershell
cd D:\Goalang\apci-app\desktop
wails dev
```

В окне APCI: регистрация или вход по username.

---

## Модули

| Модуль | Путь (server) | Путь (desktop) | Владелец |
|--------|---------------|----------------|----------|
| users | `server/internal/users` | `desktop/internal/auth` | общий |
| e2e | `server/internal/e2e` | `desktop/internal/crypto` | Петя |
| Центр Безопасности | `security-center/` | `desktop/internal/fingerprint` | Sudeeneess |
| gateway | `server/internal/gateway` | — | общий |

Private keys и plaintext сообщений **только на клиенте**.

---

## Порты

| Сервис | URL / порт |
|--------|------------|
| HTTP health | http://localhost:8080 |
| gRPC Users API | localhost:50051 |
| PostgreSQL (с хоста) | localhost:5433 |

---

## Архитектура

- **Chat API** — модульный монолит на Go (users, e2e, gateway, chats)
- **Центр Безопасности** — отдельный бинарник (позже)
- **Desktop** — Wails: UI + Go (ключи, gRPC client)
- **E2E** — Double Ratchet (Signal Protocol), roadmap

Документы: [`docs/`](docs/README.md)

---

## Лицензия

TBD
