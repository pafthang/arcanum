<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import { session } from '$lib/rw';

	let { children } = $props();
	let ready = $state(session.isAuthenticated);

	onMount(() => {
		session.init();
		if (!session.isAuthenticated) {
			goto('/auth/login');
			return;
		}
		ready = true;
	});
</script>

{#if ready}
	<AppShell>{@render children()}</AppShell>
{:else}
	<div class="flex h-screen items-center justify-center text-sm text-[var(--color-text-secondary)]">
		Loading…
	</div>
{/if}
