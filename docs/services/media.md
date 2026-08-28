# `media`

Blobs / files.

|         |                          |
| ------- | ------------------------ |
| Package | `services/media`         |
| Entry   | `cmd/media`              |
| Config  | `cfgs/media.json`        |
| Data    | `data/media/blobs/{spaceId}/{id}` or S3 `spaceId/id` |
| Order   | 90                       |
| Status  | live / kernel            |
| Class   | domain                   |

## Responsibilities

- Владеет байтами вложений. `work` / `comms` / `agents` хранят только `blob_id`
- Upload/download через gate catalog, не свой listen
- FS default. S3 если `MEDIA_S3_BUCKET` (СigV4, MinIO/R2/AWS)

## Routes

- `GET /api/spaces/{spaceId}/blobs`
- `POST /api/spaces/{spaceId}/blobs` raw / multipart / JSON `{filename,contentType,data}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}`
- `GET /api/spaces/{spaceId}/blobs/{blobId}/content` JWT или `?exp=&sig=`
- `GET /api/spaces/{spaceId}/blobs/{blobId}/url` signed URL
- `DELETE /api/spaces/{spaceId}/blobs/{blobId}`

Internal: `internal.media.get`, `internal.media.get_bytes`.

Env: `MEDIA_MAX_BYTES`, `MEDIA_SIGN_TTL`, `MEDIA_SIGN_SECRET`, `MEDIA_PUBLIC_BASE`, `MEDIA_S3_*`.

`work` comment и `comms` message отклоняют `blobId` из другого space.

## Не владеет

- issue/chat текстом
- Docker volumes Dev Machines — `runtime`
