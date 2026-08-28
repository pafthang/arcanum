# `nats`

Embedded NATS Server (+ JetStream) — platform message bus.

| | |
|--|--|
| Package | `services/nats` |
| Entry | `go run ./cmd/nats` / `task run:nats` |
| Config | `cfgs/nats.json` |
| Data | `data/nats/` (JetStream) |
| Order | 0 |

## Responsibilities

- Run NATS for all mini services
- JetStream store (when enabled by config)

## Dependencies

| Direction | Target | Why |
|-----------|--------|-----|
| used by | all services | `NATS_URL` |

## Public surface

| Kind | Notes |
|------|--------|
| HTTP | none (not a domain API) |
| Client ports | NATS URL from env (`nats://127.0.0.1:4222` dev) |

## Ops

- Starts first (`order: 0`); `task up` waits for the bus.
- Logs: `data/nats/service.log`

## See also

- Code: `services/nats/`
- [Architecture](../architecture.md)
