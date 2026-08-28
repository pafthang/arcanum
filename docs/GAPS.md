# GAPS

Разница между **нашей реализацией** (`pafthang/arcanum`, дерево на 2026-08-28) и двумя ориентирами:

- [Kuayle](https://github.com/carbogninalberto/kuayle) — Linear-like PM (`BE/`+`UI/`) и отдельно Dev Machines (`devmachine/`, TECHNICAL.md)
- [GoClaw](https://github.com/nextlevelbuilder/goclaw) — multi-tenant agent platform (Go monolith + Postgres/pgvector + embedded UI)

Это не action plan. Приоритеты — отдельным проходом после чтения.

Решения, которые gaps не оспаривают: [DECISIONS.md](./DECISIONS.md).
Что уже лежит в дереве: [TODO.md](./TODO.md).

## Как читать статусы у нас

| Метка | Значение |
| ----- | -------- |
| есть | маршрут/таблица/процесс в репо |
| kernel | есть каркас, нет продуктовой полноты |
| нет | нет cmd/cfg или нет предметной модели |
| иначе | сознательно другое (не «недоделали») |

Arcanum — шасси из процессов на NATS. Kuayle и GoClaw — продукты-монолиты. Копировать их целиком не цель.

## Соответствие понятий

| Arcanum | Kuayle | GoClaw |
| ------- | ------ | ------ |
| Space | workspace | tenant / workspace |
| Team (вложенная группа) | team + workflow/statuses | agent team (другая сущность) |
| User `actor=user\|agent` | human member | user + agent identity |
| `work` issue | issue | task board item (частично) |
| `comms` channel | нет как SoT (есть comments/inbox) | messaging channel + web chat |
| `agents` run/session/memory | Dev Machine agent run (CLI в контейнере) | in-process pipeline + memory + skills |
| `integ` connector/webhook | GitHub App + workspace webhooks | 7 messenger channels + webhooks |
| `media` | S3/FS uploads | media tools + vault files |
| `runtime` | Dev Machines (Docker) | exec/browser tools, Lite desktop |
| `gate` | Echo HTTP + nhooyr WS | embedded HTTP + WS dashboard |
| JWT `space_id` + `space_role` | workspace RBAC | tenant RBAC |

UI: один клиент на gate, линия Kuayle/Svelte. Отдельный GoClaw dashboard не поднимаем.

---

## A. Kuayle → Arcanum

Источник: [carbogninalberto/kuayle](https://github.com/carbogninalberto/kuayle) (`main`, Apache-2.0). Это **два слоя в одном монолите**, не один продукт:

| Слой | Где в репо | Что это | Куда у нас |
| ---- | ---------- | ------- | ---------- |
| PM / Linear-like | `BE/` + `UI/` | workspace, teams, issues, cycles, projects, views, inbox, GitHub App, webhooks, Svelte | `space` + `work` + `integ` + UI на gate |
| Dev Machines | `devmachine/` + [TECHNICAL.md](https://github.com/carbogninalberto/kuayle/blob/main/TECHNICAL.md) | Docker-машины, Machine Gateway, collector, agent CLI в контейнере | `runtime` (+ кусок `agents` как run, привязанный к machine) |

Агенты Kuayle — не GoClaw-pipeline. Это CLI в контейнере (Claude Code / OpenCode / Codex / generic) на машине issue. Модели: `dev_machine_agent_runs`, `…_steps`, `…_providers`. Docker socket только у Machine Manager — это уже наше решение про `runtime`.

NATS в Kuayle нет. Realtime — Echo + nhooyr WebSocket.

| Область | Kuayle | У нас | Gap |
| ------- | ------ | ----- | --- |
| Workspace RBAC | owner/admin/member/guest | те же роли в `space_members` | совпадает по модели |
| Login | JWT | `POST /api/auth/login` HS256 | нет register, refresh, switch space, invite |
| Teams | workflow, свои статусы, triage | таблица `teams` + store Create/AddMember | нет HTTP, нет workflow/statuses |
| Issues | priority, due, sub-issues, multi-assignee, history | title/body/status/assignee/labels/comments | нет priority/due/parent/relations/history/multi-assignee |
| Labels | иерархия, soft delete, defaults | плоские space labels | нет parent, delete, defaults |
| Comments | issue + project | issue comments в `work` | нет project comments |
| Relations | blocks / duplicate / related | нет | нет |
| Cycles | sprints, burndown | нет | нет сервиса/таблиц |
| Projects | cross-team + Gantt | нет | нет |
| Views | saved views, scope, reorder | нет | нет |
| Inbox | snooze/read/archive | нет | нет |
| Realtime | WS issues/comments/presence | только comms channel WS catalog | нет work presence |
| GitHub | App, repo link, ID match, status rules | integ connector + repo row + hook ingest | нет App install, нет status rules, нет ID-extract как продукта |
| Webhooks | outbound workspace | integ webhooks + deliveries | kernel есть |
| Analytics | overview, burn-up | `work/overview` счётчики | нет графиков/insights |
| Public share | token links | нет | нет |
| Assets | signed URL, S3 | `blob_id` поля, сервиса `media` нет | нет загрузки |
| Transfer | export/import | нет | нет |
| Editor / palette | Tiptap, command palette | нет UI | нет клиента |
| Dev Machines | Docker multi-container | `runtime` planned | нет |

Итог по Kuayle:

- PM: tenancy-скелет и issue-ядро есть. Нет workflow/teams HTTP, PM-слоя (cycles/projects/views/inbox), UI, файлов.
- Dev Machines: в Arcanum это не `work` и не GoClaw-tools. Схема машин (`dev_machines`, services, tickets, gateway) — только будущий `runtime`. Не тащить 20 таблиц `000033_dev_machines` в `work`/`agents`.

---

## B. GoClaw → Arcanum

Источник: [nextlevelbuilder/goclaw](https://github.com/nextlevelbuilder/goclaw) (ветка `dev`, docs.goclaw.sh).

| Область | GoClaw | У нас | Gap |
| ------- | ------ | ----- | --- |
| Форма поставки | один бинарь + Postgres (Lite: SQLite) | много процессов + SQLite на сервис | иначе, не gap |
| Tenancy | workspace + RBAC + isolated sessions | Space + membership | нет session isolation per agent workspace на диске |
| Agent identity | agent + keys AES-256-GCM | `actor=agent`, `api_keys.key_hash` в space | нет public API ключей, нет AES wrap секретов LLM |
| Pipeline | 8 стадий context→…→summarize | stub: queued→running→finish | нет LLM, tools, observe, summarize |
| Prompt modes | Full/Task/Minimal/None | нет | нет |
| Memory | working / episodic / semantic + vault + pgvector | KV memory + tier поле | нет FTS/vector, нет wikilinks vault |
| Skills | BM25+semantic, gating, use_skill | list/create body в SQLite | нет discovery/search/runtime use |
| Tools | 30+ (fs, exec, browser, web, media, teams, cron) | нет tool host | весь слой tools |
| MCP | custom tools | нет | нет |
| Providers | 20+ LLM | нет | нет |
| Agent teams | boards, delegate, 3 режима | Space.Team ≠ agent team | нет оркестрации агентов |
| Self-evolution | metrics→adapt | нет | нет |
| Channels | Telegram, Discord, Slack, Zalo, Zalo OA, Feishu, WhatsApp | comms внутренний чат + integ kind enum | нет живых адаптеров |
| Web chat UI | React dashboard / Wails Lite | нет | UI запрещён как второй dashboard; нужен общий клиент |
| Cron / heartbeat | automation tools | нет | later workers / runtime |
| Security extras | prompt injection, SSRF, 5-layer perms | gate JWT + space membership | нет tool sandbox policy |
| Observability | LLM traces, OTEL | logg + gate metrics stub | нет span на LLM |
| Knowledge vault | docs + hybrid search | нет | ближе к `media` + `agents` memory |

Итог по GoClaw: имена сущностей (run, session, memory, skill, connector kinds) уже в моделях. Исполнения агента нет. Каналы снаружи — только hook-заготовки.

---

## C. Что у нас есть, чего нет у ориентиров (не копировать обратно)

- NATS mini + `GET /_catalog` как контракт
- Один gate, доменные сервисы без listen-порта
- Владение store на сервис (не shared Postgres)
- Разделение `work` comments vs `comms` chat vs `logg` activity
- `ctrl -up` из `cfgs/*.json`

Это намеренные отличия шасси. В gaps их не «дотягиваем» до монолита.

---

## D. Корзины для приоритизации

Не порядок работ. Группы, чтобы выбрать один пункт в NOW.

### P-product (без этого UI бесполезен)

1. `space`: invite member, switch space, teams HTTP, API key для `actor=agent`
2. `work`: assignee+event реально стартует `commands.agents.run.start`
3. UI-клиент на gate (Kuayle-линия): login, spaces, issues, один канал

### P-agent (GoClaw-глубина)

1. Один OpenAI-compatible provider + tool loop (read/write через будущий runtime, не Docker socket в API-сервисах)
2. Memory search (хотя бы FTS на SQLite)
3. Skill inject в prompt run

### P-channel (GoClaw каналы / Kuayle GitHub)

1. Один живой messenger adapter в `integ` (кандидат: Telegram)
2. GitHub: разбор `ABC-123` / `#12` → событие в `work`
3. Integ inbound → `comms` уже съедает `events.integ.message.inbound` — закрыть путь адаптер→event

### P-files-runtime (Kuayle assets / Dev Machines)

1. `media`: upload через gate, blob_id уже есть в comms/work моделях
2. `runtime`: Docker socket только здесь

### P-pm (Kuayle полный Linear)

Cycles, projects, views, inbox, relations, public share — после того как issue kernel + UI живые.

---

## E. Явно не gaps

- Перенос на Postgres как общий default
- Отдельный GoClaw UI
- Shared DB
- Второй публичный HTTP
- Копирование 20 провайдеров и 30 tools оптом
