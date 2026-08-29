<script lang="ts">
	import { onMount } from 'svelte';
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
	import SquadCreateDialog from '$lib/features/squads/SquadCreateDialog.svelte';
	import SquadDetailSheet from '$lib/features/squads/SquadDetailSheet.svelte';
	import SquadsBulkBar from '$lib/features/squads/SquadsBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { sortByPosition } from '$lib/features/squads/reorder';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { ConfigInbound, InternalSquad } from '@arcanum/ts-client';

	const list = rw.squads.internal.list();
	const inboundsQ = rw.configProfiles.allInbounds();
	const createM = rw.squads.internal.create();
	const updateM = rw.squads.internal.update();
	const removeM = rw.squads.internal.remove();
	const addUsers = rw.squads.internal.addUsers();
	const removeUsers = rw.squads.internal.removeUsers();

	const colPrefs = loadColumnPrefs('internal-squads');
	const handler = new TableHandler<InternalSquad>([], {
		rowsPerPage: loadPageSize('internal-squads'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<InternalSquad>;
	const view = handler.createView([
		{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') },
		{ index: 2, name: 'Members', isVisible: columnVisible(colPrefs, 'Members') },
		{ index: 3, name: 'Inbounds', isVisible: columnVisible(colPrefs, 'Inbounds') }
	]);

	let open = $state(false);
	let selected = $state<InternalSquad | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const squads = $derived(sortByPosition(asArray<InternalSquad>(list.data, ['internalSquads'])));
	const inbounds = $derived(asArray<ConfigInbound>(inboundsQ.data, ['inbounds']));

	syncTableRows(handler, squads, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, squads, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('internal-squads', handler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create squad', onclick: () => (open = true) },
			search: { table, placeholder: 'Search squads...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('internal-squads', view) }
		});
		return () => pageChrome.clear(token);
	});

	onMount(() => listHotkeys({ oncreate: () => (open = true) }));

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
				selected = squads.find((squad) => squad.uuid === selected?.uuid) ?? null;
			}
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	function selectedSquads(): InternalSquad[] {
		const ids = new Set(selectedUuids);
		return squads.filter((squad) => ids.has(squad.uuid));
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) {
			await fn(uuid);
		}
	}

	function openRow(squad: InternalSquad, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		selected = squad;
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && squads.length === 0}
	<EmptyState
		title="No internal squads yet"
		description="Group users by inbounds. Create the first squad to get started."
		action={{ label: 'Create squad', onclick: () => (open = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<SquadsBulkBar
						count={handler.rowCount.selected}
						onaddall={() =>
							ask({
								title: `Add all users to ${handler.rowCount.selected} squads?`,
								description: 'Every user on this panel will be added to the selected squads.',
								confirmLabel: 'Add all',
								variant: 'default',
								run: () =>
									run(
										() => forSelected((uuid) => addUsers.mutate({ uuid })),
										'Add-all queued'
									)
							})}
						onremoveall={() =>
							ask({
								title: `Remove all users from ${handler.rowCount.selected} squads?`,
								description: 'Members of the selected squads will be unassigned.',
								confirmLabel: 'Remove all',
								run: () =>
									run(
										() => forSelected((uuid) => removeUsers.mutate({ uuid })),
										'Remove-all queued'
									)
							})}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} squads?`,
								description: selectedSquads()
									.slice(0, 5)
									.map((squad) => squad.name)
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
							<ThSort {table} field={(row) => row.info?.membersCount ?? 0} class="text-right" btnClass="justify-end">
								Members
							</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort
								{table}
								field={(row) => row.info?.inboundsCount ?? row.inbounds?.length ?? 0}
								class="text-right"
								btnClass="justify-end"
							>
								Inbounds
							</ThSort>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as squad (squad.uuid)}
						<Table.Row
							data-state={handler.selected.includes(squad.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(squad, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${squad.name}`}
									checked={handler.selected.includes(squad.uuid)}
									onCheckedChange={() => handler.select(squad.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button type="button" class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline" onclick={() => (selected = squad)}>
										{squad.name}
									</button>
									{#if squad.inbounds?.length}
										{@const inboundTags = squad.inbounds.map((inbound) => inbound.tag).join(' · ')}
										<CellSubtitle title={inboundTags}>{inboundTags}</CellSubtitle>
									{/if}
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
									{squad.info?.membersCount ?? 0}
								</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
									{squad.info?.inboundsCount ?? squad.inbounds?.length ?? 0}
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
										<DropdownMenu.Item onclick={() => (selected = squad)}>Edit</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={() =>
												ask({
													title: `Add all users to ${squad.name}?`,
													description: 'Every user on this panel will be added to the squad.',
													confirmLabel: 'Add all',
													variant: 'default',
													run: () => run(() => addUsers.mutate({ uuid: squad.uuid }), 'Add-all queued')
												})}
										>
											Add all users
										</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={() =>
												ask({
													title: `Remove all users from ${squad.name}?`,
													description: 'Members of this squad will be unassigned.',
													confirmLabel: 'Remove all',
													run: () => run(() => removeUsers.mutate({ uuid: squad.uuid }), 'Remove-all queued')
												})}
										>
											Remove all users
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${squad.name}?`,
													description: 'The squad will be removed. Users are not deleted.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(squad.uuid), 'Deleted')
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
									No internal squads match the current table state.
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

<SquadCreateDialog
	bind:open
	kind="internal"
	{inbounds}
	pending={createM.pending}
	oncreate={async ({ name, inbounds: inboundUuids }) => {
		await run(() => createM.mutate({ name, inbounds: inboundUuids ?? [] }), 'Created');
	}}
/>

<SquadDetailSheet
	kind="internal"
	bind:squad={selected}
	{inbounds}
	pending={updateM.pending}
	onsave={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onaddall={() => {
		if (!selected) return;
		ask({
			title: `Add all users to ${selected.name}?`,
			description: 'Every user on this panel will be added to the squad.',
			confirmLabel: 'Add all',
			variant: 'default',
			run: () => run(() => addUsers.mutate({ uuid: selected!.uuid }), 'Add-all queued')
		});
	}}
	onremoveall={() => {
		if (!selected) return;
		ask({
			title: `Remove all users from ${selected.name}?`,
			description: 'Members of this squad will be unassigned.',
			confirmLabel: 'Remove all',
			run: () => run(() => removeUsers.mutate({ uuid: selected!.uuid }), 'Remove-all queued')
		});
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'The squad will be removed. Users are not deleted.',
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
