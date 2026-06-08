# Коммуникационная платформа

Desktop-приложение и backend для платформы real-time коммуникаций: мессенджер с E2E-шифрованием, модуль antifraud, далее — голос, стримы и экономика.

**Стек:** Go (monolith API), Wails + React (desktop), PostgreSQL, Redis (позже).

---

## Структура репозитория

```
.
├── server/          # Go API monolith (deploy на сервер)
│   ├── cmd/api/     # точка входа API
│   ├── internal/    # модули: users, e2e, antifraud, gateway, …
│   └── migrations/  # SQL-миграции PostgreSQL
├── desktop/         # Wails desktop client (Go + React)
│   ├── internal/    # crypto (E2E), fingerprint, HTTP-клиент к API
│   └── frontend/    # React UI
├── shared/          # общие DTO и модели (Go)
└── docs/
    └── architecture/  # trust boundaries, модули, диаграммы
```

---

## Модули

| Модуль | Путь (server) | Путь (desktop) | Владелец |
|--------|---------------|----------------|----------|
| users | `server/internal/users` | — | общий |
| e2e | `server/internal/e2e` | `desktop/internal/crypto` | Петя |
| antifraud | `server/internal/antifraud` | `desktop/internal/fingerprint` | Sudeeneess |
| gateway | `server/internal/gateway` | — | общий |

Склейка модулей — через `users.id`. Private keys и plaintext сообщений **только на клиенте**.

---

## Требования

- Go 1.22+
- Node.js 18+ (для Wails frontend)
- [Wails v2](https://wails.io/docs/gettingstarted/installation)
- Docker (PostgreSQL)

---

## Быстрый старт

### 1. Docker (PostgreSQL + API)

```bash
cp .env.example .env
docker compose up --build -d
```

**Порты по умолчанию:**

| Сервис | URL / порт |
|--------|------------|
| API | http://localhost:8080 |
| PostgreSQL (с хоста) | `localhost:5433` |

PostgreSQL проброшен на **5433**, чтобы не конфликтовать с локальным Postgres на 5432.  
Внутри Docker-сети API подключается к `postgres:5432` — менять не нужно.

Проверка API:

```bash
curl http://localhost:8080/health
```

Ожидаемый ответ: `{"status":"ok","database":"ok"}`

Проверка БД с хоста (опционально):

```bash
psql postgres://platform:platform@localhost:5433/platform
```

Остановка:

```bash
docker compose down
```

### 2. API локально (без Docker для Go)

Подними только Postgres в Docker, API запусти на машине:

```bash
cp .env.example .env
docker compose up -d postgres
cd server
go run ./cmd/api
```

`DATABASE_URL` из `.env` должен указывать на **`localhost:5433`** (см. `.env.example`).

### 3. Desktop (dev, позже)

```bash
cd desktop
wails dev
```

---

## Архитектура

- **Модульный монолит** на Go — users, e2e, antifraud в одном процессе
- **Desktop** — Wails: React (UI) + Go (crypto, fingerprint, API client)
- **E2E** — Double Ratchet (Signal Protocol), сервер хранит только ciphertext и public prekeys
- **Сателлиты (позже):** mediasoup (C++), ML-модерация (Python) — через gRPC

Документы: `docs/architecture/`

---

## Лицензия

TBD
