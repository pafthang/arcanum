<script lang="ts">
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Button } from '$lib/components/ui/button';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
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
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import ProfilesBulkBar from '$lib/features/profiles/ProfilesBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { formatDate, formatNumber } from '$lib/format';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { HwidDevice } from '@optimawave/ts-back';

	type HwidRow = HwidDevice & { key: string };

	let start = $state(0);
	let size = $state(loadPageSize('hwid', 50));

	const list = rw.hwid.list(() => ({ start, size }));
	const stats = rw.hwid.stats();
	const removeM = rw.hwid.remove();

	const colPrefs = loadColumnPrefs('hwid');
	const handler = new TableHandler<HwidRow>([], {
		rowsPerPage: loadPageSize('hwid', 50),
		selectBy: 'key'
	});
	const table = handler as TableHandlerInterface<HwidRow>;
	const view = handler.createView([
		{ index: 1, name: 'HWID', isVisible: columnVisible(colPrefs, 'HWID') },
		{ index: 2, name: 'User', isVisible: columnVisible(colPrefs, 'User') },
		{ index: 3, name: 'Platform', isVisible: columnVisible(colPrefs, 'Platform') },
		{ index: 4, name: 'Created', isVisible: columnVisible(colPrefs, 'Created') }
	]);

	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const devices = $derived(
		(list.data?.devices ?? []).map((row) => ({ ...row, key: `${row.hwid}::${row.userId}` }))
	);
	const total = $derived(list.data?.total ?? 0);
	const statsRec = $derived((stats.data ?? {}) as Record<string, unknown>);

	syncTableRows(handler, devices, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, devices, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('hwid', handler.rowsPerPage);
	});
	$effect.pre(() => {
		const token = pageChrome.set({
			search: { table, placeholder: 'Search HWID...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('hwid', view) }
		});
		return () => pageChrome.clear(token);
	});

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount + 2, 1));
	const someRowsSelected = $derived(handler.rowCount.selected > 0 && !handler.isAllSelected);
	const selectedKeys = $derived(handler.selected.map(String));

	function ask(spec: {
		title: string;
		description: string;
		confirmLabel?: string;
		variant?: 'destructive' | 'default';
		run: () => Promise<void>;
	}) {
		confirmSpec = {
			title: spec.title,
			description: spec.description,
			confirmLabel: spec.confirmLabel ?? 'Confirm',
			variant: spec.variant ?? 'destructive',
			run: spec.run
		};
		confirmOpen = true;
	}

	async function run(fn: () => Promise<unknown>, ok: string) {
		try {
			await fn();
			appToast.success(ok);
			await list.refetch();
			await stats.refetch();
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	function selectedRows(): HwidRow[] {
		const keys = new Set(selectedKeys);
		return devices.filter((row) => keys.has(row.key));
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && total === 0}
	<EmptyState title="No HWID devices" description="Bound hardware IDs will show up here as users connect." />
{:else}
	<div class="mb-4 grid gap-3 sm:grid-cols-2">
		<div class="rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/40 px-3 py-2.5">
			<p class="text-[11px] text-[var(--color-text-tertiary)]">Devices</p>
			<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{formatNumber(total)}</p>
		</div>
		<div class="rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/40 px-3 py-2.5">
			<p class="text-[11px] text-[var(--color-text-tertiary)]">Stats</p>
			<p class="mt-1 truncate text-sm text-[var(--color-text-secondary)]">
				{typeof statsRec.total === 'number' ? formatNumber(statsRec.total) : Object.keys(statsRec).length ? 'Loaded' : '—'}
			</p>
		</div>
	</div>

	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<ProfilesBulkBar
						count={handler.rowCount.selected}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} devices?`,
								description: selectedRows()
									.slice(0, 5)
									.map((row) => row.hwid)
									.join(', ')
									.concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(async () => {
										for (const row of selectedRows()) {
											await removeM.mutate({ hwid: row.hwid, userId: row.userId });
										}
									}, 'Deleted');
									handler.clearSelection();
								}
							})}
						onclear={() => handler.clearSelection()}
					/>
				</div>
			{/if}
		{/snippet}

		<div
			aria-busy={list.loading}
		>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						<ThLabel class="sticky left-0 z-20 w-10 bg-[var(--color-bg-secondary)]">
							<Checkbox
								aria-label="Select all rows"
								checked={handler.isAllSelected}
								indeterminate={someRowsSelected}
								onCheckedChange={() => handler.selectAll()}
							/>
						</ThLabel>
						{#if view.columns[0]?.isVisible}
							<ThSort {table} field="hwid" class="sticky left-10 z-20 bg-[var(--color-bg-secondary)]">HWID</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field="userId">User</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort {table} field="platform">Platform</ThSort>
						{/if}
						{#if view.columns[3]?.isVisible}
							<ThSort {table} field="createdAt">Created</ThSort>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.key)}
						<Table.Row
							data-state={handler.selected.includes(row.key) ? 'selected' : undefined}
							class="group border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${row.hwid}`}
									checked={handler.selected.includes(row.key)}
									onCheckedChange={() => handler.select(row.key)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 font-mono text-xs group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<span class="block truncate" title={row.hwid}>{row.hwid}</span>
									{#if row.deviceModel}
										<CellSubtitle title={row.deviceModel}>{row.deviceModel}</CellSubtitle>
									{/if}
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{row.userId}</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3">
									<span class="text-[var(--color-text-secondary)]">{row.platform ?? '—'}</span>
									{#if row.osVersion}
										<CellSubtitle>{row.osVersion}</CellSubtitle>
									{/if}
								</Table.Cell>
							{/if}
							{#if view.columns[3]?.isVisible}
								<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{formatDate(row.createdAt)}</Table.Cell>
							{/if}
							<Table.Cell class="w-12 px-2 text-right">
								<DropdownMenu.Root>
									<DropdownMenu.Trigger>
										{#snippet child({ props })}
											<Button {...props} variant="ghost" size="icon-sm" aria-label="Row actions">
												<EllipsisIcon class="size-4" />
											</Button>
										{/snippet}
									</DropdownMenu.Trigger>
									<DropdownMenu.Content align="end" class="w-44">
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: 'Delete this device?',
													description: row.hwid,
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate({ hwid: row.hwid, userId: row.userId }), 'Deleted')
												})}
										>
											Delete
										</DropdownMenu.Item>
									</DropdownMenu.Content>
								</DropdownMenu.Root>
							</Table.Cell>
						</Table.Row>
					{:else}
						{#if list.loading}
							<TableSkeletonRows columns={emptyColspan} />
						{:else}
							<Table.Row class="hover:bg-transparent">
								<Table.Cell colspan={emptyColspan} class="py-10 text-center text-sm text-[var(--color-text-tertiary)]">
									No devices match the current table state.
								</Table.Cell>
							</Table.Row>
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		{#snippet footer()}
			<TableStatusBar {table} loading={list.loading} selection />
		{/snippet}
	</DataTable>
{/if}

<ConfirmAction
	bind:open={confirmOpen}
	title={confirmSpec.title}
	description={confirmSpec.description}
	confirmLabel={confirmSpec.confirmLabel}
	variant={confirmSpec.variant}
	onconfirm={confirmSpec.run}
/>
