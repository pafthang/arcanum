# `integ`

Внешние коннекторы Space. Источник inbound из мессенджеров и GitHub; outbound workspace webhooks.

|         |                                         |
| ------- | --------------------------------------- |
| Package | `services/integ`                        |
| Entry   | `go run ./cmd/integ` / `task run:integ` |
| Config  | `cfgs/integ.json`                       |
| Data    | `data/integ/integ.db`                   |
| Order   | 85                                      |
| Status  | live                                    |
| Class   | domain                                  |

## Responsibilities

Совмещает поверхность Kuayle и GoClaw, не копируя их монолиты.

Из **Kuayle** (оставить):

- workspace outbound webhooks (подписка на доменные события, HMAC-доставка)
- GitHub App / repo link: connector `github`, список репозиториев
- inbound GitHub webhook → нормализованное `events.integ.github.activity`
- выделение issue-ключей (`ABC-123`, `#12`) из branch / PR title / commit

Из **GoClaw** (оставить):

- канальные коннекторы: `telegram`, `discord`, `slack`, `zalo`, `zalo_oa`, `feishu`, `whatsapp`
- inbound hook с Bearer или HMAC (`hook`)
- секрет коннектора хранится в store сервиса, в API после create не возвращается
- внешний trigger не пишет в `comms`/`work` напрямую — только events / commands

## Не владеет

- продуктовый чат — `comms` (inbound приходит событием)
- issue-агрегат — `work`
- run / session / LLM-провайдеры — `agents`
- Docker socket — `runtime`
- principal / JWT — `space` + `gate`

## Dependencies

| Direction | Target        | Why                                      |
| --------- | ------------- | ---------------------------------------- |
| requires  | `nats`        | bus + public subjects                    |
| requires  | `space`       | membership на mutate API                 |
| consumes  | `work` events | fan-out в outbound webhooks              |
| emits     | `comms`       | `events.integ.channel.message`           |
| emits     | `work`/`agents` | `events.integ.github.activity`         |
| logs      | `logg`        | audit                                    |

## Public surface

Живой список — `GET /_catalog`. Зафиксировано в коде:

| Kind   | Path / subject |
| ------ | -------------- |
| HTTP   | `GET/POST /api/spaces/{spaceId}/integ/connectors` |
| HTTP   | `GET/PATCH /api/spaces/{spaceId}/integ/connectors/{connectorId}` |
| HTTP   | `GET/POST /api/spaces/{spaceId}/integ/connectors/{connectorId}/repos` |
| HTTP   | `GET/POST /api/spaces/{spaceId}/integ/webhooks` |
| HTTP   | `GET /api/spaces/{spaceId}/integ/deliveries` |
| HTTP   | `POST /api/integ/hooks/{connectorId}` auth none |
| HTTP   | `POST /api/integ/hooks/github/{connectorId}` auth none |
| Client | `services/integ/client` |

## Subjects

- `public.integ.connector.list` / `create` / `get` / `update`
- `public.integ.repo.list` / `create`
- `public.integ.webhook.list` / `create`
- `public.integ.delivery.list`
- `public.integ.hook.ingest` / `public.integ.hook.github`
- `internal.integ.connector.get` / `list`
- `internal.integ.ingest`
- `events.integ.inbound`
- `events.integ.github.activity`
- `events.integ.channel.message`
- `events.integ.webhook.delivered`

## Data

SQLite `data/integ/`. Схема не шарится.

- `connectors` — id, space_id, kind, name, status, config_json, secret, created_at, updated_at
- `github_repos` — id, connector_id, space_id, owner, name, installation_id, created_at
- `webhooks` — id, space_id, url, secret, events_json, active, created_at
- `deliveries` — исходящие попытки
- `inbound_events` — принятый raw + нормализация

`kind`: `github`, `hook`, `telegram`, `discord`, `slack`, `zalo`, `zalo_oa`, `feishu`, `whatsapp`.

## Boundaries

- Нет своего HTTP-порта (inbound тоже через gate catalog)
- Нет импорта `services/integ/internal`
- Нет записи в store `work` / `comms` / `agents`
- Сырой секрет не отдавать в list/get
- LLM keys сюда не класть

## Ops

- Старт: `cfgs/integ.json` + `ctrl -up`
- Логи: `data/integ/service.log`

## See also

- [comms.md](./comms.md)
- [work.md](./work.md)
- [agents.md](./agents.md)
- [DECISIONS.md](../DECISIONS.md)
