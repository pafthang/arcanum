<script lang="ts">
	import { chrome } from '$lib/components/remnawave/chrome';

	let {
		cards,
		compact = false
	}: {
		cards: {
			label: string;
			value: string;
			sub?: string;
			href?: string;
			trend?: { text: string; tone: 'up' | 'down' | 'neutral' };
		}[];
		compact?: boolean;
	} = $props();
</script>

<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
	{#each cards as card}
		{@const classes = `${chrome.tile} ${card.href ? 'transition-colors hover:border-[var(--app-accent)]/40 hover:bg-[var(--color-bg-hover)]/40' : ''}`}
		{#if card.href}
			<a href={card.href} class="block {classes}">
				<p class={chrome.hint}>{card.label}</p>
				<p class="{compact ? 'text-base' : 'text-lg'} mt-1 font-semibold tracking-tight text-[var(--color-text-primary)]">
					{card.value}
				</p>
				{#if card.trend}
					<p
						class="mt-1 text-[11px] {card.trend.tone === 'up'
							? 'text-emerald-500'
							: card.trend.tone === 'down'
								? 'text-[var(--color-error)]'
								: 'text-[var(--color-text-tertiary)]'}"
					>
						{card.trend.text}
					</p>
				{:else if card.sub}
					<p class="mt-1 {chrome.hint}">{card.sub}</p>
				{/if}
			</a>
		{:else}
			<div class={classes}>
				<p class={chrome.hint}>{card.label}</p>
				<p class="{compact ? 'text-base' : 'text-lg'} mt-1 font-semibold tracking-tight text-[var(--color-text-primary)]">
					{card.value}
				</p>
				{#if card.trend}
					<p
						class="mt-1 text-[11px] {card.trend.tone === 'up'
							? 'text-emerald-500'
							: card.trend.tone === 'down'
								? 'text-[var(--color-error)]'
								: 'text-[var(--color-text-tertiary)]'}"
					>
						{card.trend.text}
					</p>
				{:else if card.sub}
					<p class="mt-1 {chrome.hint}">{card.sub}</p>
				{/if}
			</div>
		{/if}
	{/each}
</div>
