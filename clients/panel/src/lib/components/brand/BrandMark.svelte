<script lang="ts">
	import { APP_NAME, APP_NAME_LEAD, APP_NAME_TAIL } from '$lib/brand';

	let {
		size = 'md',
		wordmark = true,
		title = APP_NAME
	}: {
		size?: 'sm' | 'md' | 'lg';
		wordmark?: boolean;
		title?: string;
	} = $props();

	const iconClass = $derived(
		size === 'lg' ? 'size-10' : size === 'sm' ? 'size-7' : 'size-8'
	);
	const textClass = $derived(
		size === 'lg' ? 'text-3xl tracking-tight' : size === 'sm' ? 'text-sm' : 'text-2xl tracking-tight'
	);
	const branded = $derived(!title || title === APP_NAME);
</script>

<div class="flex items-center gap-2.5">
	<svg
		class="{iconClass} shrink-0 text-[var(--app-accent)]"
		viewBox="0 0 16 16"
		fill="none"
		xmlns="http://www.w3.org/2000/svg"
		aria-hidden="true"
	>
		<path
			fill-rule="evenodd"
			clip-rule="evenodd"
			d="M8 1a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-1.5 0V1.75A.75.75 0 0 1 8 1Zm6 2a.75.75 0 0 1 .75.75v8.5a.75.75 0 0 1-1.5 0v-8.5A.75.75 0 0 1 14 3ZM5 4a.75.75 0 0 1 .75.75v6.5a.75.75 0 0 1-1.5 0v-6.5A.75.75 0 0 1 5 4Zm6 1a.75.75 0 0 1 .75.75v4.5a.75.75 0 0 1-1.5 0v-4.5A.75.75 0 0 1 11 5ZM2 6a.75.75 0 0 1 .75.75v2.5a.75.75 0 0 1-1.5 0v-2.5A.75.75 0 0 1 2 6Z"
			fill="currentColor"
		/>
	</svg>
	{#if wordmark}
		{#if branded}
			<span class="{textClass} font-semibold leading-none">
				<span class="text-[var(--app-accent)]">{APP_NAME_LEAD}</span><span class="text-[var(--color-text-primary)]">{APP_NAME_TAIL}</span>
			</span>
		{:else}
			<span class="{textClass} font-semibold leading-none text-[var(--color-text-primary)]">{title}</span>
		{/if}
	{/if}
</div>
