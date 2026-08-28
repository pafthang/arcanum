# `space`

Identity / tenancy: users, spaces, nested teams, membership, agent principal.

|         |                                         |
| ------- | --------------------------------------- |
| Package | `services/space`                        |
| Entry   | `go run ./cmd/space` · `task run:space` |
| Config  | `cfgs/space.json`                       |
| Data    | `data/space/space.db`                   |
| Order   | 50                                      |
| Status  | live                                    |
| Class   | domain                                  |

## Public surface (код)

| Kind | Path / subject |
| ---- | -------------- |
| HTTP | `POST /api/auth/login` (auth none) |
| HTTP | `POST /api/auth/register` (auth none) |
| HTTP | `POST /api/auth/switch-space` |
| HTTP | `GET /api/spaces` `POST /api/spaces` |
| HTTP | `GET /api/spaces/{spaceId}` |
| HTTP | `GET /api/spaces/{spaceId}/members` |
| HTTP | `POST /api/spaces/{spaceId}/members` (invite by email) |
| HTTP | `PATCH /api/spaces/{spaceId}/members/{userId}` |
| HTTP | `GET/POST /api/spaces/{spaceId}/teams` |
| HTTP | `GET /api/spaces/{spaceId}/teams/{teamId}` |
| HTTP | `POST /api/spaces/{spaceId}/teams/{teamId}/members` |
| HTTP | `GET/POST /api/spaces/{spaceId}/keys` (secret once) |
| RPC | `internal.space.get` / `list_for_user` / `user.get` / `can` |
| Client | `services/space/client` |

Invite принимает уже зарегистрированный email. Register создаёт user + personal space (owner). Switch перевыпускает JWT с новым `space_id`/`space_role`. Key создаёт `actor=agent` + membership + hash.

Seed: `admin@kuayle.local` + space `default`.

Подробно: [TODO.md](../TODO.md), [GAPS.md](../GAPS.md).
