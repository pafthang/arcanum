<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import BootScreen from '$lib/components/auth/BootScreen.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';

	const callback = rw.auth.oauth2Callback();

	onMount(async () => {
		const provider = page.params.provider ?? '';
		const code = page.url.searchParams.get('code') ?? undefined;
		const state = page.url.searchParams.get('state') ?? undefined;
		try {
			await callback.mutate({ provider, code, state });
			goto('/dashboard/home');
		} catch (err) {
			appToast.apiError(err, 'OAuth2 callback failed');
			goto('/auth/login');
		}
	});
</script>

<BootScreen message="Signing in…" />
