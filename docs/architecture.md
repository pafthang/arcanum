# Arcanum architecture

Monorepo: independent Go processes (NATS mini) + HTTP/WS edge (`gate`).

## Model

| Piece           | Role                                                                                                                                                                                     |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Processes       | `services/<name>` + entry `cmd/<name>`                                                                                                                                                   |
| Bus             | NATS (+ JetStream)                                                                                                                                                                       |
| Public edge     | `gate` only                                                                                                                                                                              |
| Contract        | `GET /_catalog`                                                                                                                                                                          |
| Tenancy         | Workspace = **Space** (JWT `space_id` / `space_role`). Nested **Teams** group users/agents inside a space — [ADR 0004](./adr/0004-space-and-nested-teams.md). Service: `services/space`. |
| Roles           | `platform_admin` is a user platform role — no separate admin entity                                                                                                                      |
| Shared Go       | `pkg/*`                                                                                                                                                                                  |
| Browser         | gate only                                                                                                                                                                                |
| Service→service | `services/<name>/client` + `pkg/subjects`                                                                                                                                                |

```
Browser / Svelte
      │  HTTPS + WS + JWT
      ▼
  gate ──public.*──► domain services
      │
      │  internal.* / events.* / commands.*
      ▼
  service ◄── typed client ──► service
```

## Service map

| Service   | Role                                   | Order | Status  | Doc                                          |
| --------- | -------------------------------------- | ----- | ------- | -------------------------------------------- |
| `nats`    | Embedded NATS + JetStream              | 0     | live    | [services/nats.md](./services/nats.md)       |
| `gate`    | HTTP/WS edge, catalog, auth middleware | 10    | live    | [services/gate.md](./services/gate.md)       |
| `space`   | Identity / tenancy                     | 50    | now     | [services/space.md](./services/space.md)     |
| `work`    | Issue aggregate                        | 60    | planned | [services/work.md](./services/work.md)       |
| `agents`  | Run / session / memory                 | 70    | planned | [services/agents.md](./services/agents.md)   |
| `comms`   | Channels / messages                    | 80    | planned | [services/comms.md](./services/comms.md)     |
| `integ`   | External connectors                    | 85    | planned | [services/integ.md](./services/integ.md)     |
| `media`   | Blobs                                  | 90    | planned | [services/media.md](./services/media.md)     |
| `runtime` | Machines, Docker socket                | 100   | planned | [services/runtime.md](./services/runtime.md) |
| `logg`    | Logs + activity                        | 110   | live    | [services/logg.md](./services/logg.md)       |
| `ctrl`    | Lifecycle + supervisor                 | 120   | live    | [services/ctrl.md](./services/ctrl.md)       |

Index: [services/README.md](./services/README.md). Order/env: `cfgs/<name>.json`. Layout: [services-structure.md](./services-structure.md).

## Boundaries

- Domain services do **not** listen on public HTTP.
- Do not import another service’s `services/*/internal`.
- One service owns its SQLite tables; others use RPC/events.
- Shared logic lives in `pkg/`, not copy-pasted across services.
- Docker socket только у `runtime`.

## Data

- Default: SQLite under `data/<name>/`.
- Blobs: `data/files/blobs/` (`media`).
- JetStream: `data/nats/jetstream/`.
- Runtime data is gitignored.

## Non-goals

- A second public HTTP edge next to gate (исключение: `machine-gateway` у Dev Machines)
- Shared DB across services
- gRPC/HTTP mesh instead of NATS mini (current model)
- A separate admin microservice
- Отдельный dashboard агентов

## Decisions

- [ADR index](./adr/README.md)
- [DECISIONS.md](./DECISIONS.md)

## References

- [services-structure.md](./services-structure.md)
- [API policy](./api/README.md)
- Root [README](../README.md)