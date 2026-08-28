<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as Table from '$lib/components/ui/table';
	import {
		DataTable,
		TableStatusBar,
		ThLabel,
		ThSort,
		TableSkeletonRows,

		syncTableRows
	} from '$lib/components/ui/data-table';
	import CellSubtitle from '$lib/components/remnawave/CellSubtitle.svelte';
	import ConfirmAction from '$lib/components/remnawave/ConfirmAction.svelte';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { formatDate, formatNumber } from '$lib/format';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray, pretty } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type TorrentRow = {
		id: number;
		userId: number;
		nodeId: number;
		user?: { username?: string };
		node?: { name?: string; uuid?: string; countryCode?: string };
		report?: unknown;
		createdAt?: string;
	};

	let start = $state(0);
	let size = $state(loadPageSize('tblocker', 25));
	let confirmOpen = $state(false);

	const reports = rw.plugins.torrentBlocker.reports(() => ({ start, size }));
	const statsQ = rw.plugins.torrentBlocker.stats();
	const truncate = rw.plugins.torrentBlocker.truncate();

	const colPrefs = loadColumnPrefs('tblocker');
	const handler = new TableHandler<TorrentRow>([], { rowsPerPage: size, selectBy: 'id' });
	const table = handler as TableHandlerInterface<TorrentRow>;
	const view = handler.createView([
		{ index: 0, name: 'Time', isVisible: columnVisible(colPrefs, 'Time') },
		{ index: 1, name: 'User', isVisible: columnVisible(colPrefs, 'User') },
		{ index: 2, name: 'Node', isVisible: columnVisible(colPrefs, 'Node') }
	]);

	let selected = $state<TorrentRow | null>(null);
	const records = $derived(asArray<TorrentRow>(reports.data, ['records']));
	const total = $derived((reports.data as { total?: number } | undefined)?.total ?? records.length);
	const stats = $derived(
		((statsQ.data as { stats?: Record<string, number> } | undefined)?.stats ?? {}) as Record<string, number>
	);

	syncTableRows(handler, records, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, records, { ignoreEmpty: reports.loading });
	});
	$effect(() => {
		const next = handler.rowsPerPage;
		if (next !== size) {
			size = next;
			start = 0;
			savePageSize('tblocker', next);
		}
	});
	$effect.pre(() => {
		const token = pageChrome.set({
			search: { table, placeholder: 'Search reports...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('tblocker', view) }
		});
		return () => pageChrome.clear(token);
	});

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount, 1));
	const reportJson = $derived(selected ? pretty(selected.report ?? selected) : '{}');

	function openRow(row: TorrentRow, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input')) return;
		selected = row;
	}
</script>

{#if reports.error && !reports.data}
	<ErrorState message={reports.error.message} onretry={() => reports.refetch()} />
{:else if !reports.loading && total === 0}
	<EmptyState title="No torrent reports" description="Blocked torrent events from node plugins show up here." />
{:else}
	<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class={chrome.tile}>
			<p class={chrome.hint}>Reports</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(stats.totalReports ?? total)}</p>
		</div>
		<div class={chrome.tile}>
			<p class={chrome.hint}>Last 24h</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(stats.reportsLast24Hours ?? 0)}</p>
		</div>
		<div class={chrome.tile}>
			<p class={chrome.hint}>Users</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(stats.distinctUsers ?? 0)}</p>
		</div>
		<div class={chrome.tile}>
			<p class={chrome.hint}>Nodes</p>
			<p class="mt-1 text-sm font-medium">{formatNumber(stats.distinctNodes ?? 0)}</p>
		</div>
	</div>

	<DataTable {table} headless>
		{#snippet header()}
			<div class="mb-3 flex w-full min-w-0 flex-wrap items-center justify-end gap-2">
				<Button size="sm" variant="destructive" onclick={() => (confirmOpen = true)}>Truncate</Button>
			</div>
		{/snippet}

		<div aria-busy={reports.loading}>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						{#if view.columns[0]?.isVisible}
							<ThSort {table} field="createdAt">Time</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field="userId">User</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThLabel>Node</ThLabel>
						{/if}
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.id)}
						<Table.Row
							class="cursor-pointer border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(row, event)}
						>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="px-3 whitespace-nowrap text-[var(--color-text-secondary)]">
									{formatDate(row.createdAt)}
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3">
									<span class="font-medium">{row.user?.username ?? row.userId}</span>
									<CellSubtitle>#{row.userId}</CellSubtitle>
								</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="max-w-[16rem] px-3">
									<span class="block truncate">{row.node?.name ?? row.nodeId}</span>
									<CellSubtitle>{row.node?.countryCode ?? ''}</CellSubtitle>
								</Table.Cell>
							{/if}
						</Table.Row>
					{:else}
						{#if reports.loading}
							<TableSkeletonRows columns={emptyColspan} />
						{:else}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={emptyColspan} class="py-10 text-center text-sm text-[var(--color-text-tertiary)]">
									No reports match the current table state.
								</Table.Cell>
							</Table.Row>
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		{#snippet footer()}
			<TableStatusBar {total} {start} {size} onstart={(next) => (start = next)} />
		{/snippet}
	</DataTable>
{/if}

<DetailSheet
	open={selected !== null}
	title={selected ? `Report #${selected.id}` : 'Report'}
	description={selected ? formatDate(selected.createdAt) : ''}
	actions={[]}
	onrun={() => {}}
	onOpenChange={(value) => {
		if (!value) selected = null;
	}}
>
	{#if selected}
		<div class={chrome.stack}>
			<div class="grid grid-cols-2 gap-3">
				<div class={chrome.tile}>
					<p class={chrome.hint}>User</p>
					<p class="mt-1 text-sm font-medium">{selected.user?.username ?? selected.userId}</p>
				</div>
				<div class={chrome.tile}>
					<p class={chrome.hint}>Node</p>
					<p class="mt-1 text-sm font-medium">{selected.node?.name ?? selected.nodeId}</p>
				</div>
			</div>
			<div class={chrome.field}>
				<span class={chrome.label}>Report</span>
				<pre class="max-h-80 overflow-auto rounded-md border border-[var(--app-border)] p-3 font-mono text-xs text-[var(--color-text-secondary)]">{reportJson}</pre>
			</div>
		</div>
	{/if}
</DetailSheet>

<ConfirmAction
	bind:open={confirmOpen}
	title="Truncate all torrent reports?"
	description="This deletes every stored blocker event. It cannot be undone."
	confirmLabel="Truncate"
	onconfirm={async () => {
		await truncate.mutate();
		appToast.success('Truncated');
		await reports.refetch();
		await statsQ.refetch();
	}}
/>
