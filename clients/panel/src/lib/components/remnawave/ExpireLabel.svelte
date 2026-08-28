<script lang="ts">
	import { expireMeta } from '$lib/format';

	let { value }: { value: string | null | undefined } = $props();
	const meta = $derived(expireMeta(value));
	const tone = $derived(
		meta.tone === 'expired'
			? 'text-[var(--color-error)]'
			: meta.tone === 'soon'
				? 'text-amber-500'
				: 'text-[var(--color-text-secondary)]'
	);
</script>

<div class="{tone} min-w-0">
	<span class="text-sm">{meta.label}</span>
	{#if meta.detail}
		<div class="w-0 min-w-full truncate text-[11px] leading-tight text-[var(--color-text-tertiary)]" title={meta.detail}>
			{meta.detail}
		</div>
	{/if}
</div>
