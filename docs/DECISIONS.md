# DECISIONS

Короткие факты. Новая строка = новое решение. Не переписывать историю.

- 2026-08-28: Шасси платформы — этот репозиторий (Arcanum). Kuayle и GoClaw не становятся корнем.
- 2026-08-28: Tenant публично называется Space. `space_id` в JWT. Отдельного публичного `workspace_id` нет.
- 2026-08-28: Space = Kuayle workspace = GoClaw tenant.
- 2026-08-28: Team — вложенная группа внутри Space, не tenant.
- 2026-08-28: Публичный HTTP/WS только `gate`. Исключение: wildcard `machine-gateway` для Dev Machines.
- 2026-08-28: Сервис владеет своим store. Shared DB запрещена.
- 2026-08-28: Cross-service только typed client + `pkg/subjects`, не raw subjects из чужого кода.
- 2026-08-28: `work` владеет issue-агрегатом. `agents` владеет run/session/memory. Связь — events (`issue.assigned` → `commands.agents.run.start`).
- 2026-08-28: Агент — отдельный principal (`actor=agent`), не человеческая сессия и не shared user password.
- 2026-08-28: Docker socket только у `runtime` manager. API-сервисы сокет не монтируют.
- 2026-08-28: Default store — SQLite на сервис. Postgres можно включить по сервису позже, не ломая владение схемой.
- 2026-08-28: UI — один клиент на gate (линия Kuayle/Svelte). Отдельный GoClaw dashboard не поднимаем.
- 2026-08-28: Текущая работа фиксируется только в `docs/NOW.md`.
- 2026-08-28: Карта процессов — `nats`, `gate`, `space`, `work`, `agents`, `comms`, `integ`, `media`, `runtime`, `logg`, `ctrl`. Черновики `auth` / `task` / `agent` / `files` / `exec` / `cron` / `memo` / `pipe` / `conn` сервисами не являются.
- 2026-08-28: Комментарии issue — агрегат `work`. Продуктовый чат — `comms`. Шины и audit — `nats` / `logg`.
- 2026-08-28: Точка входа сервиса — корневой `cmd/<name>/main.go`. Cfg — корневой `cfgs/<name>.json`.
- 2026-08-28: SQLite только `pkg/sqldb` (`modernc.org/sqlite`). CGO/`mattn` запрещены.