# Arcanum

Шасси платформы: независимые Go-процессы на NATS mini + один публичный HTTP/WS край (`gate`).

Продуктовые ориентиры — не корни репозитория:

- [Kuayle](https://github.com/carbogninalberto/kuayle) — Linear-like workspace (issues, teams, cycles)
- [GoClaw](https://github.com/nextlevelbuilder/goclaw) — multi-tenant agent platform (runs, memory, channels, tools)

Сравнение и пробелы: [docs/GAPS.md](docs/GAPS.md).
Что уже в дереве: [docs/TODO.md](docs/TODO.md).
Текущий кусок кода: [docs/NOW.md](docs/NOW.md).

## Карта процессов

| Order | Service | Роль | Статус |
|------:|---------|------|--------|
| 0 | `nats` | Embedded NATS + JetStream | live |
| 10 | `gate` | HTTP/WS edge, catalog, JWT | live |
| 50 | `space` | users, spaces, teams, principals | live (kernel) |
| 60 | `work` | issues, labels, comments | live (kernel) |
| 70 | `agents` | run / session / memory / skills | live (kernel) |
| 80 | `comms` | channels, messages, WS catalog | live (kernel) |
| 85 | `integ` | connectors, hooks, outbound webhooks | live (kernel) |
| 90 | `media` | blobs / files | planned |
| 100 | `runtime` | Dev Machines, Docker socket | planned |
| 110 | `logg` | logs + activity | live |
| 120 | `ctrl` | lifecycle + supervisor `-up` | live |

`kernel` = процесс, cfg, store, catalog-маршруты есть; продуктовая глубина меньше Kuayle/GoClaw.

Живой HTTP-контракт: `GET /_catalog` на gate, не эти таблицы.

## Запуск

```bash
go install github.com/go-task/task/v3/cmd/task@latest
task up          # ctrl -up: nats → gate → domain
task catalog     # GET http://127.0.0.1:8080/_catalog
task down
```

Один сервис: `task run:space` / `go run ./cmd/space`.
Сборка и тесты: `task build` · `task test`.

Seed: при старте `space` создаёт space `default` и пользователя `admin@kuayle.local` (пароль `admin` или `SPACE_SEED_PASSWORD`), роль `owner`, `platform_admin`.

## Layout

```
cmd/<name>/main.go      # только Name.Run()
cfgs/<name>.json        # order + env для ctrl -up
services/<name>/        # service.go, models/, client/, internal/
pkg/                    # mini, subjects, sqldb, svcutil, …
docs/                   # карта, решения, gaps
```

Правила: [docs/services-structure.md](docs/services-structure.md), решения: [docs/DECISIONS.md](docs/DECISIONS.md).

## Не это

- Второй публичный HTTP рядом с `gate` (исключение позже: `machine-gateway` у runtime)
- Shared DB между сервисами
- Отдельный admin-микросервис и отдельный GoClaw dashboard
- Имена-черновики `auth` / `task` / `agent` / `files` / `exec` как сервисы
