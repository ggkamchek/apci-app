# APCI Desktop

Минимальный UI: экран входа / регистрации и главный экран с выходом.

**Полная инструкция по запуску:** [docs/getting-started.md](../docs/getting-started.md)

---

## Запуск

API должен быть запущен (см. getting-started). Затем:

```powershell
cd D:\Goalang\apci-app\desktop
wails dev
```

Готовый exe: `build\bin\apci-desktop.exe` (сборка: `wails build`).

---

## Использование

1. **Регистрация** — «Нет аккаунта?» → username → «Зарегистрироваться»
2. **Вход** — username → «Войти» (только на ПК, где регистрировались)
3. **Главный экран** — приветствие и ID пользователя
4. **Выйти** — возврат на экран входа

Ключи: `%AppData%\apci-desktop\keys\`

gRPC по умолчанию: `localhost:50051` (переменная `GRPC_ADDR`).
