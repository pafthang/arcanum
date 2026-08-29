# Remnawave web panel

Svelte 5 + SvelteKit SPA for the Remnawave backend. Talks to the API through [`@arcanum/ts-client`](../svelte).

## Develop

API is expected on `http://127.0.0.1:3000`. From the repo root, `task up` starts redis, backend, node, and this panel (`http://127.0.0.1:5173`). Vite proxies `/api`.

```sh
# full stack
task up

# panel only (API already running)
task web

# or from this workspace
cd clients
bun install
bun run dev
```

Open `/auth/login`. After login the panel lives under `/dashboard`.

## Routes

Mirrors the React panel in `clients/frontend`:

- `/auth/login`, `/oauth2/callback/:provider`
- `/dashboard/home`
- `/dashboard/management/{users,hosts,nodes,internal-squads,external-squads,config-profiles,plugins,settings,subscription-settings,response-rules}`
- `/dashboard/templates/:type/:uuid`
- `/dashboard/subpage/:uuid`
- `/dashboard/crm/infra-billing`
- `/dashboard/tools/{hwid-inspector,srh-inspector,torrent-blocker-reports,sessions-explorer,http-stats,quick-open,snippets}`
