# TODO

Инвентарь по дереву `pafthang/arcanum` на 2026-08-28.
Активный кусок кода: [NOW.md](./NOW.md). Пробелы к продукту: [GAPS.md](./GAPS.md).

## Сделано (есть процесс + cfg + поверхность)

### Шасси

- [x] `nats` — embedded NATS + JetStream, order 0
- [x] `gate` — HTTP/WS edge, catalog, JWT/HMAC, CORS, rate limit
- [x] `logg` — SQLite logs, `GET /api/logs`, space activity
- [x] `ctrl` — inventory/lifecycle + `ctrl -up` supervisor
- [x] `pkg/mini` catalog (`mini.Public`, WS, `$SRV.INFO`)
- [x] `pkg/sqldb` (`modernc.org/sqlite`, без CGO)
- [x] `pkg/subjects` / `pkg/events` / `pkg/svcutil` / `pkg/passwd`

### `space` kernel

- [x] schema: users, spaces, space_members, teams, team_members, api_keys
- [x] seed: `admin@kuayle.local` + space `default` + owner
- [x] `POST /api/auth/login` → JWT `sub`, `space_id`, `space_role`
- [x] `GET/POST /api/spaces`, `GET /api/spaces/{spaceId}`
- [x] `GET /api/spaces/{spaceId}/members`
- [x] internal: get space, list for user, get user
- [x] typed `services/space/client`
- [ ] public invite / update member
- [ ] public teams CRUD
- [ ] public API keys / agent principal create
- [ ] register / switch space
- [ ] `internal.space.can` (константа есть, handler нет)

### `work` kernel

- [x] issues: list/create/get/update
- [x] comments list/create
- [x] labels list/create + issue label ids
- [x] `GET /api/spaces/{spaceId}/work/overview`
- [x] events `issue.created` / `updated` / `assigned`
- [x] internal get/list/overview
- [ ] priority, due, sub-issues, relations, multi-assignee
- [ ] cycles / projects / views / inbox

### `agents` kernel

- [x] runs list/create/get/cancel + session get
- [x] memory list/put (tiers working/episodic/semantic в модели)
- [x] skills list/create (store body, без search/gating)
- [x] commands `run.start` / `run.cancel`
- [x] stub pipeline (queued → running → finish, без LLM/tools)
- [ ] providers, tool loop, MCP
- [ ] consume `events.work.issue.assigned` → start run
- [ ] agent teams / delegation

### `comms` kernel (C1 закрыт)

- [x] `cmd/comms` + `cfgs/comms.json`
- [x] channels list/create/get
- [x] messages list/create, parent_id threads, blob_id
- [x] WS catalog `/api/spaces/{spaceId}/channels/{channelId}/ws`
- [x] internal get/list/create + inbound ingest
- [ ] membership check через space на каждый mutate (как work) — сверить при ревью
- [ ] presence / typing / reactions

### `integ` kernel

- [x] connectors CRUD + github repos attach
- [x] outbound webhooks + deliveries
- [x] inbound `POST /api/integ/hooks/{connectorId}`
- [x] inbound GitHub `POST /api/integ/hooks/github/{connectorId}`
- [x] fan-out work events → webhook deliveries
- [ ] живые адаптеры telegram/discord/slack/zalo/feishu/whatsapp
- [ ] GitHub App installation flow (не только запись repo)

### Нет в дереве

- [ ] `media` (нет cmd/cfg/service)
- [ ] `runtime` (нет cmd/cfg/service)
- [ ] UI на gate (Svelte / Kuayle-линия)
- [ ] `docs/adr/*`, `docs/api/*`, `docs/runbooks/*` (ссылки в старых доках были битые)

## Не сейчас

Не начинать вне NOW:

- integ messenger adapters
- media blobs
- agents LLM runtime
- labels hierarchy / cycles / projects
- Dev Machines
