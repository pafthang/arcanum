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
	import TrafficBar from '$lib/components/remnawave/TrafficBar.svelte';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import NodeCreateDialog from '$lib/features/nodes/NodeCreateDialog.svelte';
	import NodeDetailSheet from '$lib/features/nodes/NodeDetailSheet.svelte';
	import NodesBulkBar from '$lib/features/nodes/NodesBulkBar.svelte';
	import { nodeEndpoint, nodeProfile, nodeStatus } from '$lib/features/nodes/status';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { ConfigProfile, Node } from '@arcanum/ts-client';

	const list = rw.nodes.list();
	const profilesQ = rw.configProfiles.list();
	const createM = rw.nodes.create();
	const updateM = rw.nodes.update();
	const removeM = rw.nodes.remove();
	const enableM = rw.nodes.enable();
	const disableM = rw.nodes.disable();
	const restartM = rw.nodes.restart();
	const resetM = rw.nodes.resetTraffic();
	const restartAllM = rw.nodes.restartAll();
	const bulkActions = rw.nodes.bulkActions();

	const colPrefs = loadColumnPrefs('nodes');
	const handler = new TableHandler<Node>([], {
		rowsPerPage: loadPageSize('nodes'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<Node>;
	const view = handler.createView([
		{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') },
		{ index: 2, name: 'Status', isVisible: columnVisible(colPrefs, 'Status') },
		{ index: 3, name: 'Online', isVisible: columnVisible(colPrefs, 'Online') },
		{ index: 4, name: 'Traffic', isVisible: columnVisible(colPrefs, 'Traffic') },
		{ index: 5, name: 'Profile', isVisible: columnVisible(colPrefs, 'Profile') }
	]);

	let openCreate = $state(false);
	let selected = $state<Node | null>(null);
	let statusFilter = $state('all');
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const nodes = $derived(asArray<Node>(list.data, ['nodes']));
	const profiles = $derived(asArray<ConfigProfile>(profilesQ.data, ['configProfiles']));
	const profileName = $derived.by(() => {
		const map = new Map(profiles.map((profile) => [profile.uuid, profile.name]));
		return (uuid?: string) => (uuid ? (map.get(uuid) ?? uuid.slice(0, 8)) : '—');
	});
	const visible = $derived.by(() => {
		if (statusFilter === 'all') return nodes;
		return nodes.filter((node) => nodeStatus(node) === statusFilter);
	});

	syncTableRows(handler, visible, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, visible, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('nodes', handler.rowsPerPage);
	});

	const statusOptions = [
		{ value: 'all', label: 'All' },
		{ value: 'CONNECTED', label: 'Connected' },
		{ value: 'DISABLED', label: 'Disabled' },
		{ value: 'OFFLINE', label: 'Offline' }
	];

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create node', onclick: () => (openCreate = true) },
			action: {
				label: 'Restart all',
				icon: 'restart',
				onclick: () =>
					ask({
						title: 'Restart all nodes?',
						description: 'Every enabled node will be restarted.',
						confirmLabel: 'Restart all',
						variant: 'default',
						run: () => run(() => restartAllM.mutate({ forceRestart: false }), 'Restart queued')
					}),
				pending: () => restartAllM.pending
			},
			search: { table, placeholder: 'Search nodes...' },
			toolbar: {
				table,
				view,
				onviewchange: () => saveColumnPrefs('nodes', view),
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
			if (selected) {
				selected = nodes.find((node) => node.uuid === selected?.uuid) ?? selected;
			}
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) await fn(uuid);
	}

	function bulk(action: string, ok: string) {
		return run(() => bulkActions.mutate({ uuids: [...selectedUuids], action }), ok);
	}

	function openNode(node: Node) {
		selected = node;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== node.uuid) {
			url.searchParams.set('id', node.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(node: Node, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openNode(node);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = nodes.find((node) => node.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
		}
	});
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && nodes.length === 0}
	<EmptyState
		title="No nodes yet"
		description="Add the first node and attach a config profile to start traffic."
		action={{ label: 'Create node', onclick: () => (openCreate = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<NodesBulkBar
						count={handler.rowCount.selected}
						onenable={() => void bulk('ENABLE', 'Enabled')}
						ondisable={() =>
							ask({
								title: `Disable ${handler.rowCount.selected} nodes?`,
								description: 'Disabled nodes will stop accepting connections.',
								confirmLabel: 'Disable',
								run: () => bulk('DISABLE', 'Disabled')
							})}
						onrestart={() => void bulk('RESTART', 'Restart queued')}
						onreset={() =>
							ask({
								title: `Reset traffic for ${handler.rowCount.selected} nodes?`,
								description: 'Used traffic counters will be set to zero.',
								confirmLabel: 'Reset',
								run: () => bulk('RESET_TRAFFIC', 'Traffic reset')
							})}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} nodes?`,
								description: nodes
									.filter((node) => selectedUuids.includes(node.uuid))
									.slice(0, 5)
									.map((node) => node.name)
									.join(', ')
									.concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(() => forSelected((uuid) => removeM.mutate(uuid)), 'Deleted');
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
							<ThSort {table} field="name" class="sticky left-10 z-20 bg-[var(--color-bg-secondary)]">Name</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field={(row) => nodeStatus(row)}>Status</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort {table} field={(row) => row.usersOnline ?? 0} class="text-right" btnClass="justify-end">
								Online
							</ThSort>
						{/if}
						{#if view.columns[3]?.isVisible}
							<ThSort {table} field={(row) => row.trafficUsedBytes ?? 0} class="text-right" btnClass="justify-end">
								Traffic
							</ThSort>
						{/if}
						{#if view.columns[4]?.isVisible}
							<ThLabel>Profile</ThLabel>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as node (node.uuid)}
						{@const endpoint = nodeEndpoint(node)}
						{@const profile = nodeProfile(node)}
						<Table.Row
							data-state={handler.selected.includes(node.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(node, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${node.name}`}
									checked={handler.selected.includes(node.uuid)}
									onCheckedChange={() => handler.select(node.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openNode(node)}
									>
										{node.name}
									</button>
									<CellSubtitle title={endpoint}>{endpoint}</CellSubtitle>
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3"><StatusBadge value={nodeStatus(node)} /></Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
									{node.usersOnline ?? 0}
								</Table.Cell>
							{/if}
							{#if view.columns[3]?.isVisible}
								<Table.Cell class="px-3">
									<TrafficBar compact used={node.trafficUsedBytes} limit={node.trafficLimitBytes} />
								</Table.Cell>
							{/if}
							{#if view.columns[4]?.isVisible}
								<Table.Cell class="max-w-[12rem] px-3">
									<span class="block truncate text-[var(--color-text-secondary)]">{profileName(profile.uuid)}</span>
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
										<DropdownMenu.Item onclick={() => openNode(node)}>Open</DropdownMenu.Item>
										{#if node.isDisabled}
											<DropdownMenu.Item onclick={() => run(() => enableM.mutate(node.uuid), 'Enabled')}>
												Enable
											</DropdownMenu.Item>
										{:else}
											<DropdownMenu.Item
												onclick={() =>
													ask({
														title: `Disable ${node.name}?`,
														description: 'This node will stop accepting connections.',
														confirmLabel: 'Disable',
														run: () => run(() => disableM.mutate(node.uuid), 'Disabled')
													})}
											>
												Disable
											</DropdownMenu.Item>
											<DropdownMenu.Item
												onclick={() =>
													run(() => restartM.mutate({ uuid: node.uuid, forceRestart: true }), 'Restart queued')}
											>
												Restart
											</DropdownMenu.Item>
										{/if}
										<DropdownMenu.Item
											onclick={() =>
												ask({
													title: `Reset traffic for ${node.name}?`,
													description: 'Used traffic will be set to zero.',
													confirmLabel: 'Reset',
													run: () => run(() => resetM.mutate(node.uuid), 'Traffic reset')
												})}
										>
											Reset traffic
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${node.name}?`,
													description: 'The node will be removed. Users are not deleted.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(node.uuid), 'Deleted')
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
									No nodes match the current table state.
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

<NodeCreateDialog
	bind:open={openCreate}
	{profiles}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'Node created');
	}}
/>

<NodeDetailSheet
	bind:node={selected}
	{profiles}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onenable={() => {
		if (selected) void run(() => enableM.mutate(selected.uuid), 'Enabled');
	}}
	ondisable={() => {
		if (!selected) return;
		ask({
			title: `Disable ${selected.name}?`,
			description: 'This node will stop accepting connections.',
			confirmLabel: 'Disable',
			run: () => run(() => disableM.mutate(selected!.uuid), 'Disabled')
		});
	}}
	onrestart={() => {
		if (selected) void run(() => restartM.mutate({ uuid: selected.uuid, forceRestart: true }), 'Restart queued');
	}}
	onreset={() => {
		if (!selected) return;
		ask({
			title: `Reset traffic for ${selected.name}?`,
			description: 'Used traffic will be set to zero.',
			confirmLabel: 'Reset',
			run: () => run(() => resetM.mutate(selected!.uuid), 'Traffic reset')
		});
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'The node will be removed. Users are not deleted.',
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
