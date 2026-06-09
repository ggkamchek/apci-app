# Разработка

---

## Git flow

```
main
 └── develop
      ├── module/e2e
      │    └── feature/<название>
      └── module/antifraud
           └── feature/<название>
```

### Правила

| Ветка | Назначение |
|-------|------------|
| `main` | Стабильный код, релизы |
| `develop` | Интеграция всех модулей |
| `module/<имя-модуля>` | Долгоживущая ветка модуля |
| `feature/<название>` | Конкретная задача |

### Merge flow

```
feature/* → module/* → develop → main
```

Имена людей в ветках **не используются** — только имена модулей.

### Примеры веток

- `module/e2e` → `feature/double-ratchet-session`
- `module/antifraud` → `feature/device-fingerprint`
- `module/antifraud` → `feature/admin-links-table`

---

## Структура репозитория

```
.
├── server/                  # Chat API (монолит)
│   ├── cmd/api/
│   ├── internal/
│   │   ├── users/
│   │   ├── e2e/
│   │   ├── gateway/
│   │   └── chats/
│   └── migrations/
├── security-center/         # Центр Безопасности (отдельный бинарник)
│   ├── cmd/security/
│   ├── internal/
│   │   ├── ingest/
│   │   ├── graph/
│   │   └── admin/
│   └── migrations/
├── desktop/                 # Wails client
│   ├── internal/
│   │   ├── crypto/          # E2E
│   │   └── fingerprint/     # device hash
│   └── frontend/            # React UI
├── shared/
│   └── proto/               # gRPC контракты
└── docs/
```

---

## Владельцы модулей

| Модуль | Владелец | Ветка |
|--------|----------|-------|
| E2E | Петя | `module/e2e` |
| Центр Безопасности | Sudeeneess | `module/antifraud` |
| Users, Gateway, Chats | общий минимум | `develop` / по договорённости |

---

## Требования к окружению

- Go 1.22+
- Node.js 18+
- Wails v2
- Docker (PostgreSQL)
- protoc + плагины Go для gRPC

---

## Быстрый старт

См. корневой [README.md](../README.md) для Docker и запуска Chat API.
