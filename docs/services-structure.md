# Arcanum service structure

The `services/` tree is a set of independent processes. **One layout: edge-style.**

## Skeleton (all services)
cmd//main.go                 # only calls .Run()
cfgs/.json                   # name, order, env
services//
service.go                       # wiring + process entry (Run)
export.go                        # aliases / Config if the root package exports them
models/                          # public DTOs for client (preferred over leaking internal)
client/                          # typed client for other services
internal/
config/
store/                         # SQLite + migrate; path svcutil.DataDir(name)
apis/                          # mini endpoints
textНе класть entry в `services/<name>/cmd/` и cfg в `services/<name>/cfgs/`.

### Rules

| Rule                    | Description                                                                                     |
| ----------------------- | ----------------------------------------------------------------------------------------------- |
| Entrypoint              | `cmd/<name>/main.go` only calls `Run()`.                                                        |
| Client                  | Cross-service calls go through `services/<name>/client`, not raw NATS subjects.                 |
| `internal/`             | Closed to code **outside** `services/<name>/`.                                                  |
| `export.go` / `models/` | Public types for other packages. Store and apis stay internal.                                  |
| Public subpackages      | Only when another service actually imports them.                                                |
| Config                  | Env + defaults; `cfgs/<name>.json` via `svcutil` / `mini`.                                      |
| Data                    | SQLite via `pkg/sqldb` (`modernc.org/sqlite`). No CGO, no shared DB.                            |
| HTTP API                | Declared with `mini.WithPublicHTTP` + `mini.WithPublicSubject`. Live contract: `GET /_catalog`. |
| Process                 | Domain services do not listen on HTTP. `gate` is the only public edge.                          |

### Root package contract

| Symbol                   | Purpose                                                  |
| ------------------------ | -------------------------------------------------------- |
| `Run()`                  | process entrypoint (bootstrap → store → register → wait) |
| `export.go`              | `Config` / `FromEnv` aliases or domain types if needed   |
| `models/`                | public DTOs imported by `client/` and others             |
| `Service` + `NewService` | only complex processes (`gate`, `ctrl`)                  |
| `client/`                | typed NATS client (`mini.NewClient` + `RequestJSON`)     |

Образец wiring: `services/logg/service.go` + `pkg/svcutil.BootstrapWithConfig`.

---

## Service map

| Service   | `internal/` (typical)                                                                             | Status  | Entrypoint    |
| --------- | ------------------------------------------------------------------------------------------------- | ------- | ------------- |
| `nats`    | `config`                                                                                          | live    | `cmd/nats`    |
| `gate`    | `config`, `core`, `auth`, `middleware`, `policy`, `proxy`, `routing`, `websocket`, `ops`, `admin` | live    | `cmd/gate`    |
| `space`   | `config`, `store`, `apis`                                                                         | now     | `cmd/space`   |
| `work`    | `config`, `store`, `apis`                                                                         | planned | `cmd/work`    |
| `agents`  | `config`, `store`, `apis`                                                                         | planned | `cmd/agents`  |
| `comms`   | `config`, `store`, `apis`                                                                         | planned | `cmd/comms`   |
| `integ`   | `config`, `store`, `apis`                                                                         | planned | `cmd/integ`   |
| `media`   | `config`, `store`, `apis`                                                                         | planned | `cmd/media`   |
| `runtime` | `config`, `store`, `apis`                                                                         | planned | `cmd/runtime` |
| `logg`    | `config`, `store`, `apis`                                                                         | live    | `cmd/logg`    |
| `ctrl`    | `apis`, `config`, `models`, `supervise`, `edgecfg`, `admin`                                       | live    | `cmd/ctrl`    |

Описания: [services/README.md](./services/README.md). Роли и order: [architecture.md](./architecture.md).

Имена-черновики `auth`, `task`, `agent`, `files`, `exec`, `cron`, `memo`, `pipe`, `conn` — не сервисы.

---

## Layout examples

### Domain (`space`, later `work` / `agents` / …)
cmd/space/main.go
cfgs/space.json
services/space/
service.go
models/
client/
internal/
config/
store/
apis/
text### Edge (`gate`)
cmd/gate/main.go
cfgs/gate.json
services/gate/
service.go
export.go
client/
internal/
core/
auth/, proxy/, routing/, websocket/, …
text`cmd/gate/` живёт в корне репо, не внутри `services/gate/`.

### Control / bus (`ctrl`, `nats`)
cmd/ctrl/main.go
services/ctrl/
service.go          # Run(); -up supervise
export.go
internal/
apis/, supervise/, edgecfg/, admin/, …
cmd/nats/main.go
services/nats/
service.go
export.go
internal/config/
text---

## Cross-service boundaries

| Who                     | How it calls another service              |
| ----------------------- | ----------------------------------------- |
| Browser / external HTTP | `gate` → `public.*`                       |
| Service → service       | `services/<name>/client` + `pkg/subjects` |
| Gate HTTP client (ops)  | `services/gate/client`                    |

**Do not** import `services/<other>/internal/...`.

---

## New service checklist

1. `services/<name>/` — `service.go`, `internal/{config,store,apis}`, `models/` если есть DTO.
2. `cmd/<name>/main.go` → `<name>.Run()`.
3. `cfgs/<name>.json` (`name`, `order`, `env`).
4. RPC — `client/` + константы в `pkg/subjects`.
5. Публичные маршруты — `mini.WithPublicHTTP` + `mini.WithPublicSubject`, не свой listen-порт.
6. Store только через `pkg/sqldb`. Не экспортировать store/apis.
7. Страница `docs/services/<name>.md` + строка в [services/README.md](./services/README.md).
8. Пока сервис не в [NOW.md](./NOW.md) — не писать код.

---

## Processes and transport

| Service   | Role                    | Transport               |
| --------- | ----------------------- | ----------------------- |
| `nats`    | message bus             | NATS server             |
| `gate`    | HTTP/WS edge            | HTTP/WS → mini/NATS     |
| `space`   | identity / tenancy      | NATS mini               |
| `work`    | issues                  | NATS mini               |
| `agents`  | run / session / memory  | NATS mini               |
| `comms`   | channels / messages     | NATS mini + WS via gate |
| `integ`   | connectors              | NATS mini               |
| `media`   | blobs                   | NATS mini               |
| `runtime` | machines; Docker socket | NATS mini               |
| `logg`    | logs + activity         | NATS mini               |
| `ctrl`    | lifecycle + supervise   | NATS mini               |

Configs: `cfgs/<name>.json`.