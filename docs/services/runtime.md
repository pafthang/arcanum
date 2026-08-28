# `runtime`

Dev Machines и единственное место с Docker socket.

|         |                     |
| ------- | ------------------- |
| Package | `services/runtime`  |
| Entry   | `cmd/runtime`       |
| Config  | `cfgs/runtime.json` |
| Order   | 100                 |
| Status  | live / kernel       |
| Class   | domain              |

## Routes

- `GET/POST /api/spaces/{spaceId}/machines`
- `GET /api/spaces/{spaceId}/machines/{machineId}`
- `POST /api/spaces/{spaceId}/machines/{machineId}/stop`

`RUNTIME_DOCKER_HOST` пустой → статус `recorded` (учёт, без контейнера). Сокет только здесь, не в `agents`.

## Не владеет

- agent run/session — `agents`
- blob store — `media`
- живой Docker start/exec и machine-gateway — следующий срез

## See also

- [GAPS.md](../GAPS.md) Kuayle Dev Machines / GoClaw exec+browser
- [agents.md](./agents.md)
- [DECISIONS.md](../DECISIONS.md)
