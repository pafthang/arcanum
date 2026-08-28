# `space`

Identity / tenancy: users, spaces, nested teams, membership, agent principal stub.

|         |                                         |
| ------- | --------------------------------------- |
| Package | `services/space`                        |
| Entry   | `go run ./cmd/space` · `task run:space` |
| Config  | `cfgs/space.json`                       |
| Data    | `data/space/space.db`                   |
| Order   | 50                                      |
| Class   | domain                                  |

## Responsibilities

- Users (actor `user|agent`, optional `platform_admin`)
- Spaces and `space_members` with `owner|admin|member|guest`
- Nested teams inside a space (`parent_id`)
- API key hash for agent principals
- Login → JWT claims `sub`, `space_id`, `space_role`
- Public routes via gate catalog (`mini.public`)

## Dependencies

| Direction | Target             | Why                    |
| --------- | ------------------ | ---------------------- |
| requires  | `nats`             | mini + public subjects |
| store     | SQLite `pkg/sqldb` | own schema             |

## Public surface

| Kind   | Path / subject                       |
| ------ | ------------------------------------ |
| HTTP   | `POST /api/auth/login` (auth none)   |
| HTTP   | `GET /api/spaces` `POST /api/spaces` |
| HTTP   | `GET /api/spaces/{spaceId}`          |
| HTTP   | `GET /api/spaces/{spaceId}/members`  |
| Client | `services/space/client`              |

## Layout

```
cmd/space/main.go
cfgs/space.json
services/space/service.go
services/space/models/
services/space/client/
services/space/internal/config/
services/space/internal/store/
services/space/internal/apis/
```

Не держать entry в `services/space/cmd/` и cfg в `services/space/cfgs/`.

## See also

- [NOW.md](../NOW.md)
- [TODO.md](../TODO.md)
- [services-structure.md](../services-structure.md)