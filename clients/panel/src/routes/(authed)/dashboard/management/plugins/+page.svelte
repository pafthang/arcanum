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
	import PluginCreateDialog from '$lib/features/plugins/PluginCreateDialog.svelte';
	import PluginDetailSheet from '$lib/features/plugins/PluginDetailSheet.svelte';
	import PluginsBulkBar from '$lib/features/plugins/PluginsBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { NodePlugin } from '@optimawave/ts-back';

	const list = rw.plugins.list();
	const createM = rw.plugins.create();
	const updateM = rw.plugins.update();
	const removeM = rw.plugins.remove();
	const cloneM = rw.plugins.clone();
	const syncM = rw.plugins.sync();

	const colPrefs = loadColumnPrefs('node-plugins');
	const handler = new TableHandler<NodePlugin>([], {
		rowsPerPage: loadPageSize('node-plugins'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<NodePlugin>;
	const view = handler.createView([{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') }]);

	let openCreate = $state(false);
	let selected = $state<NodePlugin | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const plugins = $derived(asArray<NodePlugin>(list.data, ['nodePlugins']));

	syncTableRows(handler, plugins, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, plugins, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('node-plugins', handler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create plugin', onclick: () => (openCreate = true) },
			action: {
				label: 'Sync',
				icon: 'sync',
				onclick: () => run(() => syncM.mutate({}), 'Sync queued'),
				pending: () => syncM.pending
			},
			search: { table, placeholder: 'Search plugins...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('node-plugins', view) }
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
			if (selected) selected = plugins.find((plugin) => plugin.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) await fn(uuid);
	}

	function openPlugin(plugin: NodePlugin) {
		selected = plugin;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== plugin.uuid) {
			url.searchParams.set('id', plugin.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(plugin: NodePlugin, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openPlugin(plugin);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = plugins.find((plugin) => plugin.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
		}
	});

	function configHint(plugin: NodePlugin): string {
		if (plugin.pluginConfig && typeof plugin.pluginConfig === 'object' && !Array.isArray(plugin.pluginConfig)) {
			const keys = Object.keys(plugin.pluginConfig as object);
			if (keys.length) return keys.slice(0, 6).join(' · ');
		}
		return 'Plugin config';
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && plugins.length === 0}
	<EmptyState
		title="No plugins yet"
		description="Create a plugin config, then sync it to connected nodes."
		action={{ label: 'Create plugin', onclick: () => (openCreate = true) }}
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
								title: `Delete ${handler.rowCount.selected} plugins?`,
								description: plugins
									.filter((plugin) => selectedUuids.includes(plugin.uuid))
									.slice(0, 5)
									.map((plugin) => plugin.name)
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
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as plugin (plugin.uuid)}
						{@const hint = configHint(plugin)}
						<Table.Row
							data-state={handler.selected.includes(plugin.uuid) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(plugin, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${plugin.name}`}
									checked={handler.selected.includes(plugin.uuid)}
									onCheckedChange={() => handler.select(plugin.uuid)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[24rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openPlugin(plugin)}
									>
										{plugin.name}
									</button>
									<CellSubtitle title={hint}>{hint}</CellSubtitle>
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
										<DropdownMenu.Item onclick={() => openPlugin(plugin)}>Open</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={() => run(() => cloneM.mutate({ cloneFromUuid: plugin.uuid }), 'Cloned')}
										>
											Clone
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${plugin.name}?`,
													description: 'The plugin config will be removed from this panel.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(plugin.uuid), 'Deleted')
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
									No plugins match the current table state.
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

<PluginCreateDialog
	bind:open={openCreate}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'Created');
	}}
/>

<PluginDetailSheet
	bind:plugin={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onclone={() => {
		if (selected) void run(() => cloneM.mutate({ cloneFromUuid: selected.uuid }), 'Cloned');
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'The plugin config will be removed from this panel.',
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
