# `space`

Identity / tenancy: users, spaces, nested teams, membership, agent principal.

|         |                                         |
| ------- | --------------------------------------- |
| Package | `services/space`                        |
| Entry   | `go run ./cmd/space` · `task run:space` |
| Config  | `cfgs/space.json`                       |
| Data    | `data/space/space.db`                   |
| Order   | 50                                      |
| Status  | live / kernel                           |
| Class   | domain                                  |

## Public surface (код)

| Kind   | Path / subject |
| ------ | -------------- |
| HTTP   | `POST /api/auth/login` (auth none) |
| HTTP   | `GET /api/spaces` `POST /api/spaces` |
| HTTP   | `GET /api/spaces/{spaceId}` |
| HTTP   | `GET /api/spaces/{spaceId}/members` |
| RPC    | `internal.space.get` / `list_for_user` / `user.get` |
| Client | `services/space/client` |

Store уже имеет `teams` и `api_keys`. HTTP для них нет. Seed: `admin@kuayle.local` + space `default`.

Подробно: [TODO.md](../TODO.md), [GAPS.md](../GAPS.md).
