# `ctrl`

Platform control plane: lifecycle API + optional process supervisor (`-up`).

| | |
|--|--|
| Package | `services/ctrl` |
| Entry | `go run ./cmd/ctrl` · `go run ./cmd/ctrl -up` |
| Config | `cfgs/ctrl.json` |
| Data | `data/` children when supervising |
| Order | 120 |

## Responsibilities

- Lifecycle/reload signals for services
- Inventory/admin APIs (platform)
- `-up`: start enabled services from `cfgs/` (skip self), wait NATS, restart on crash

## Dependencies

| Direction | Target | Why |
|-----------|--------|-----|
| requires | `nats` (for mini API); all children when `-up` | supervise |
| reads | `cfgs/*.json` | order/env |

## Public surface

| Kind | Notes |
|------|--------|
| HTTP | platform/lifecycle (via gate) |
| CLI | `-up` supervisor mode |

## Ops

```bash
task up
# logs per child: data/<name>/service.log
task down
```

## See also

- [runbooks/local-dev.md](../runbooks/local-dev.md)
