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
	import SubpageCreateDialog from '$lib/features/subpage/SubpageCreateDialog.svelte';
	import SubpageDetailSheet, { type PageConfig } from '$lib/features/subpage/SubpageDetailSheet.svelte';
	import PluginsBulkBar from '$lib/features/plugins/PluginsBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	const reserved = '00000000-0000-0000-0000-000000000000';
	const list = rw.subscription.pageConfigs.list();
	const createM = rw.subscription.pageConfigs.create();
	const updateM = rw.subscription.pageConfigs.update();
	const removeM = rw.subscription.pageConfigs.remove();
	const cloneM = rw.subscription.pageConfigs.clone();

	const colPrefs = loadColumnPrefs('subpage-configs');
	const handler = new TableHandler<PageConfig>([], {
		rowsPerPage: loadPageSize('subpage-configs'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<PageConfig>;
	const view = handler.createView([{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') }]);

	let openCreate = $state(false);
	let selected = $state<PageConfig | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const configs = $derived(asArray<PageConfig>(list.data, ['configs', 'subscriptionPageConfigs']));

	syncTableRows(handler, configs, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, configs, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('subpage-configs', handler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create page config', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search configs...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('subpage-configs', view) }
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
			if (selected) selected = configs.find((row) => row.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) await fn(uuid);
	}

	function openConfig(row: PageConfig) {
		selected = row;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== row.uuid) {
			url.searchParams.set('id', row.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(row: PageConfig, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openConfig(row);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	const peeked = rw.subscription.pageConfigs.byUuid(() => selected?.uuid ?? openId);
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = configs.find((row) => row.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
		}
	});
	$effect(() => {
		const data = peeked.data as PageConfig | undefined;
		if (!data || !selected || data.uuid !== selected.uuid) return;
		if (selected.config !== undefined) return;
		selected = { ...selected, ...data };
	});
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && configs.length === 0}
	<EmptyState
		title="No page configs yet"
		description="Create a hosted subscription landing config."
		action={{ label: 'Create page config', onclick: () => (openCreate = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<PluginsBulkBar
						count={handler.rowCount.selected}
						onclone={() =>
							void run(
								() => forSelected((uuid) => cloneM.mutate({ cloneFromUuid: uuid })),
								'Cloned'
							)}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} configs?`,
								description: configs
									.filter((row) => selectedUuids.includes(row.uuid) && row.uuid !== reserved)
									.slice(0, 5)
									.map((row) => row.name)
									.join(', ')
									.concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(
										() =>
											forSelected((uuid) =>
												uuid === reserved ? Promise.resolve() : removeM.mutate(uuid)
											),
										'Deleted'
									);
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
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.uuid)}
						<Table.Row
							data-state={handler.selected.includes(row.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(row, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${row.name}`}
									checked={handler.selected.includes(row.uuid)}
									onCheckedChange={() => handler.select(row.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[24rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openConfig(row)}
									>
										{row.name}
									</button>
									<CellSubtitle>
										{row.uuid === reserved ? 'Reserved default' : 'Landing config'}
									</CellSubtitle>
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
										<DropdownMenu.Item onclick={() => openConfig(row)}>Open</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={() => run(() => cloneM.mutate({ cloneFromUuid: row.uuid }), 'Cloned')}
										>
											Clone
										</DropdownMenu.Item>
										{#if row.uuid !== reserved}
											<DropdownMenu.Separator />
											<DropdownMenu.Item
												variant="destructive"
												onclick={() =>
													ask({
														title: `Delete ${row.name}?`,
														description: 'The hosted landing config will be removed.',
														confirmLabel: 'Delete',
														run: () => run(() => removeM.mutate(row.uuid), 'Deleted')
													})}
											>
												Delete
											</DropdownMenu.Item>
										{/if}
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
									No page configs match the current table state.
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

<SubpageCreateDialog
	bind:open={openCreate}
	pending={createM.pending}
	oncreate={async (name) => {
		await run(() => createM.mutate({ name }), 'Created');
	}}
/>

<SubpageDetailSheet
	bind:config={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onclone={() => {
		if (selected) void run(() => cloneM.mutate({ cloneFromUuid: selected.uuid }), 'Cloned');
	}}
	ondelete={() => {
		if (!selected || selected.uuid === reserved) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'The hosted landing config will be removed.',
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
