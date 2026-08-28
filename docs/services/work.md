# `work`

Issue-агрегат платформы. Один work item живёт в Space; исполнение агента work не хранит.

|         |                                                                         |
| ------- | ----------------------------------------------------------------------- |
| Package | `services/work`                                                         |
| Entry   | `go run ./cmd/work` / `task run:work`                                   |
| Config  | `cfgs/work.json`                                                        |
| Data    | `data/work/work.db`                                                     |
| Order   | 60                                                                      |
| Status  | planned — код не начинать, пока не закрыт [NOW.md](../NOW.md) (`space`) |
| Class   | domain                                                                  |

## Responsibilities

- Владеет агрегатом issue: id, space, title, body, status, assignee, labels
- Назначение человека или агента (`actor` из `space`)
- Комментарии и смена статуса — часть агрегата work, не comms-канал
- Публикует доменные события, не вызывает `agents` напрямую через чужой store
- Связь с агентами: `events.work.issue.assigned` → `commands.agents.run.start`

## Не владеет

- run / session / memory агента — `agents`
- файлами вложений — `media` (`blob_id` в issue)
- каналами чата — `comms`
- principal / JWT — `space` + `gate`

## Dependencies

| Direction | Target   | Why                              |
| --------- | -------- | -------------------------------- |
| requires  | `nats`   | bus                              |
| requires  | `space`  | space_id, member/agent principal |
| emits     | `agents` | команда старта рана              |
| refs      | `media`  | вложения                         |
| logs      | `logg`   | activity                         |

## Public surface

Живой список путей — только catalog после реализации. Черновик:

| Kind     | Notes                                               |
| -------- | --------------------------------------------------- |
| HTTP     | `GET/POST /api/spaces/{spaceId}/issues`             |
| HTTP     | `GET/PATCH /api/spaces/{spaceId}/issues/{issueId}`  |
| HTTP     | назначение, статус, комментарии агрегата            |
| Client   | `services/work/client`                              |
| Subjects | `public.work.*`, `internal.work.*`, `events.work.*` |

Примеры subjects (зафиксировать в `pkg/subjects` в ходе реализации, не раньше):

- `public.work.issue.list` / `get` / `create` / `update`
- `internal.work.issue.get`
- `events.work.issue.created`
- `events.work.issue.assigned`
- `events.work.issue.updated`

## Data

SQLite `data/work/`. Схема не копируется в другие сервисы.

Минимальный набросок (не миграция):

- `issues` — id, space_id, title, body, status, assignee_id, created_at, updated_at
- `issue_comments` — id, issue_id, actor_id, body, created_at
- вложение: `blob_id` → `media`, не BLOB в этой БД

## Boundaries

- Чужие пакеты не импортят `services/work/internal`
- Нет своего HTTP-порта
- Нет записи в store `agents` / `space`
- Assignee без membership в space — отказ, проверка через `space` client

## Ops

- Старт: `cfgs/work.json` подхватит `ctrl -up`
- Логи: `data/work/service.log`
- Пока статуса planned — cfg/cmd в репо не обязательны

## See also

- [agents.md](./agents.md)
- [space.md](./space.md)
- [media.md](./media.md)
- [DECISIONS.md](../DECISIONS.md)
- [services-structure.md](../services-structure.md)