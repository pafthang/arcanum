# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — media kernel

Закрыто:

- `cmd/media` + `cfgs/media.json` order 90
- SQLite metadata + files under `data/media/blobs/{spaceId}/{id}`
- `GET/POST /api/spaces/{spaceId}/blobs`
- `GET /api/spaces/{spaceId}/blobs/{blobId}` metadata
- `GET /api/spaces/{spaceId}/blobs/{blobId}/content` raw bytes
- internal `internal.media.get` / `internal.media.get_bytes`
- typed `services/media/client`
- лимит `MEDIA_MAX_BYTES` (default 1MiB — NATS payload)

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- S3 backend
- Dev Machines / Docker socket (`runtime`)
- LLM-провайдеры и tool loop (`agents`)
- cycles / projects / views / inbox
- Svelte UI
- второй HTTP-listen

## Следом

После разбора [GAPS.md](./GAPS.md) — один пункт.
