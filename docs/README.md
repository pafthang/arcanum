# Services

Живой контракт HTTP: `GET /_catalog` на gate, не эти страницы.

Статусы:

- **live** — процесс в репо (`cmd/`, `cfgs/`, `services/`), поверхность в catalog
- **kernel** — live, но без продуктовой полноты Kuayle / GoClaw (см. [GAPS.md](./GAPS.md))
- **planned** — есть решение, `cmd/`/`cfgs/` нет, не начинать вне [NOW.md](./NOW.md)
- **next** — следующий кусок кода только после приоритизации gaps

| Service   | Role                             | Order | Status        | Doc                          |
| --------- | -------------------------------- | ----- | ------------- | ---------------------------- |
| `nats`    | Embedded NATS + JetStream        | 0     | live          | [nats.md](./services/nats.md)     |
| `gate`    | HTTP/WS edge, catalog, auth      | 10    | live          | [gate.md](./services/gate.md)     |
| `space`   | Users, spaces, teams, principals | 50    | live / kernel | [space.md](./services/space.md)   |
| `work`    | Issues / labels / comments       | 60    | live / kernel | [work.md](./services/work.md)     |
| `agents`  | Run, session, memory, skills     | 70    | live / kernel | [agents.md](./services/agents.md) |
| `comms`   | Channels, messages, WS catalog   | 80    | live / kernel | [comms.md](./services/comms.md)   |
| `integ`   | Connectors, hooks, webhooks      | 85    | live / kernel | [integ.md](./services/integ.md)   |
| `media`   | Blobs / files                    | 90    | planned       | [media.md](./services/media.md)   |
| `runtime` | Dev machines, Docker socket      | 100   | planned       | [runtime.md](./services/runtime.md) |
| `logg`    | Logs + activity                  | 110   | live          | [logg.md](./services/logg.md)     |
| `ctrl`    | Lifecycle + supervisor           | 120   | live          | [ctrl.md](./services/ctrl.md)     |

Срез «что закрыто / чего нет»: [TODO.md](./TODO.md).
Сравнение с Kuayle и GoClaw: [GAPS.md](./GAPS.md).
Активный кусок реализации: [NOW.md](./NOW.md).

Имена из старого черновика (`auth`, `task`, `agent`, `files`, `exec`, `cron`, `memo`, `pipe`, `conn`) **не сервисы**. Соответствия:

| Черновик        | Куда                                       |
| --------------- | ------------------------------------------ |
| `auth`          | `space` (login, principals) + `gate` (JWT) |
| `task`          | `work`                                     |
| `agent`         | `agents`                                   |
| `files`         | `media`                                    |
| `conn` / `pipe` | `integ` / `comms`                          |
| `exec` / `cron` | `runtime` + later workers                  |
| `memo`          | `agents` memory                            |

Layout: [services-structure.md](./services-structure.md). Решения: [DECISIONS.md](./DECISIONS.md). Архитектура: [architecture.md](./architecture.md).
