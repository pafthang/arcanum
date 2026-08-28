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
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import ProfileCreateDialog from '$lib/features/profiles/ProfileCreateDialog.svelte';
	import ProfileDetailSheet from '$lib/features/profiles/ProfileDetailSheet.svelte';
	import ProfilesBulkBar from '$lib/features/profiles/ProfilesBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { ConfigProfile } from '@optimawave/ts-back';

	const list = rw.configProfiles.list();
	const createM = rw.configProfiles.create();
	const updateM = rw.configProfiles.update();
	const removeM = rw.configProfiles.remove();

	const colPrefs = loadColumnPrefs('config-profiles');
	const handler = new TableHandler<ConfigProfile>([], {
		rowsPerPage: loadPageSize('config-profiles'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<ConfigProfile>;
	const view = handler.createView([
		{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') },
		{ index: 2, name: 'Inbounds', isVisible: columnVisible(colPrefs, 'Inbounds') },
		{ index: 3, name: 'Nodes', isVisible: columnVisible(colPrefs, 'Nodes') }
	]);

	let openCreate = $state(false);
	let selected = $state<ConfigProfile | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const profiles = $derived(asArray<ConfigProfile>(list.data, ['configProfiles']));

	syncTableRows(handler, profiles, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, profiles, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('config-profiles', handler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create profile', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search profiles...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('config-profiles', view) }
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
			if (selected) selected = profiles.find((profile) => profile.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) await fn(uuid);
	}

	function openProfile(profile: ConfigProfile) {
		selected = profile;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== profile.uuid) {
			url.searchParams.set('id', profile.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(profile: ConfigProfile, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openProfile(profile);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	const peeked = rw.configProfiles.byUuid(() => openId);
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = profiles.find((profile) => profile.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
			return;
		}
		if (peeked.data && peeked.data.uuid === openId) {
			selected = peeked.data;
			openedId = openId;
		}
	});
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && profiles.length === 0}
	<EmptyState
		title="No config profiles yet"
		description="Create an Xray profile. Inbounds are extracted after save."
		action={{ label: 'Create profile', onclick: () => (openCreate = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<ProfilesBulkBar
						count={handler.rowCount.selected}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} profiles?`,
								description: profiles
									.filter((profile) => selectedUuids.includes(profile.uuid))
									.slice(0, 5)
									.map((profile) => profile.name)
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
							<ThSort
								{table}
								field={(row) => row.inbounds?.length ?? 0}
								class="text-right"
								btnClass="justify-end"
							>
								Inbounds
							</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort {table} field={(row) => row.nodes?.length ?? 0} class="text-right" btnClass="justify-end">
								Nodes
							</ThSort>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as profile (profile.uuid)}
						{@const inboundTags = (profile.inbounds ?? []).map((inbound) => inbound.tag).join(' · ')}
						<Table.Row
							data-state={handler.selected.includes(profile.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(profile, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${profile.name}`}
									checked={handler.selected.includes(profile.uuid)}
									onCheckedChange={() => handler.select(profile.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openProfile(profile)}
									>
										{profile.name}
									</button>
									{#if inboundTags}
										<CellSubtitle title={inboundTags}>{inboundTags}</CellSubtitle>
									{/if}
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
									{profile.inbounds?.length ?? 0}
								</Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
									{profile.nodes?.length ?? 0}
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
										<DropdownMenu.Item onclick={() => openProfile(profile)}>Open</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${profile.name}?`,
													description: 'Nodes using this profile will need another profile assigned.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(profile.uuid), 'Deleted')
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
									No profiles match the current table state.
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

<ProfileCreateDialog
	bind:open={openCreate}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'Created');
	}}
/>

<ProfileDetailSheet
	bind:profile={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'Nodes using this profile will need another profile assigned.',
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
