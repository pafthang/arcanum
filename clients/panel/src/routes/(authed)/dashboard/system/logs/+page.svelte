<script lang="ts">
	import { page } from '$app/state';
	import SubscriptionLog from '$lib/features/system/SubscriptionLog.svelte';
	import TorrentLog from '$lib/features/system/TorrentLog.svelte';
	import SessionsAccess from '$lib/features/system/SessionsAccess.svelte';

	const tab = $derived(page.url.searchParams.get('tab') ?? 'srh');
</script>

{#if tab === 'torrent'}
	<TorrentLog />
{:else if tab === 'sessions'}
	<SessionsAccess />
{:else}
	<SubscriptionLog />
{/if}
