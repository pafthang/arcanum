# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — API key login

Закрыто:

- `api_keys.key_hash` = SHA-256 секрета (`passwd.KeyHash`), не argon2
- `POST /api/auth/api-key` `{secret}` → JWT как login
- пароль агента по-прежнему argon2id (email+secret login тоже работает)

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- байты вложений (`media`)
- Dev Machines / Docker socket (`runtime`)
- LLM-провайдеры и tool loop (`agents`)
- cycles / projects / views / inbox
- Svelte UI
- второй HTTP-listen

## Следом

После разбора [GAPS.md](./GAPS.md) — один пункт. Кандидаты:

- UI-клиент на gate
- `agents` один OpenAI-compatible provider
- `integ` один живой канал или GitHub App
- `media` blob store
