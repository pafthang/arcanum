<script lang="ts">
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.js';

	let {
		total,
		start,
		size,
		onstart
	}: {
		total: number;
		start: number;
		size: number;
		onstart: (next: number) => void;
	} = $props();

	const pageCount = $derived(Math.max(1, Math.ceil((total || 0) / Math.max(size, 1))));
	const current = $derived(Math.floor(start / Math.max(size, 1)) + 1);
	const from = $derived(total === 0 ? 0 : start + 1);
	const to = $derived(Math.min(start + size, total));

	const pages = $derived.by(() => {
		const out: (number | null)[] = [];
		for (let page = 1; page <= pageCount; page += 1) {
			if (page === 1 || page === pageCount || Math.abs(page - current) <= 1) {
				out.push(page);
			} else if (out.at(-1) !== null) {
				out.push(null);
			}
		}
		return out;
	});
</script>

<div class="flex flex-wrap items-center gap-3">
	<p class="text-[11px] leading-8 text-[var(--color-text-tertiary)]">
		{#if total > 0}
			{from}–{to} of {total}
		{:else}
			No entries
		{/if}
	</p>
	<nav class="flex flex-wrap items-center gap-1" aria-label="Pagination">
		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			disabled={current <= 1}
			aria-label="Previous page"
			onclick={() => onstart(Math.max(0, start - size))}
		>
			<ChevronLeftIcon class="size-4" />
		</Button>
		{#each pages as page, index (`${page ?? 'ellipsis'}-${index}`)}
			{#if page === null}
				<span class="px-1 text-xs text-[var(--color-text-tertiary)]">…</span>
			{:else}
				<Button
					type="button"
					variant={current === page ? 'secondary' : 'ghost'}
					size="icon-sm"
					class={cn('text-xs', current === page && 'font-medium')}
					aria-current={current === page ? 'page' : undefined}
					onclick={() => onstart((page - 1) * size)}
				>
					{page}
				</Button>
			{/if}
		{/each}
		<Button
			type="button"
			variant="ghost"
			size="icon-sm"
			disabled={current >= pageCount}
			aria-label="Next page"
			onclick={() => onstart(start + size)}
		>
			<ChevronRightIcon class="size-4" />
		</Button>
	</nav>
</div>
