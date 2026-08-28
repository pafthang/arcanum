# Services

Живой контракт HTTP: `GET /_catalog` на gate, не эти страницы.

Статусы: **live** — процесс в репо и `cfgs/`. **now** — единственный активный кусок, [NOW.md](../NOW.md). **planned** — есть решение, кода нет, не начинать.

| Service   | Role                             | Order | Status  | Doc                        |
| --------- | -------------------------------- | ----- | ------- | -------------------------- |
| `nats`    | Embedded NATS + JetStream        | 0     | live    | [nats.md](./nats.md)       |
| `gate`    | HTTP/WS edge, catalog, auth      | 10    | live    | [gate.md](./gate.md)       |
| `space`   | Users, spaces, teams, principals | 50    | live    | [space.md](./space.md)     |
| `work`    | Issues / work items              | 60    | live    | [work.md](./work.md)       |
| `agents`  | Agent run, session, memory       | 70    | planned | [agents.md](./agents.md)   |
| `comms`   | Channels, messages               | 80    | now     | [comms.md](./comms.md)     |
| `integ`   | External connectors              | 85    | planned | [integ.md](./integ.md)     |
| `media`   | Blobs / files                    | 90    | planned | [media.md](./media.md)     |
| `runtime` | Dev machines, Docker socket      | 100   | planned | [runtime.md](./runtime.md) |
| `logg`    | Logs + activity                  | 110   | live    | [logg.md](./logg.md)       |
| `ctrl`    | Lifecycle + supervisor           | 120   | live    | [ctrl.md](./ctrl.md)       |

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

Layout: [services-structure.md](../services-structure.md). Решения: [DECISIONS.md](../DECISIONS.md).
