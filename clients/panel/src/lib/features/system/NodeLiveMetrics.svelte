<script lang="ts">
	import * as Table from '$lib/components/ui/table';
	import {
		DataTable,
		TableStatusBar,
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
	import { loadPageSize, savePageSize } from '$lib/features/layout/table-prefs';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type Traffic = { tag: string; upload: string; download: string };
	type NodeMetric = {
		nodeUuid: string;
		nodeName: string;
		countryEmoji?: string;
		providerName?: string;
		usersOnline?: number;
		inboundsStats?: Traffic[];
		outboundsStats?: Traffic[];
	};

	const q = rw.system.nodesMetrics();
	const handler = new TableHandler<NodeMetric>([], {
		rowsPerPage: loadPageSize('node-metrics', 20),
		selectBy: 'nodeUuid'
	});
	const table = handler as TableHandlerInterface<NodeMetric>;
	const rows = $derived(asArray<NodeMetric>(q.data, ['nodes']));
	let selected = $state<NodeMetric | null>(null);

	syncTableRows(handler, rows, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, rows, { ignoreEmpty: q.loading });
	});
	$effect(() => {
		savePageSize('node-metrics', handler.rowsPerPage);
	});
	$effect.pre(() => {
		const token = pageChrome.set({
			search: { table, placeholder: 'Search nodes...', fields: ['nodeName', 'providerName'] },
			toolbar: { table }
		});
		return () => pageChrome.clear(token);
	});

	function openRow(row: NodeMetric, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input')) return;
		selected = row;
	}
</script>

{#if q.error && !q.data}
	<ErrorState message={q.error.message} onretry={() => q.refetch()} />
{:else if !q.loading && rows.length === 0}
	<EmptyState title="No live metrics" description="Connected nodes publish inbound/outbound counters here." />
{:else}
	<DataTable {table} headless>
		<div aria-busy={q.loading}>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						<ThSort {table} field="nodeName">Node</ThSort>
						<ThSort {table} field={(row) => row.usersOnline ?? 0} class="text-right" btnClass="justify-end">Online</ThSort>
						<ThSort {table} field={(row) => row.inboundsStats?.length ?? 0} class="text-right" btnClass="justify-end">
							Inbounds
						</ThSort>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.nodeUuid)}
						{@const tags = (row.inboundsStats ?? []).map((item) => item.tag).join(' · ')}
						<Table.Row
							class="cursor-pointer border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(row, event)}
						>
							<Table.Cell class="max-w-[18rem] px-3">
								<span class="font-medium">{row.countryEmoji ?? ''} {row.nodeName}</span>
								<CellSubtitle>{row.providerName ?? row.nodeUuid}</CellSubtitle>
							</Table.Cell>
							<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">{row.usersOnline ?? 0}</Table.Cell>
							<Table.Cell class="max-w-[14rem] px-3 text-right">
								<span class="text-[var(--color-text-secondary)]">{row.inboundsStats?.length ?? 0}</span>
								{#if tags}
									<CellSubtitle title={tags}>{tags}</CellSubtitle>
								{/if}
							</Table.Cell>
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

<DetailSheet
	open={selected !== null}
	title={selected ? selected.nodeName : 'Node metrics'}
	description={selected?.providerName ?? ''}
	actions={[]}
	onrun={() => {}}
	onOpenChange={(value) => {
		if (!value) selected = null;
	}}
>
	{#if selected}
		<div class={chrome.stack}>
			<div class={chrome.tile}>
				<p class={chrome.hint}>Online</p>
				<p class="mt-1 text-sm font-medium">{selected.usersOnline ?? 0}</p>
			</div>
			<div>
				<p class={chrome.section}>Inbounds</p>
				<ul class="mt-2 space-y-1 text-sm">
					{#each selected.inboundsStats ?? [] as item}
						<li class="flex justify-between gap-3 text-[var(--color-text-secondary)]">
							<span class="truncate">{item.tag}</span>
							<span class="shrink-0">{item.download} / {item.upload}</span>
						</li>
					{:else}
						<li class="text-[var(--color-text-tertiary)]">None</li>
					{/each}
				</ul>
			</div>
			<div>
				<p class={chrome.section}>Outbounds</p>
				<ul class="mt-2 space-y-1 text-sm">
					{#each selected.outboundsStats ?? [] as item}
						<li class="flex justify-between gap-3 text-[var(--color-text-secondary)]">
							<span class="truncate">{item.tag}</span>
							<span class="shrink-0">{item.download} / {item.upload}</span>
						</li>
					{:else}
						<li class="text-[var(--color-text-tertiary)]">None</li>
					{/each}
				</ul>
			</div>
		</div>
	{/if}
</DetailSheet>
