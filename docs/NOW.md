# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — media S3 / signed URL / attach

Закрыто:

- FS default, S3 если `MEDIA_S3_BUCKET` (СigV4, без AWS SDK)
- `DELETE /api/spaces/{spaceId}/blobs/{blobId}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}/url` — presign или HMAC
- content принимает `?exp=&sig=` без JWT
- comment/message `blobId` только из того же space

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- Dev Machines / Docker socket (`runtime`)
- LLM-провайдеры и tool loop (`agents`)
- cycles / projects / views / inbox
- Svelte UI
- второй HTTP-listen

## Следом

После разбора [GAPS.md](./GAPS.md) — один пункт.
