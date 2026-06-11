# Интеграция Chat API и Центра Безопасности

---

## Принципы

1. **Однонаправленная связь:** Chat API → Центр Безопасности. Центр не вызывает Chat API.
2. **Асинхронность:** события отправляются fire-and-forget после успешного действия. Пользователь не ждёт ответа Центра.
3. **Отказоустойчивость:** если Центр недоступен, чаты продолжают работать. Ошибки логируются.
4. **Идемпотентность:** каждое событие имеет `event_id` (UUID) для защиты от дублей при retry.
5. **Раздельные БД:** данные чатов и граф связей не смешиваются.

---

## Протокол

| Параметр | Значение |
|----------|----------|
| Транспорт | gRPC |
| Метод | `SecurityIngest.IngestEvent` |
| Контракты | `shared/proto/security/v1/` |

---

## Поток событий

### Регистрация

```
1. Desktop → Chat API: Register
2. Chat API: сохраняет пользователя, отвечает OK
3. Chat API (goroutine): IngestEvent { user.registered, user_id, device_hash }
4. Центр: создаёт узел; при совпадении hash — ребро same_device
```

### Вход

```
1. Desktop → Chat API: Login
2. Chat API: аутентификация, ответ OK
3. Chat API (goroutine): IngestEvent { user.logged_in, ... }
```

### Добавление контакта

```
1. Desktop → Chat API: AddContact
2. Chat API: сохраняет связь, ответ OK
3. Chat API (goroutine): IngestEvent { contact.added, user_id, contact_id }
4. Центр: рёбра shared_contact при пересечениях
```

### Отправка сообщения

```
1. Desktop → Chat API: SendMessage (ciphertext)
2. Chat API: сохраняет ciphertext, ответ OK
3. Chat API (goroutine): IngestEvent { message.sent, from, to, message_id }
4. Центр: ребро communicates_with
```

---

## Что НЕ отправляется в Центр

| Данные / событие | Причина |
|------------------|---------|
| Текст / ciphertext сообщения | E2E, приватность |
| Результаты локального поиска | Только на клиенте |
| Presence (в сети / не в сети) | UX, не нужен для графа в MVP |
| Typing («печатает») | UX, не нужен для графа в MVP |
| Поиск пользователей | Не в scope MVP |

Presence и typing обрабатываются в Chat API и доставляются клиентам через realtime-канал (gRPC stream или WebSocket — на выбор при реализации).

---

## Построение рёбер в Центре

| Тип ребра | Условие | Вес |
|-----------|---------|-----|
| `same_device` | Одинаковый `device_hash` | Высокий |
| `shared_contact` | Пересечение списков контактов | Средний |
| `communicates_with` | Факт переписки (from/to) | Средний |

---

## Обработка ошибок

| Ситуация | Поведение Chat API |
|----------|-------------------|
| Центр недоступен | Запрос пользователя успешен; событие логируется как failed |
| Retry | Roadmap: outbox-таблица в БД чатов + фоновый воркер |
| Дубликат event_id | Центр возвращает OK без повторной записи |

---

## Admin API

Центр Безопасности предоставляет API для админки:

**Ответ:** список связей

| Поле | Описание |
|------|----------|
| account_a | ID первого аккаунта |
| account_b | ID второго аккаунта |
| link_type | Тип связи |
| weight | Вес связи |
| detected_at | Дата обнаружения |

Для простой таблицы на защите допустим HTTP + React; gRPC — для внутренних вызовов.

---

## Docker (целевая схема)

```yaml
services:
  postgres-chat:
  postgres-security:
  chat-api:
  security-center:
  # admin-ui (опционально)
```

Chat API знает адрес Центра через переменную окружения, например `SECURITY_CENTER_ADDR`.
