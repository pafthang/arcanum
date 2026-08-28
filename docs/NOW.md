# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь галок: [TODO.md](./TODO.md).

## Делаем

`services/comms` — продуктовый чат Space. Треды Kuayle + inbound/agent post GoClaw.
Внешние мессенджеры не в comms (`integ` later). Issue-комментарии остаются в `work`.

Нужно:

- каркас `cmd/comms` + `cfgs/comms.json` + store + client
- `GET/POST /api/spaces/{spaceId}/channels`
- `GET /api/spaces/{spaceId}/channels/{channelId}`
- `GET/POST /api/spaces/{spaceId}/channels/{channelId}/messages`
- WS catalog `/api/spaces/{spaceId}/channels/{channelId}/ws` → `events.comms.{spaceId}.{channelId}`
- internal get/list/create + inbound ingest
- consume `events.integ.message.inbound`
- membership через `space` client (как `work`)

## Не делаем

- Telegram/Discord/Slack адаптеры (`integ`)
- байты вложений (`media`)
- issue comments (`work`)
- analytics / agents runtime
- второй HTTP-listen

## Готово когда

- `go test ./services/comms/...`
- `go build ./cmd/comms`

## Следом

`integ` inbound connectors. `agents` пишет в канал через `services/comms/client`.
