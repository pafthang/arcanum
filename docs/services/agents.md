# `agents`

Исполнение агентов: run, session, memory. Principal агента выдаёт `space`, не этот сервис.

|         |                                                               |
| ------- | ------------------------------------------------------------- |
| Package | `services/agents`                                             |
| Entry   | `go run ./cmd/agents` / `task run:agents`                     |
| Config  | `cfgs/agents.json`                                            |
| Data    | `data/agents/agents.db`                                       |
| Order   | 70                                                            |
| Status  | planned — после каркаса `work`, не вместо [NOW.md](../NOW.md) |
| Class   | domain                                                        |

## Responsibilities

- Владеет run: старт, статус, выход, привязка к issue и space
- Владеет session (контекст одного прогона) и memory (долгоживущая память агента в space)
- Стартует по команде `commands.agents.run.start`, обычно после `events.work.issue.assigned`
- Агент — отдельный principal: `actor=agent` + API key hash в `space`, не пароль человека и не shared user
- Может просить `runtime` поднять machine; Docker socket сам не монтирует

## Не владеет

- issue-агрегатом — `work`
- регистрацией пользователя/агента как аккаунта — `space`
- Docker socket — `runtime`
- каналом чата с человеком — `comms` (агент может писать туда через client)

## Dependencies

| Direction | Target        | Why                               |
| --------- | ------------- | --------------------------------- |
| requires  | `nats`        | bus, commands                     |
| requires  | `space`       | agent principal, space membership |
| consumes  | `work` events | assigned → start run              |
| may call  | `runtime`     | песочница                         |
| may call  | `media`       | артефакты рана                    |
| may call  | `comms`       | сообщения в канал                 |
| logs      | `logg`        | audit + activity                  |

## Public surface

Живой список путей — catalog после реализации. Черновик:

| Kind     | Notes                                                                          |
| -------- | ------------------------------------------------------------------------------ |
| HTTP     | `GET/POST /api/spaces/{spaceId}/runs`                                          |
| HTTP     | `GET /api/spaces/{spaceId}/runs/{runId}`                                       |
| HTTP     | session / memory space-scoped                                                  |
| Client   | `services/agents/client`                                                       |
| Subjects | `public.agents.*`, `internal.agents.*`, `commands.agents.*`, `events.agents.*` |

Примеры subjects (в `pkg/subjects` только в ходе реализации):

- `commands.agents.run.start`
- `commands.agents.run.cancel`
- `public.agents.run.list` / `get`
- `internal.agents.run.get`
- `events.agents.run.started`
- `events.agents.run.finished`

## Data

SQLite `data/agents/`. Схема не шарится.

Набросок (не миграция):

- `runs` — id, space_id, issue_id, agent_id, status, started_at, finished_at
- `sessions` — id, run_id, payload
- `memories` — id, space_id, agent_id, key, value, updated_at

Сырой API key здесь не хранить.

## Boundaries

- Не импортировать `services/agents/internal` снаружи
- Нет своего HTTP-порта
- Нет прямой записи в store `work` (статус issue — событием/client work)
- Нет Docker socket
- Старт рана без валидного agent principal из `space` — отказ

## Ops

- Старт: `cfgs/agents.json` + `ctrl -up`
- Логи: `data/agents/service.log`
- Пока planned — cfg/cmd не обязательны

## See also

- [work.md](./work.md)
- [space.md](./space.md)
- [runtime.md](./runtime.md)
- [DECISIONS.md](../DECISIONS.md)