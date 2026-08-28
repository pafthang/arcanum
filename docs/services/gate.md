# `gate`

HTTP/WS edge: auth middleware, catalog-based routing, proxy to NATS public routes, WS fan-out.

| | |
|--|--|
| Package | `services/gate` |
| Entry | `go run ./cmd/gate` / `task run:gate` |
| Config | `cfgs/gate.json`, JWT keys from `cfgs/comm.json` |
| Data | `data/gate/` (logs) |
| Order | 10 |
| Default URL | `http://127.0.0.1:8080` |

## Responsibilities

- Sole public HTTP/WS for browsers/clients
- `GET /_catalog` — live contract
- JWT/HMAC (and other auth modes in `internal/auth`)
- CORS, rate limit, policy hooks

## Dependencies

| Direction | Target | Why |
|-----------|--------|-----|
| requires | `nats` | bus + public subjects |
| proxies to | all mini services with `mini.Public` | |

## Public surface

| Kind | Notes |
|------|--------|
| HTTP | edge for all public routes + ops |
| WS | catalog `kind=ws` routes: upgrade → NATS Subscribe (out) / Publish (in); auth via `Authorization` or `?access_token=` |
| Catalog | `GET /_catalog` (HTTP + WS routes) |
| Client | `services/gate/client` (HTTP), `@arcanum/svelte` |

Full domain route list lives in the catalog and owners’ package comments — not here.

## Config (operators)

| Env | Meaning |
|-----|---------|
| `GATE_JWT_HMAC_SECRET` | HS256 secret (must match auth) |
| `GATE_JWT_ISSUER` | JWT `iss` |
| `WS_ALLOWED_ORIGINS` | WS origin ACL |
| `CLAIM_HEADERS` | JWT claims → upstream headers |
| `RATE_LIMIT_*` | See [Gate Rate Limiting](../runbooks/gate-rate-limiting.md) |
| `METRICS_ENABLED` / `METRICS_PATH` | See [Observability](../runbooks/observability.md) |

## Ops

- Health: `task health` / curl catalog
- Logs: `data/gate/service.log`

## See also

- [API policy](../api/README.md)
- [ADR 0001](../adr/0001-nats-mini-gate-edge.md)
