# `media`

Blobs / files. Сервиса в дереве нет.

|         |                      |
| ------- | -------------------- |
| Package | `services/media`     |
| Entry   | `cmd/media` — нет    |
| Config  | `cfgs/media.json` — нет |
| Data    | `data/files/blobs/` (план) |
| Order   | 90                   |
| Status  | planned              |
| Class   | domain               |

## Responsibilities (когда начнётся)

- Владеет байтами вложений. `work` / `comms` / `agents` хранят только `blob_id`
- Upload/download через gate catalog, не свой listen
- Локальный FS default; S3-совместимое — позже (Kuayle так умеет)

## Не владеет

- issue/chat текстом
- Docker volumes Dev Machines — `runtime`

## Не начинать

Пока нет строки в [NOW.md](../NOW.md). Ссылки `blob_id` в comms уже есть.

## See also

- [GAPS.md](../GAPS.md) корзина P-files-runtime
- [comms.md](./comms.md)
