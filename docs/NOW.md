# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — issue activity

Закрыто:

- create/update/comment issue пишут activity в `logg` (best-effort)
- `GET /api/spaces/{spaceId}/issues/{issueId}/activity`
- предыдущий glue-срез space/work на месте

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
