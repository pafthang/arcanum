# `@arcanum/ts-client`

Svelte 5 (runes) client for the Go Remnawave API. This is not the panel UI.

```svelte
<script>
  import { createRemnawave, session } from '@arcanum/ts-client';

  session.init();
  const api = createRemnawave({ baseUrl: '' }); // same-origin /api via proxy
  const status = api.auth.status();
  const login = api.auth.login();
</script>
```

- Unwraps `{ response }` and maps `{ timestamp, path, message, errorCode }`
- Sends `X-Remnawave-Client-Type: browser` and `Authorization: Bearer …`
- `createQuery` / `createMutation` are runes (call from a component)
- Covers every `/api` route from Go `router.go` except asynqmon HTML (`/api/backend-tools/queues`)
- React still has user/node mutations (create/bulk/restart) that Go has not registered yet — those are not in this client until they land in `back`
