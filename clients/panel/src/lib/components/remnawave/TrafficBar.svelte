<script lang="ts">
	import { formatBytes, formatTrafficLimit, trafficPercent } from '$lib/format';

	let {
		used = 0,
		limit = 0,
		compact = false
	}: {
		used?: number | null;
		limit?: number | null;
		compact?: boolean;
	} = $props();

	const pct = $derived(trafficPercent(used, limit));
	const bar = $derived(pct == null ? 0 : pct);
	const hot = $derived(pct != null && pct >= 90);
</script>

<div class={compact ? 'min-w-[7rem]' : ''}>
	<div class="{compact ? 'text-right' : 'text-left'} text-[var(--color-text-secondary)]">
		{formatBytes(used)}
		<span class="text-[var(--color-text-tertiary)]"> / {formatTrafficLimit(limit)}</span>
	</div>
	{#if pct != null}
		<div class="mt-1 h-1 overflow-hidden rounded-full bg-[var(--color-bg-tertiary)]">
			<div
				class="h-full rounded-full {hot ? 'bg-[var(--color-error)]' : 'bg-[var(--app-accent)]'}"
				style="width: {bar}%"
			></div>
		</div>
	{/if}
</div>
