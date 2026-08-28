<script lang="ts">
	import { page } from '$app/state';
	import PanelMetrics from '$lib/features/system/PanelMetrics.svelte';
	import NodeLiveMetrics from '$lib/features/system/NodeLiveMetrics.svelte';
	import NodeTrafficStats from '$lib/features/system/NodeTrafficStats.svelte';
	import HttpMetrics from '$lib/features/system/HttpMetrics.svelte';

	const tab = $derived(page.url.searchParams.get('tab') ?? 'panel');
</script>

{#if tab === 'nodes'}
	<div class="flex min-h-0 flex-col gap-8">
		<NodeLiveMetrics />
		<NodeTrafficStats embed />
	</div>
{:else if tab === 'http'}
	<HttpMetrics />
{:else}
	<PanelMetrics />
{/if}
