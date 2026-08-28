# Arcanum architecture

Monorepo: independent Go processes (NATS mini) + HTTP/WS edge (`gate`).

Ориентиры продукта (не корни репо): Kuayle workspace, GoClaw agents.
См. [GAPS.md](./GAPS.md), [DECISIONS.md](./DECISIONS.md).

## Model

| Piece           | Role |
| --------------- | ---- |
| Processes       | `services/<name>` + entry `cmd/<name>` |
| Bus             | NATS (+ JetStream) |
| Public edge     | `gate` only |
| Contract        | `GET /_catalog` |
| Tenancy         | Workspace = **Space** (JWT `space_id` / `space_role`). Nested **Teams** group users/agents inside a space. Service: `services/space`. |
| Roles           | `platform_admin` is a user platform flag — no separate admin entity |
| Shared Go       | `pkg/*` |
| Browser         | gate only |
| Service→service | `services/<name>/client` + `pkg/subjects` |

```
Browser / Svelte
      |  HTTPS + WS + JWT
      v
  gate --public.*--> domain services
      |
      |  internal.* / events.* / commands.*
      v
  service <-- typed client --> service
```

## Service map

| Service   | Role                                   | Order | Status        | Doc |
| --------- | -------------------------------------- | ----- | ------------- | --- |
| `nats`    | Embedded NATS + JetStream              | 0     | live          | [services/nats.md](./services/nats.md) |
| `gate`    | HTTP/WS edge, catalog, auth middleware | 10    | live          | [services/gate.md](./services/gate.md) |
| `space`   | Identity / tenancy                     | 50    | live / kernel | [services/space.md](./services/space.md) |
| `work`    | Issue aggregate                        | 60    | live / kernel | [services/work.md](./services/work.md) |
| `agents`  | Run / session / memory / skills        | 70    | live / kernel | [services/agents.md](./services/agents.md) |
| `comms`   | Channels / messages                    | 80    | live / kernel | [services/comms.md](./services/comms.md) |
| `integ`   | External connectors                    | 85    | live / kernel | [services/integ.md](./services/integ.md) |
| `media`   | Blobs                                  | 90    | planned       | [services/media.md](./services/media.md) |
| `runtime` | Machines, Docker socket                | 100   | planned       | [services/runtime.md](./services/runtime.md) |
| `logg`    | Logs + activity                        | 110   | live          | [services/logg.md](./services/logg.md) |
| `ctrl`    | Lifecycle + supervisor                 | 120   | live          | [services/ctrl.md](./services/ctrl.md) |

Index: [README.md](./README.md). Сделано по коду: [TODO.md](./TODO.md).

## Boundaries

- Domain services do **not** listen on public HTTP.
- Do not import another service’s `services/*/internal`.
- One service owns its SQLite tables; others use RPC/events.
- Docker socket только у `runtime`.

## Data

- Default: SQLite under `data/<name>/`.
- Blobs: `data/files/blobs/` (`media`, planned).
- JetStream: `data/nats/jetstream/`.

## Non-goals

- A second public HTTP edge next to gate (исключение: `machine-gateway` у Dev Machines)
- Shared DB across services
- A separate admin microservice
- Отдельный dashboard агентов

## Decisions

[DECISIONS.md](./DECISIONS.md). ADR-каталога в репо пока нет.
