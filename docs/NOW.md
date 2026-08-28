# NOW

Один активный кусок. Всё остальное — не сейчас.
Очередь и инвентарь: [TODO.md](./TODO.md). Пробелы: [GAPS.md](./GAPS.md).

## Срез (2026-08-28) — runtime kernel

- `cmd/runtime` + `cfgs/runtime.json`
- machines list/create/get/stop
- без Docker exec (только metadata; `RUNTIME_DOCKER_HOST` опционален)

Ранее — agents provider + tool loop:

- OpenAI-compatible `AGENTS_LLM_*` (пустой ключ → stub)
- tools: `memory_search` / `memory_put` / `skill_list` / `work_get_issue`
- skills + memory в system prompt

Ранее — media S3 / signed URL / attach:

Закрыто:

- FS default, S3 если `MEDIA_S3_BUCKET` (SigV4, без AWS SDK)
- `DELETE /api/spaces/{spaceId}/blobs/{blobId}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}/url` — presign или HMAC
- content принимает `?exp=&sig=` без JWT
- comment/message `blobId` только из того же space

## Не делаем сейчас

- Telegram / Discord / Slack адаптеры (`integ`)
- Docker start/exec / machine-gateway
- cycles / projects / views / inbox
- Svelte UI
- второй HTTP-listen

## Следом

После разбора [GAPS.md](./GAPS.md) — один пункт.
