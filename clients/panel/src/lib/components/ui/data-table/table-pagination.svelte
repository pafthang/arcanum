<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import { Button } from '$lib/components/ui/button';
	import { cn } from '$lib/utils.js';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';

	type Props = {
		table: TableHandlerInterface<Row>;
	};

	let { table }: Props = $props();

	const pagesWithEllipsis = $derived(table.pagesWithEllipsis as ReadonlyArray<number | null>);
	const isFirstPage = $derived(table.currentPage === 1);
	const isLastPage = $derived(table.currentPage === table.pageCount || table.pageCount === 0);
</script>

<nav class="flex flex-wrap items-center gap-1" aria-label="Pagination">
	<Button
		type="button"
		variant="ghost"
		size="icon-sm"
		disabled={isFirstPage}
		aria-label="Previous page"
		onclick={() => table.setPage('previous')}
	>
		<ChevronLeftIcon class="size-4" />
	</Button>
	{#each pagesWithEllipsis as page, index (`${page ?? 'ellipsis'}-${index}`)}
		{#if page === null}
			<span class="px-1 text-xs text-[var(--color-text-tertiary)]">…</span>
		{:else}
			<Button
				type="button"
				variant={table.currentPage === page ? 'secondary' : 'ghost'}
				size="icon-sm"
				class={cn('text-xs', table.currentPage === page && 'font-medium')}
				aria-current={table.currentPage === page ? 'page' : undefined}
				onclick={() => table.setPage(page)}
			>
				{page}
			</Button>
		{/if}
	{/each}
	<Button
		type="button"
		variant="ghost"
		size="icon-sm"
		disabled={isLastPage}
		aria-label="Next page"
		onclick={() => table.setPage('next')}
	>
		<ChevronRightIcon class="size-4" />
	</Button>
</nav>
