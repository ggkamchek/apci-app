# Запуск проекта локально

**Одна команда** поднимает Docker, Chat API и Центр Безопасности:

```powershell
cd D:\Goalang\apci-app
.\scripts\dev.ps1
```

Перезапуск (освободить порты и поднять заново):

```powershell
.\scripts\dev.ps1 -Restart
```

С desktop-приложением:

```powershell
.\scripts\dev.ps1 -Restart -Desktop
```

---

## Что поднимается

| Компонент | Порт | Проверка |
|-----------|------|----------|
| PostgreSQL (platform) | 5433 | `docker compose ps` |
| PostgreSQL (security) | 5434 | `docker compose ps` |
| Chat API HTTP | 8080 | `curl http://localhost:8080/health` |
| Chat API gRPC | 50051 | desktop подключается сюда |
| Центр Безопасности HTTP | 8081 | `curl http://localhost:8081/health` |
| Центр Безопасности gRPC | 50052 | ingest от Chat API |

**Admin UI:** http://localhost:8081/admin/  
**Токен:** `dev-admin-token` (из `.env` → `SECURITY_ADMIN_TOKEN`)

---

## Первичная настройка (один раз)

```powershell
cd D:\Goalang\apci-app
Copy-Item .env.example .env
```

Нужно: [Go 1.22+](https://go.dev/dl/), [Docker Desktop](https://www.docker.com/products/docker-desktop/), для desktop — [Wails v2](https://wails.io/docs/gettingstarted/installation).

---

## Админка: экраны

| Вкладка | Назначение |
|---------|------------|
| **Обзор** | Счётчики (клоны, связи), лента рисков, граф активности за 24ч |
| **Устройства** | Реестр хешей устройств (мультиаккаунты) |
| **Связи** | Таблица всех связей между аккаунтами |
| **Расследование** | Поиск по UUID или хешу устройства |
| **Сравнение** | Два аккаунта: общие устройства |

---

## Desktop (мессенджер)

После `.\scripts\dev.ps1` в **отдельном** терминале:

```powershell
cd D:\Goalang\apci-app\desktop
wails dev
```

| Действие | Как |
|----------|-----|
| Регистрация | «Нет аккаунта?» → username → «Зарегистрироваться» |
| Вход | username → «Войти» |
| Демо антифрода | Два аккаунта с одного ПК → перелогин → смотреть **Связи** в админке |

Chat API шлёт события в Центр, если в `.env` задано `SECURITY_CENTER_ADDR=localhost:50052`.

---

## Частые ошибки

| Симптом | Решение |
|---------|---------|
| `DATABASE_URL is required` | Запускайте через `.\scripts\dev.ps1`, не голый `go run` |
| Порт занят | `.\scripts\dev.ps1 -Restart` |
| Пустая админка | Security Center не запущен или desktop не логинился после старта API |
| `docker compose` failed | Запустите Docker Desktop |

---

## Сброс данных

```powershell
docker compose down -v
.\scripts\dev.ps1 -Restart
```

Ключи desktop: `%AppData%\apci-desktop\keys\`

---

## Дальше

Архитектура, модули, демо-сценарии — в [docs/README.md](./README.md) (для разработчиков, не для ежедневного запуска).
