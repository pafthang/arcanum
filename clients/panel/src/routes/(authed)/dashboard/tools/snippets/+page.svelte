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
	import SnippetCreateDialog from '$lib/features/snippets/SnippetCreateDialog.svelte';
	import SnippetDetailSheet from '$lib/features/snippets/SnippetDetailSheet.svelte';
	import ProfilesBulkBar from '$lib/features/profiles/ProfilesBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { Snippet } from '@optimawave/ts-back';

	const list = rw.snippets.list();
	const createM = rw.snippets.create();
	const updateM = rw.snippets.update();
	const removeM = rw.snippets.remove();
	const syncM = rw.snippets.sync();

	const colPrefs = loadColumnPrefs('snippets');
	const handler = new TableHandler<Snippet>([], {
		rowsPerPage: loadPageSize('snippets'),
		selectBy: 'name'
	});
	const table = handler as TableHandlerInterface<Snippet>;
	const view = handler.createView([{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') }]);

	let openCreate = $state(false);
	let selected = $state<Snippet | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const snippets = $derived(asArray<Snippet>(list.data, ['snippets']));

	syncTableRows(handler, snippets, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, snippets, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('snippets', handler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create snippet', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search snippets...' },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('snippets', view) }
		});
		return () => pageChrome.clear(token);
	});

	onMount(() => listHotkeys({ oncreate: () => (openCreate = true) }));

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount + 2, 1));
	const someRowsSelected = $derived(handler.rowCount.selected > 0 && !handler.isAllSelected);
	const selectedNames = $derived(handler.selected.map(String));

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
			if (selected) selected = snippets.find((row) => row.name === selected?.name) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (name: string) => Promise<unknown>) {
		const names = [...selectedNames];
		for (const name of names) await fn(name);
	}

	function openRow(row: Snippet, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		selected = row;
	}

	function hint(row: Snippet): string {
		if (Array.isArray(row.snippet)) return `${row.snippet.length} items`;
		if (row.snippet && typeof row.snippet === 'object') {
			const keys = Object.keys(row.snippet as object);
			if (keys.length) return keys.slice(0, 6).join(' · ');
		}
		return 'JSON snippet';
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && snippets.length === 0}
	<EmptyState
		title="No snippets yet"
		description="Create a named JSON fragment to reuse in config profiles."
		action={{ label: 'Create snippet', onclick: () => (openCreate = true) }}
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
								title: `Delete ${handler.rowCount.selected} snippets?`,
								description: selectedNames.slice(0, 5).join(', ').concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(() => forSelected((name) => removeM.mutate({ name })), 'Deleted');
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
					{#each handler.rows as row (row.name)}
						{@const subtitle = hint(row)}
						<Table.Row
							data-state={handler.selected.includes(row.name) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(row, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${row.name}`}
									checked={handler.selected.includes(row.name)}
									onCheckedChange={() => handler.select(row.name)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[24rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => (selected = row)}
									>
										{row.name}
									</button>
									<CellSubtitle title={subtitle}>{subtitle}</CellSubtitle>
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
										<DropdownMenu.Item onclick={() => (selected = row)}>Open</DropdownMenu.Item>
										<DropdownMenu.Item onclick={() => run(() => syncM.mutate({ name: row.name }), 'Synced')}>
											Sync
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${row.name}?`,
													description: 'Profiles using this snippet will need another fragment.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate({ name: row.name }), 'Deleted')
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
									No snippets match the current table state.
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

<SnippetCreateDialog
	bind:open={openCreate}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'Created');
	}}
/>

<SnippetDetailSheet
	bind:snippet={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onsync={() => {
		if (selected) void run(() => syncM.mutate({ name: selected.name }), 'Synced');
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'Profiles using this snippet will need another fragment.',
			confirmLabel: 'Delete',
			run: async () => {
				await run(() => removeM.mutate({ name: selected!.name }), 'Deleted');
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
