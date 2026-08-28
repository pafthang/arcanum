<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		children
	}: {
		title?: string;
		description?: string;
		children?: Snippet;
	} = $props();
</script>

<section class="mt-8">
	{#if title}
		<h2 class="text-sm font-medium text-[var(--color-text-secondary)]">{title}</h2>
	{/if}
	{#if description}
		<p class="mt-1 text-xs text-[var(--color-text-tertiary)]">{description}</p>
	{/if}
	<div class="mt-3 divide-y divide-[var(--app-border)] overflow-hidden rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]">
		{@render children?.()}
	</div>
</section>
