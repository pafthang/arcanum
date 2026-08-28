<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import {
		DataTable,
		TableStatusBar,
		ThSort,
		TableSkeletonRows,
		syncTableRows
	} from '$lib/components/ui/data-table';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { loadPageSize, savePageSize } from '$lib/features/layout/table-prefs';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	let { embed = false }: { embed?: boolean } = $props();

	type NodeDay = { nodeName: string; date: string; totalBytes: string; key?: string };

	const q = rw.system.nodesStatistics();
	const handler = new TableHandler<NodeDay>([], { rowsPerPage: loadPageSize('node-stats', 20) });
	const table = handler as TableHandlerInterface<NodeDay>;

	const rows = $derived(
		asArray<NodeDay>(q.data, ['lastSevenDays']).map((row, index) => ({
			...row,
			key: `${row.nodeName}:${row.date}:${index}`
		}))
	);

	syncTableRows(handler, rows, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, rows, { ignoreEmpty: q.loading });
	});
	$effect(() => {
		savePageSize('node-stats', handler.rowsPerPage);
	});

	$effect.pre(() => {
		if (embed) return;
		const token = pageChrome.set({
			search: { table, placeholder: 'Search nodes...', fields: ['nodeName', 'date'] },
			toolbar: { table }
		});
		return () => pageChrome.clear(token);
	});
</script>

{#if q.error && !q.data}
	<ErrorState message={q.error.message} onretry={() => q.refetch()} />
{:else if !q.loading && rows.length === 0}
	<EmptyState title="No node statistics" description="Usage totals appear after nodes report traffic." />
{:else}
	{#if embed}
		<h2 class="mb-3 text-[13px] font-medium text-[var(--color-text-secondary)]">Traffic, last 7 days</h2>
	{/if}
	<DataTable {table} headless>
		<div aria-busy={q.loading}>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						<ThSort {table} field="nodeName">Node</ThSort>
						<ThSort {table} field="date">Date</ThSort>
						<ThSort {table} field="totalBytes" class="text-right" btnClass="justify-end">Traffic</ThSort>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.key)}
						<Table.Row class="border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)]">
							<Table.Cell class="px-3 font-medium">{row.nodeName}</Table.Cell>
							<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{row.date}</Table.Cell>
							<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">{row.totalBytes}</Table.Cell>
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
