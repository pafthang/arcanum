<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import {
		DataTable,
		TableStatusBar,
		ThSort,
		TableSkeletonRows,
		syncTableRows
	} from '$lib/components/ui/data-table';
	import { chrome } from '$lib/components/remnawave/chrome';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { loadPageSize, savePageSize } from '$lib/features/layout/table-prefs';
	import { formatNumber } from '$lib/format';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type RouteStat = { method: string; route: string; count: number; key?: string };

	const q = rw.system.httpStats();
	const handler = new TableHandler<RouteStat>([], { rowsPerPage: loadPageSize('http-stats', 25) });
	const table = handler as TableHandlerInterface<RouteStat>;

	const total = $derived((q.data as { total?: number } | undefined)?.total ?? 0);
	const rows = $derived(
		asArray<RouteStat>(q.data, ['routes']).map((row) => ({
			...row,
			key: `${row.method}:${row.route}`
		}))
	);

	syncTableRows(handler, rows, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, rows, { ignoreEmpty: q.loading });
	});
	$effect(() => {
		savePageSize('http-stats', handler.rowsPerPage);
	});
	$effect.pre(() => {
		const token = pageChrome.set({
			search: { table, placeholder: 'Search routes...', fields: ['method', 'route'] },
			toolbar: { table }
		});
		return () => pageChrome.clear(token);
	});
</script>

{#if q.error && !q.data}
	<ErrorState message={q.error.message} onretry={() => q.refetch()} />
{:else if !q.loading && rows.length === 0}
	<EmptyState title="No HTTP stats" description="Route counters appear after the API serves traffic." />
{:else}
	<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
		<div class={chrome.tile}>
			<p class={chrome.hint}>Total hits</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(total)}</p>
		</div>
		<div class={chrome.tile}>
			<p class={chrome.hint}>Routes</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(rows.length)}</p>
		</div>
	</div>
	<DataTable {table} headless>
		<div aria-busy={q.loading}>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						<ThSort {table} field="method">Method</ThSort>
						<ThSort {table} field="route">Route</ThSort>
						<ThSort {table} field="count" class="text-right" btnClass="justify-end">Count</ThSort>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.key)}
						<Table.Row class="border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)]">
							<Table.Cell class="px-3 font-mono text-xs">{row.method}</Table.Cell>
							<Table.Cell class="max-w-[28rem] px-3">
								<span class="block truncate font-mono text-xs" title={row.route}>{row.route}</span>
							</Table.Cell>
							<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">{formatNumber(row.count)}</Table.Cell>
						</Table.Row>
					{:else}
						{#if q.loading}
							<TableSkeletonRows columns={3} />
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
		{#snippet footer()}
			<TableStatusBar {table} loading={q.loading} />
		{/snippet}
	</DataTable>
{/if}
