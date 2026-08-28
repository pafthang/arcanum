# `comms`

Каналы и сообщения внутри Space. Это продуктовый чат, не NATS и не activity log.

|         |                                         |
| ------- | --------------------------------------- |
| Package | `services/comms`                        |
| Entry   | `go run ./cmd/comms` / `task run:comms` |
| Config  | `cfgs/comms.json`                       |
| Data    | `data/comms/comms.db`                   |
| Order   | 80                                      |
| Status  | now                                     |
| Class   | domain                                  |

## Responsibilities

- Каналы / треды / сообщения, привязанные к space и опционально к team
- Fan-out в браузер через gate WS (`kind=ws` в catalog)
- Участники — principals из `space` (user или agent)
- Вложения — ссылка на `media` (`blob_id`), не байты в comms DB
- Внешние мессенджеры не source of truth; inbound приходит из `integ` событием

## Не владеет

- шиной платформы — `nats`
- audit/activity лентой — `logg`
- issue-комментариями как агрегатом задачи — `work`
- OAuth/webhook коннекторами — `integ`

## Dependencies

| Direction | Target         | Why                 |
| --------- | -------------- | ------------------- |
| requires  | `nats`         | bus + WS subjects   |
| requires  | `space`        | membership, team    |
| refs      | `media`        | вложения            |
| consumes  | `integ` events | inbound извне       |
| used by   | `agents`       | агент пишет в канал |
| logs      | `logg`         | audit               |

## Public surface

Живой список — catalog после реализации. Черновик:

| Kind     | Notes                                                          |
| -------- | -------------------------------------------------------------- |
| HTTP     | `GET/POST /api/spaces/{spaceId}/channels`                      |
| HTTP     | `GET/POST /api/spaces/{spaceId}/channels/{channelId}/messages` |
| WS       | subscribe/publish канала через gate                            |
| Client   | `services/comms/client`                                        |
| Subjects | `public.comms.*`, `internal.comms.*`, `events.comms.*`         |

Примеры subjects (в `pkg/subjects` в ходе реализации):

- `public.comms.channel.list` / `create`
- `public.comms.message.list` / `create`
- `events.comms.message.created`
- WS: subscribe `events.comms.{spaceId}.{channelId}` (уточнить при коде)

Auth на HTTP — required. WS — как у gate: `Authorization` или `?access_token=`.

## Data

SQLite `data/comms/`. Схема только comms.

Набросок (не миграция):

- `channels` — id, space_id, team_id nullable, name, created_at
- `messages` — id, channel_id, actor_id, body, created_at
- вложение через `blob_id`

## Boundaries

- Нет своего HTTP/WS сервера рядом с gate
- Нет импорта `services/comms/internal`
- Нет записи в store `work` / `integ`
- Сообщение без membership в space — отказ (проверка через `space` client)

## Ops

- Старт: `cfgs/comms.json` + `ctrl -up` / `go run ./cmd/comms`
- Логи: `data/comms/service.log`

## See also

- [gate.md](./gate.md) (WS)
- [integ.md](./integ.md)
- [space.md](./space.md)
- [DECISIONS.md](../DECISIONS.md)
