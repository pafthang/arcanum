<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
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
	import StatusBadge from '$lib/components/remnawave/StatusBadge.svelte';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import HostCreateDialog from '$lib/features/hosts/HostCreateDialog.svelte';
	import HostDetailSheet from '$lib/features/hosts/HostDetailSheet.svelte';
	import HostsBulkBar from '$lib/features/hosts/HostsBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { ConfigInbound, ConfigProfile, Host } from '@arcanum/ts-client';

	const list = rw.hosts.list();
	const profiles = rw.configProfiles.list();
	const allInbounds = rw.configProfiles.allInbounds();
	const createM = rw.hosts.create();
	const updateM = rw.hosts.update();
	const removeM = rw.hosts.remove();
	const bulkEnable = rw.hosts.bulkEnable();
	const bulkDisable = rw.hosts.bulkDisable();
	const bulkDelete = rw.hosts.bulkDelete();

	const colPrefs = loadColumnPrefs('hosts');
	const handler = new TableHandler<Host>([], {
		rowsPerPage: loadPageSize('hosts'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<Host>;
	const view = handler.createView([
		{ index: 1, name: 'Remark', isVisible: columnVisible(colPrefs, 'Remark') },
		{ index: 2, name: 'Status', isVisible: columnVisible(colPrefs, 'Status') },
		{ index: 3, name: 'Tags', isVisible: columnVisible(colPrefs, 'Tags') }
	]);

	let openCreate = $state(false);
	let selected = $state<Host | null>(null);
	let statusFilter = $state('all');
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const hosts = $derived(asArray<Host>(list.data, ['hosts']));
	const inbounds = $derived.by(() => {
		const fromAll = asArray<ConfigInbound>(allInbounds.data, ['inbounds']);
		if (fromAll.length) return fromAll;
		return asArray<ConfigProfile>(profiles.data, ['configProfiles']).flatMap((item) => item.inbounds ?? []);
	});
	const visible = $derived.by(() => {
		if (statusFilter === 'disabled') return hosts.filter((host) => host.isDisabled);
		if (statusFilter === 'active') return hosts.filter((host) => !host.isDisabled);
		return hosts;
	});

	syncTableRows(handler, visible, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, visible, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('hosts', handler.rowsPerPage);
	});

	const statusOptions = [
		{ value: 'all', label: 'All' },
		{ value: 'active', label: 'Active' },
		{ value: 'disabled', label: 'Disabled' }
	];

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create host', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search hosts...' },
			toolbar: {
				table,
				view,
				onviewchange: () => saveColumnPrefs('hosts', view),
				filter: {
					value: () => statusFilter,
					options: statusOptions,
					onselect: (value) => (statusFilter = value)
				}
			}
		});
		return () => pageChrome.clear(token);
	});

	onMount(() => listHotkeys({ oncreate: () => (openCreate = true) }));

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount + 2, 1));
	const someRowsSelected = $derived(handler.rowCount.selected > 0 && !handler.isAllSelected);
	const selectedUuids = $derived(handler.selected.map(String));

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
			if (selected) selected = hosts.find((host) => host.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	function openHost(host: Host) {
		selected = host;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== host.uuid) {
			url.searchParams.set('id', host.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(host: Host, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openHost(host);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = hosts.find((host) => host.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
		}
	});
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && hosts.length === 0}
	<EmptyState
		title="No hosts yet"
		description="Create the first subscription host to publish endpoints."
		action={{ label: 'Create host', onclick: () => (openCreate = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<HostsBulkBar
						count={handler.rowCount.selected}
						onenable={() => void run(() => bulkEnable.mutate({ uuids: [...selectedUuids] }), 'Enabled')}
						ondisable={() =>
							ask({
								title: `Disable ${handler.rowCount.selected} hosts?`,
								description: 'Disabled hosts will not appear in subscriptions.',
								confirmLabel: 'Disable',
								run: () => run(() => bulkDisable.mutate({ uuids: [...selectedUuids] }), 'Disabled')
							})}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} hosts?`,
								description: hosts
									.filter((host) => selectedUuids.includes(host.uuid))
									.slice(0, 5)
									.map((host) => host.remark)
									.join(', ')
									.concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(() => bulkDelete.mutate({ uuids: [...selectedUuids] }), 'Deleted');
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
							<ThSort {table} field="remark" class="sticky left-10 z-20 bg-[var(--color-bg-secondary)]">Remark</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field={(row) => row.isDisabled}>Status</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThLabel>Tags</ThLabel>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as host (host.uuid)}
						{@const endpoint = `${host.address}:${host.port}`}
						{@const tags = host.tags?.join(' · ') || '—'}
						<Table.Row
							data-state={handler.selected.includes(host.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(host, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${host.remark}`}
									checked={handler.selected.includes(host.uuid)}
									onCheckedChange={() => handler.select(host.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openHost(host)}
									>
										{host.remark}
									</button>
									<CellSubtitle title={endpoint}>{endpoint}</CellSubtitle>
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3">
									<StatusBadge value={host.isDisabled ? 'DISABLED' : 'ACTIVE'} />
								</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="max-w-[12rem] px-3">
									<span class="block truncate text-[var(--color-text-secondary)]" title={tags}>{tags}</span>
								</Table.Cell>
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
										<DropdownMenu.Item onclick={() => openHost(host)}>Open</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${host.remark}?`,
													description: 'This host will disappear from subscriptions.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(host.uuid), 'Deleted')
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
									No hosts match the current table state.
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

<HostCreateDialog
	bind:open={openCreate}
	{inbounds}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'Host created');
	}}
/>

<HostDetailSheet
	bind:host={selected}
	{inbounds}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onenable={() => {
		if (selected) void run(() => bulkEnable.mutate({ uuids: [selected.uuid] }), 'Enabled');
	}}
	ondisable={() => {
		if (!selected) return;
		ask({
			title: `Disable ${selected.remark}?`,
			description: 'This host will not appear in subscriptions.',
			confirmLabel: 'Disable',
			run: () => run(() => bulkDisable.mutate({ uuids: [selected!.uuid] }), 'Disabled')
		});
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.remark}?`,
			description: 'This host will disappear from subscriptions.',
			confirmLabel: 'Delete',
			run: async () => {
				await run(() => removeM.mutate(selected!.uuid), 'Deleted');
				selected = null;
			}
		});
	}}
/>

<ConfirmAction
	bind:open={confirmOpen}
	title={confirmSpec.title}
	description={confirmSpec.description}
	confirmLabel={confirmSpec.confirmLabel}
	variant={confirmSpec.variant}
	onconfirm={confirmSpec.run}
/>
