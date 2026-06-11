# Коммуникационная платформа

Desktop-приложение и backend: мессенджер, модуль antifraud (Центр Безопасности), далее E2E и чаты.

**Стек:** Go, Wails, PostgreSQL.

---

## Быстрый старт

```powershell
cd D:\Goalang\apci-app
Copy-Item .env.example .env -ErrorAction SilentlyContinue
.\scripts\dev.ps1
```

Откроются два окна: **Chat API** и **Центр Безопасности**. Подождите ~10 сек.

| Что | URL |
|-----|-----|
| Admin UI | http://localhost:8081/admin/ (token: `dev-admin-token`) |
| API health | http://localhost:8080/health |

Desktop:

```powershell
.\scripts\dev.ps1 -Desktop
# или отдельно: cd desktop; wails dev
```

**Документация по запуску:** [`docs/getting-started.md`](docs/getting-started.md)

---

## Структура

```
server/            Chat API
security-center/   Центр Безопасности + админка
desktop/           Wails-клиент
shared/proto/      gRPC-контракты
scripts/dev.ps1    Запуск локальной среды
```

---

## Порты

| Сервис | Порт |
|--------|------|
| Chat API HTTP | 8080 |
| Chat API gRPC | 50051 |
| Security HTTP | 8081 |
| Security gRPC | 50052 |
| PostgreSQL | 5433, 5434 |

---

## Документация

Ежедневный запуск — только [`docs/getting-started.md`](docs/getting-started.md).  
Архитектура и модули — [`docs/README.md`](docs/README.md).
