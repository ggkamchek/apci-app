# Центр Безопасности

Отдельный сервис: ingest событий от Chat API, граф связей, web-админка.

**Запуск:** из корня репозитория:

```powershell
.\scripts\dev.ps1
```

**Admin UI:** http://localhost:8081/admin/ · token: `dev-admin-token`

Полная инструкция: [docs/getting-started.md](../docs/getting-started.md)

---

## Порты

| Сервис | Порт |
|--------|------|
| HTTP (health + admin) | `:8081` |
| gRPC IngestEvent | `:50052` |
| PostgreSQL | `localhost:5434` |

---

## События ingest

| event_type | Источник | Статус |
|------------|----------|--------|
| `user.registered` | Chat API после регистрации | ✅ |
| `user.logged_in` | Chat API после входа | ✅ |
| `contact.added` | Chat API после добавления контакта | 📋 |
| `message.sent` | Chat API после сохранения ciphertext | 📋 |

---

## Admin API (кратко)

| Endpoint | Описание |
|----------|----------|
| `GET /admin/stats` | Сводка |
| `GET /admin/links` | Связи |
| `GET /admin/lookup?q=` | Поиск по UUID или хешу |
| `GET /admin/account/{id}/related` | Связанные аккаунты |

Auth: `Authorization: Bearer <SECURITY_ADMIN_TOKEN>`

Proto: `shared/proto/apci/security/v1/ingest.proto`

Документация модуля: [docs/modules/security-center.md](../docs/modules/security-center.md)
