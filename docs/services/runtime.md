# `runtime`

Dev Machines и единственное место с Docker socket. Сервиса в дереве нет.

|         |                         |
| ------- | ----------------------- |
| Package | `services/runtime`      |
| Entry   | `cmd/runtime` — нет     |
| Config  | `cfgs/runtime.json` — нет |
| Order   | 100                     |
| Status  | planned                 |
| Class   | domain                  |

## Responsibilities (когда начнётся)

- Поднять/остановить machine для агента или человека
- Docker socket только здесь. API-сервисы сокет не монтируют
- Исключение из «один gate»: wildcard `machine-gateway` для сессии машины (code-server / tty) — как у Kuayle Dev Machines

## Не владеет

- agent run/session — `agents` просит runtime через client
- blob store — `media`

## Не начинать

Пока нет строки в [NOW.md](../NOW.md).

## See also

- [GAPS.md](../GAPS.md) Kuayle Dev Machines / GoClaw exec+browser
- [agents.md](./agents.md)
- [DECISIONS.md](../DECISIONS.md)
