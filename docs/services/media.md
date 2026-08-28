# `media`

Blobs / files.

|         |                          |
| ------- | ------------------------ |
| Package | `services/media`         |
| Entry   | `cmd/media`              |
| Config  | `cfgs/media.json`        |
| Data    | `data/media/blobs/{spaceId}/{id}` |
| Order   | 90                       |
| Status  | live / kernel            |
| Class   | domain                   |

## Responsibilities

- Владеет байтами вложений. `work` / `comms` / `agents` хранят только `blob_id`
- Upload/download через gate catalog, не свой listen
- Локальный FS default; S3 — позже

## Routes

- `GET /api/spaces/{spaceId}/blobs`
- `POST /api/spaces/{spaceId}/blobs` raw / multipart / JSON `{filename,contentType,data}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}/content`

Internal: `internal.media.get`, `internal.media.get_bytes`.

Лимит: `MEDIA_MAX_BYTES` (default 1MiB, NATS payload).

## Не владеет

- issue/chat текстом
- Docker volumes Dev Machines — `runtime`
