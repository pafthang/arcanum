# `logg`

Centralized logging aggregator and multi-sink logg component.

| | |
|--|--|
| Package | `services/logg` |
| Entry | `go run ./cmd/logg` / `task run:logg` |
| Config | `cfgs/logg.json` |
| Data | `data/logg/logg.db` |
| Order | 99 |
| Class | infrastructure |

## Responsibilities

- Subscribes to `logs.>` NATS subject.
- Centralizes application logs from all mini-services.
- Stores logs in SQLite (`logg.db`).
- Exposes logs via API for debugging and monitoring.

## Public surface

| Kind | Notes |
|------|--------|
| HTTP | `GET /api/logs` |
| Internal | `logg` (logs.list) |

## Component (`pkg/logx`)

The `pkg/logx` package provides a custom `slog.Handler` that multiplexes log output:
- **term**: `slog.NewTextHandler` (stdout)
- **file**: `slog.NewJSONHandler` (`data/$SVC/service.log`)
- **nats**: Publishes JSON records to NATS (`logs.$SVC`)

This is automatically injected into all mini-services via `svcutil.Bootstrap`.
