# Запуск проекта локально

Пошаговая инструкция: открыли репозиторий в VS Code → подняли Docker → проверили API → открыли desktop-приложение.

**Что тестируем:** регистрация и вход через интерфейс APCI (Wails).  
**Транспорт:** gRPC `:50051`, HTTP `:8080` — только health.

---

## Что нужно установить

| Инструмент | Зачем |
|------------|--------|
| [Go](https://go.dev/dl/) 1.22+ | API и desktop |
| [Docker Desktop](https://www.docker.com/products/docker-desktop/) | PostgreSQL |
| [Wails v2](https://wails.io/docs/gettingstarted/installation) | desktop-приложение |
| WebView2 | обычно уже есть в Windows 10/11 |

Проверка Wails:

```powershell
wails doctor
```

---

## 1. Первичная настройка (один раз)

Откройте папку репозитория в VS Code, терминал — **PowerShell**.

```powershell
cd D:\Goalang\apci-app
Copy-Item .env.example .env
```

В `.env` должны быть (по умолчанию уже так):

| Переменная | Значение |
|------------|----------|
| `POSTGRES_PORT` | `5433` |
| `DATABASE_URL` | `postgres://platform:platform@localhost:5433/platform?sslmode=disable` |
| `GRPC_ADDR` | `:50051` |
| `SERVER_ADDR` | `:8080` |

---

## 2. PostgreSQL (терминал 1)

```powershell
cd D:\Goalang\apci-app
docker compose up -d postgres
docker compose ps
```

Контейнер `platform-postgres` должен быть **healthy**.

---

## 3. API (тот же или новый терминал)

API читает переменные из окружения. Загрузите `.env`, затем запустите сервер:

```powershell
cd D:\Goalang\apci-app
Get-Content .env | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.*)$') { Set-Item -Path "env:$($matches[1].Trim())" -Value $matches[2].Trim() }
}
cd server
go run ./cmd/api
```

В логах должно появиться:

```
grpc listening on :50051
http listening on :8080
```

**Терминал с API не закрывайте** — он должен работать всё время.

### Проверка health

Новый терминал:

```powershell
curl http://localhost:8080/health
```

Ожидаемый ответ: `{"status":"ok","database":"ok"}`

---

## 4. Desktop-приложение (терминал 2)

```powershell
cd D:\Goalang\apci-app\desktop
wails dev
```

Через несколько секунд откроется окно **APCI** (480×640).

### Готовый exe (без `wails dev`)

```powershell
cd D:\Goalang\apci-app\desktop
.\build\bin\apci-desktop.exe
```

Если exe нет — соберите один раз: `wails build`.

---

## 5. Как пользоваться интерфейсом

| Действие | Шаги |
|----------|------|
| **Регистрация** | «Нет аккаунта?» → введите username (например `lera`) → «Зарегистрироваться» |
| **Вход** | Введите username → «Войти» (ключи должны быть на этом ПК) |
| **Главный экран** | Приветствие, ID пользователя, «Сессия активна» |
| **Выход** | «Выйти» → снова экран входа |

**Ограничение MVP:** один аккаунт привязан к устройству, где вы регистрировались. Вход с другого ПК — позже (привязка устройств / бэкап ключа).

Ключи хранятся локально:

```
%AppData%\apci-desktop\keys\<username>.key
```

---

## 6. Проверка в БД (опционально)

```powershell
docker exec -it platform-postgres psql -U platform -d platform
```

```sql
SELECT id, username, created_at FROM users;
SELECT u.username, d.device_hash FROM user_devices d JOIN users u ON u.id = d.user_id;
\q
```

---

## 7. Сброс данных (начать с чистого листа)

```powershell
cd D:\Goalang\apci-app
docker compose down -v
docker compose up -d postgres
```

Затем снова `go run ./cmd/api`. Локальные ключи desktop при необходимости удалите:

```
%AppData%\apci-desktop\keys\
```

---

## 8. Частые ошибки и решения

| Симптом | Причина | Решение |
|---------|---------|---------|
| `DATABASE_URL is required` | `.env` не загружен в терминал | Выполните блок `Get-Content .env \| ForEach-Object ...` перед `go run` |
| `database: error` в `/health` | Postgres не запущен или не healthy | `docker compose up -d postgres`, подождите ~10 сек |
| `connection refused` на `:50051` | API не запущен | Терминал с `go run ./cmd/api` |
| Порт **50051** занят | API уже запущен в другом терминале | Не запускайте второй экземпляр; закройте лишний процесс |
| Порт **5433** занят | Другой Postgres на машине | Смените `POSTGRES_PORT` в `.env` |
| Desktop: «нет подключения к серверу» | API остановлен | Запустите API (шаг 3) |
| «имя пользователя уже занято» | Username уже в БД | Войдите или выберите другой username |
| «ключи не найдены на этом устройстве» | Вход без регистрации на этом ПК | Зарегистрируйтесь или используйте ПК, где регистрировались |
| `migrate: read migration` | API запущен не из `server/` | `cd server` перед `go run ./cmd/api` |
| `wails: command not found` | Wails не установлен | [Установка Wails](https://wails.io/docs/gettingstarted/installation) |
| `use of internal package ... not allowed` (desktop) | Старая версия кода | Protobuf должен быть в `server/pkg/pb`, не в `internal/pb` |
| PowerShell: `&&` не работает | Синтаксис bash | Используйте `;` или отдельные команды |

---

## 9. Схема терминалов

```
Терминал 1                    Терминал 2
──────────                    ──────────
docker compose up -d postgres
go run ./cmd/api      ──gRPC──►  wails dev
       ▲                              │
       │                              ▼
  localhost:50051              Окно APCI (UI)
```

---

## Связанные файлы

| Путь | Назначение |
|------|------------|
| `server/cmd/api/main.go` | Точка входа API |
| `desktop/` | Wails-клиент |
| `shared/proto/apci/users/v1/users.proto` | gRPC-контракт auth |
| `server/migrations/001_users.sql` | Таблицы users, sessions |
