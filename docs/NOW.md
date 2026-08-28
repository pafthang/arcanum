# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез закрыт (2026-08-28)

`services/comms` kernel закрыт по чеклисту C1:

- `cmd/comms` + `cfgs/comms.json`
- store channels / messages / threads / blob_id / integ inbound
- public HTTP + WS catalog
- internal RPC + ingest
- `go test ./services/comms/...`, `go build ./cmd/comms`

Не открывать новый код, пока gaps не приоритизированы в action plan.

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- байты вложений (`media`)
- Dev Machines / Docker socket (`runtime`)
- LLM-провайдеры и tool loop (`agents`)
- cycles / projects / sub-issues (`work`)
- Svelte UI
- второй HTTP-listen

## Следом

После разбора [GAPS.md](./GAPS.md) — один пункт в этот файл. Кандидаты (не очередь):

- `space` public: invite / teams / api keys / switch space
- `agents` реальный pipeline + provider
- `integ` один живой канал (Telegram) или GitHub App не-заглушка
- `media` blob store
- UI-клиент на gate
