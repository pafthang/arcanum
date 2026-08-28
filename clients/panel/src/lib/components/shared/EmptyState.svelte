<script lang="ts">
	import { Plus } from 'lucide-svelte';

	let {
		title,
		description,
		action
	}: {
		title: string;
		description?: string;
		action?: { label: string; onclick: () => void };
	} = $props();
</script>

<div class="flex min-h-[calc(100dvh-11rem)] flex-col items-center justify-center px-4">
	<div class="flex w-full max-w-sm flex-col items-center px-8 py-12 text-center">
		{#if action}
			<button
				type="button"
				class="mb-4 flex size-12 items-center justify-center rounded-full border border-[var(--app-border)] bg-[var(--color-bg)] text-[var(--app-accent)] shadow-[0_0_24px_color-mix(in_srgb,var(--app-accent)_18%,transparent)] transition-colors hover:border-[var(--app-border-hover)] hover:bg-[var(--color-bg-hover)]"
				title={action.label}
				aria-label={action.label}
				onclick={action.onclick}
			>
				<Plus class="size-5" />
			</button>
		{/if}
		<h2 class="text-sm font-medium text-[var(--color-text-primary)]">{title}</h2>
		{#if description}
			<p class="mt-1.5 text-[13px] leading-5 text-[var(--color-text-tertiary)]">{description}</p>
		{/if}
	</div>
</div>
