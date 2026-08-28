<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		children
	}: {
		title: string;
		description?: string;
		children?: Snippet;
	} = $props();
</script>

<div class="flex items-center justify-between gap-4 px-5 py-4">
	<div class="min-w-0">
		<p class="text-sm font-medium text-[var(--color-text-primary)]">{title}</p>
		{#if description}
			<p class="mt-0.5 text-xs text-[var(--color-text-tertiary)]">{description}</p>
		{/if}
	</div>
	<div class="shrink-0">{@render children?.()}</div>
</div>
