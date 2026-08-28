<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { rw } from '$lib/rw';

	const map: Record<string, (id: string) => string> = {
		user: (id) => `/dashboard/management/users?id=${id}`,
		node: (id) => `/dashboard/management/nodes?id=${id}`,
		'config-profile': (id) => `/dashboard/management/config-profiles?id=${id}`,
		'internal-squad': () => `/dashboard/management/internal-squads`,
		'external-squad': () => `/dashboard/management/external-squads`,
		'node-plugin': (id) => `/dashboard/management/plugins?id=${id}`,
		'subpage-config': (id) => `/dashboard/subpage?id=${id}`
	};

	const resolve = rw.users.resolve();

	onMount(async () => {
		const entity = page.params.entity ?? '';
		const id = page.params.id ?? '';
		if (entity === 'user') {
			try {
				const n = Number(id);
				await resolve.mutate(Number.isFinite(n) ? { id: n } : { shortUuid: id });
			} catch {
				/* still navigate to users */
			}
		}
		const to = map[entity]?.(id) ?? '/dashboard/home';
		goto(to);
	});
</script>

<div class="text-sm text-muted-foreground">Opening…</div>
