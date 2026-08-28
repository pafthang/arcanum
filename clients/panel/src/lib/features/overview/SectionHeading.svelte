<script lang="ts">
	import type { ComponentType, Snippet } from 'svelte';
	import { chrome } from '$lib/components/remnawave/chrome';

	let {
		icon: Icon,
		title,
		description,
		actions
	}: {
		icon: ComponentType;
		title: string;
		description?: string;
		actions?: Snippet;
	} = $props();
</script>

<div class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
	<div class="flex items-start gap-2.5">
		<Icon size={15} class="mt-0.5 text-[var(--app-accent-light)]" />
		<div>
			<h2 class={chrome.title}>{title}</h2>
			{#if description}
				<p class={chrome.description}>{description}</p>
			{/if}
		</div>
	</div>
	{#if actions}
		<div class="flex shrink-0 flex-wrap items-center gap-2">{@render actions()}</div>
	{/if}
</div>
