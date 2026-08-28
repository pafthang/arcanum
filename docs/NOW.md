# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — space public + work fields

Закрыто по приказу «добить backend из TODO»:

- `space`: register, switch-space, invite, update member, teams HTTP, API keys / agent principal, `internal.space.can`
- `work`: priority, due, parent, extra assignees, relations
- `agents` consume `events.work.issue.assigned` — уже было в коде
- `comms` membership check на mutate — уже было в коде

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
