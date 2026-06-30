# Разработка

---

## Git flow

```
main
 └── develop              # стабильная интеграция всех модулей
      └── apci-future     # активная разработка продукта
           └── feature/<название>
```

### Правила

| Ветка | Назначение |
|-------|------------|
| `main` | Стабильный код, релизы |
| `develop` | Интеграция проверенных изменений |
| `apci-future` | Текущая рабочая ветка: новые фичи, рефакторинг, документация |
| `module/<имя-модуля>` | Долгоживущая ветка модуля (при необходимости) |
| `feature/<название>` | Конкретная задача |

### Merge flow

```
feature/* → apci-future → develop → main
```

Имена людей в ветках **не используются** — только имена модулей или задач.

### Примеры веток

- `apci-future` → `feature/e2e-group-chat`
- `apci-future` → `feature/security-ingest-messages`
- `module/antifraud` → `feature/admin-investigation` (историческая ветка модуля)

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
│   │   ├── securitycenter/
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
│   └── frontend/
├── shared/
│   └── proto/               # gRPC контракты
├── scripts/
│   └── dev.ps1              # Локальный запуск
└── docs/
```

---

## Владельцы модулей

| Модуль | Владелец |
|--------|----------|
| E2E | Петя |
| Центр Безопасности | Sudeeneess |
| Users, Gateway, инфраструктура | общая разработка |

---

## Требования к окружению

- Go 1.22+
- Node.js 18+
- Wails v2
- Docker (PostgreSQL)
- protoc + плагины Go для gRPC

---

## Быстрый старт

См. [getting-started.md](./getting-started.md) и корневой [README.md](../README.md).
