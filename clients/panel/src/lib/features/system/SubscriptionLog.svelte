<script lang="ts">
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
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { formatDate, formatNumber } from '$lib/format';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type SrhRow = {
		id: number;
		userId: number;
		srrResponseType?: string;
		srrRuleName?: string | null;
		requestIp?: string | null;
		userAgent?: string | null;
		requestAt?: string;
	};

	let start = $state(0);
	let size = $state(loadPageSize('srh', 25));
	let search = $state('');
	let searchDraft = $state('');

	const list = rw.subscription.requestHistory.list(() => ({
		start,
		size,
		search: search || undefined
	}));
	const statsQ = rw.subscription.requestHistory.stats();

	const colPrefs = loadColumnPrefs('srh');
	const handler = new TableHandler<SrhRow>([], { rowsPerPage: size, selectBy: 'id' });
	const table = handler as TableHandlerInterface<SrhRow>;
	const view = handler.createView([
		{ index: 0, name: 'Time', isVisible: columnVisible(colPrefs, 'Time') },
		{ index: 1, name: 'User', isVisible: columnVisible(colPrefs, 'User') },
		{ index: 2, name: 'Type', isVisible: columnVisible(colPrefs, 'Type') },
		{ index: 3, name: 'IP', isVisible: columnVisible(colPrefs, 'IP') }
	]);

	let selected = $state<SrhRow | null>(null);
	const records = $derived(asArray<SrhRow>(list.data, ['records']));
	const total = $derived((list.data as { total?: number } | undefined)?.total ?? records.length);
	const stats = $derived((statsQ.data ?? {}) as { byParsedApp?: { app: string; count: number }[] });

	syncTableRows(handler, records, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, records, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		const next = handler.rowsPerPage;
		if (next !== size) {
			size = next;
			start = 0;
			savePageSize('srh', next);
		}
	});

	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	function onsearch() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			search = searchDraft.trim();
			start = 0;
		}, 200);
	}

	$effect.pre(() => {
		const token = pageChrome.set({
			search: {
				placeholder: 'Search history...',
				value: () => searchDraft,
				setValue: (value) => {
					searchDraft = value;
				},
				oninput: onsearch
			},
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('srh', view) }
		});
		return () => pageChrome.clear(token);
	});

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount, 1));
	const topApps = $derived((stats.byParsedApp ?? []).slice(0, 4));

	function openRow(row: SrhRow, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input')) return;
		selected = row;
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && total === 0}
	<EmptyState title="No subscription requests" description="History appears here as clients fetch subscriptions." />
{:else}
	<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
		<div class={chrome.tile}>
			<p class={chrome.hint}>Records</p>
			<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{formatNumber(total)}</p>
		</div>
		{#each topApps as app}
			<div class={chrome.tile}>
				<p class={chrome.hint}>{app.app || 'App'}</p>
				<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{formatNumber(app.count)}</p>
			</div>
		{/each}
	</div>

	<DataTable {table} headless>
		<div aria-busy={list.loading}>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						{#if view.columns[0]?.isVisible}
							<ThSort {table} field="requestAt">Time</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field="userId">User</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort {table} field="srrResponseType">Type</ThSort>
						{/if}
						{#if view.columns[3]?.isVisible}
							<ThLabel>Client</ThLabel>
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
									{formatDate(row.requestAt)}
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3">{row.userId}</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3">
									<span class="text-[var(--color-text-secondary)]">{row.srrResponseType ?? '—'}</span>
									{#if row.srrRuleName}
										<CellSubtitle>{row.srrRuleName}</CellSubtitle>
									{/if}
								</Table.Cell>
							{/if}
							{#if view.columns[3]?.isVisible}
								<Table.Cell class="max-w-[16rem] px-3">
									<span class="block truncate font-mono text-xs" title={row.requestIp ?? ''}>{row.requestIp ?? '—'}</span>
									<CellSubtitle title={row.userAgent ?? ''}>{row.userAgent ?? '—'}</CellSubtitle>
								</Table.Cell>
							{/if}
						</Table.Row>
					{:else}
						{#if list.loading}
							<TableSkeletonRows columns={emptyColspan} />
						{:else}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={emptyColspan} class="py-10 text-center text-sm text-[var(--color-text-tertiary)]">
									No requests match the current table state.
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
	title={selected ? `Request #${selected.id}` : 'Request'}
	description={selected ? formatDate(selected.requestAt) : ''}
	actions={[]}
	onrun={() => {}}
	onOpenChange={(value) => {
		if (!value) selected = null;
	}}
>
	{#if selected}
		<div class={chrome.stack}>
			<div class={chrome.tile}>
				<p class={chrome.hint}>User</p>
				<p class="mt-1 text-sm font-medium">{selected.userId}</p>
			</div>
			<div class={chrome.tile}>
				<p class={chrome.hint}>Type</p>
				<p class="mt-1 text-sm font-medium">{selected.srrResponseType ?? '—'}</p>
				{#if selected.srrRuleName}
					<p class={chrome.hint}>{selected.srrRuleName}</p>
				{/if}
			</div>
			<div class={chrome.tile}>
				<p class={chrome.hint}>IP</p>
				<p class="mt-1 font-mono text-xs">{selected.requestIp ?? '—'}</p>
			</div>
			<div class={chrome.tile}>
				<p class={chrome.hint}>User-Agent</p>
				<p class="mt-1 break-all text-xs text-[var(--color-text-secondary)]">{selected.userAgent ?? '—'}</p>
			</div>
		</div>
	{/if}
</DetailSheet>
