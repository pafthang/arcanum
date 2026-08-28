<script lang="ts">
	import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
	import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';
	import FilterPills from '$lib/components/remnawave/FilterPills.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import TableToolbar from './table-toolbar.svelte';

	let {
		table,
		total = 0,
		start = 0,
		size = 0,
		onstart,
		loading = false,
		selection = false
	}: {
		table?: TableHandlerInterface<import('@vincjo/datatables/server').Row>;
		total?: number;
		start?: number;
		size?: number;
		onstart?: (next: number) => void;
		loading?: boolean;
		selection?: boolean;
	} = $props();

	const count = $derived.by(() => {
		if (table) {
			const { start: from, end: to, total: all, selected } = table.rowCount;
			return { from, to, total: all, selected };
		}
		const all = total || 0;
		const step = Math.max(size, 1);
		return {
			from: all === 0 ? 0 : start + 1,
			to: Math.min(start + step, all),
			total: all,
			selected: 0
		};
	});

	const pager = $derived.by(() => {
		if (table) {
			const current = table.currentPage;
			const pages = Math.max(table.pageCount, 1);
			return {
				current,
				pages,
				canPrev: current > 1,
				canNext: current < table.pageCount && table.pageCount > 0,
				prev: () => table.setPage('previous'),
				next: () => table.setPage('next')
			};
		}
		const step = Math.max(size, 1);
		const pages = Math.max(1, Math.ceil((total || 0) / step));
		const current = Math.floor(start / step) + 1;
		return {
			current,
			pages,
			canPrev: current > 1,
			canNext: current < pages,
			prev: () => onstart?.(Math.max(0, start - step)),
			next: () => onstart?.(start + step)
		};
	});

	const chrome = $derived(pageChrome.toolbar);
	const showSettings = $derived(Boolean(chrome) && (table == null || chrome.table === table));
	const filter = $derived(chrome?.filter);
	const filterValue = $derived(filter?.value() ?? '');

	const iconBtn =
		'inline-flex size-7 items-center justify-center rounded-md text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] disabled:pointer-events-none disabled:opacity-40';
</script>

<footer
	data-slot="table-status-bar"
	class="flex h-8 shrink-0 items-center justify-between gap-3 border-t border-[var(--app-border)] bg-[var(--color-bg)] px-1.5 text-[11px] text-[var(--color-text-secondary)]"
>
	<div class="flex min-w-0 items-center gap-2 px-1">
		{#if loading && count.total === 0}
			<span>Loading…</span>
		{:else}
			{#if selection && table && count.selected > 0}
				<span class="inline-flex items-center gap-1.5">
					<span class="font-medium text-[var(--color-text-primary)]">{count.selected}</span>
					selected
					<button
						type="button"
						class="rounded-md px-1.5 py-0.5 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
						onclick={() => table.clearSelection()}
					>
						Clear
					</button>
				</span>
				<span class="h-3.5 w-px bg-[var(--app-border)]" aria-hidden="true"></span>
			{/if}
			<span class="truncate tabular-nums">
				{#if count.total > 0}
					{count.from}–{count.to}
					<span class="text-[var(--color-text-tertiary)]"> of {count.total}</span>
				{:else}
					No entries
				{/if}
			</span>
		{/if}
	</div>
	<div class="flex shrink-0 items-center gap-1.5">
		<nav class="flex items-center" aria-label="Pagination">
			<button type="button" class={iconBtn} disabled={!pager.canPrev} aria-label="Previous page" onclick={pager.prev}>
				<ChevronLeftIcon class="size-3.5" />
			</button>
			<span class="min-w-12 px-0.5 text-center tabular-nums" title="Current page">
				{pager.current}
				<span class="text-[var(--color-text-tertiary)]"> of {pager.pages}</span>
			</span>
			<button type="button" class={iconBtn} disabled={!pager.canNext} aria-label="Next page" onclick={pager.next}>
				<ChevronRightIcon class="size-3.5" />
			</button>
		</nav>
		{#if showSettings && chrome}
			<span class="mx-0.5 h-3.5 w-px bg-[var(--app-border)]" aria-hidden="true"></span>
			<div class="flex items-center">
				{#if filter}
					<FilterPills
						value={filterValue}
						options={filter.options}
						onselect={filter.onselect}
						embedded
					/>
				{/if}
				<TableToolbar
					table={chrome.table as never}
					view={chrome.view}
					onviewchange={chrome.onviewchange}
					embedded
				/>
			</div>
		{/if}
	</div>
</footer>
