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
	import TemplateCreateDialog from '$lib/features/templates/TemplateCreateDialog.svelte';
	import TemplateDetailSheet, { type TemplateRecord } from '$lib/features/templates/TemplateDetailSheet.svelte';
	import ProfilesBulkBar from '$lib/features/profiles/ProfilesBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { TEMPLATE_TYPES } from '$lib/nav';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	const type = $derived(page.params.type ?? 'XRAY_JSON');
	const typeLabel = $derived(TEMPLATE_TYPES.find((item) => item.type === type)?.label ?? type);

	const list = rw.subscription.templates.list();
	const createM = rw.subscription.templates.create();
	const updateM = rw.subscription.templates.update();
	const removeM = rw.subscription.templates.remove();

	const colPrefs = loadColumnPrefs('templates');
	const handler = new TableHandler<TemplateRecord>([], {
		rowsPerPage: loadPageSize('templates'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<TemplateRecord>;
	const view = handler.createView([
		{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') },
		{ index: 2, name: 'Type', isVisible: columnVisible(colPrefs, 'Type') }
	]);

	let openCreate = $state(false);
	let selected = $state<TemplateRecord | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const all = $derived(asArray<TemplateRecord>(list.data, ['templates', 'subscriptionTemplates']));
	const rows = $derived(all.filter((row) => !row.templateType || row.templateType === type));

	syncTableRows(handler, rows, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, rows, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		savePageSize('templates', handler.rowsPerPage);
	});

	const typeOptions = TEMPLATE_TYPES.map((item) => ({ value: item.type, label: item.label }));

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create template', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search templates...' },
			toolbar: {
				table,
				view,
				onviewchange: () => saveColumnPrefs('templates', view),
				filter: {
					value: () => type,
					options: typeOptions,
					onselect: (value) => goto(`/dashboard/templates/${value}`)
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
			if (selected) selected = all.find((row) => row.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>) {
		const uuids = [...selectedUuids];
		for (const uuid of uuids) await fn(uuid);
	}

	function openTemplate(row: TemplateRecord) {
		selected = row;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== row.uuid) {
			url.searchParams.set('id', row.uuid);
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(row: TemplateRecord, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openTemplate(row);
	}

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	const peeked = rw.subscription.templates.byUuid(() => selected?.uuid ?? openId);
	let openedId = $state('');
	$effect(() => {
		if (!openId || openedId === openId) return;
		const found = all.find((row) => row.uuid === openId);
		if (found) {
			selected = found;
			openedId = openId;
		}
	});
	$effect(() => {
		const data = peeked.data as TemplateRecord | undefined;
		if (!data || !selected || data.uuid !== selected.uuid) return;
		if (selected.templateJson !== undefined || selected.encodedTemplateYaml != null) return;
		selected = { ...selected, ...data };
	});
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && rows.length === 0 && all.length === 0}
	<EmptyState
		title="No templates yet"
		description="Create a client subscription template for {typeLabel}."
		action={{ label: 'Create template', onclick: () => (openCreate = true) }}
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
								title: `Delete ${handler.rowCount.selected} templates?`,
								description: rows
									.filter((row) => selectedUuids.includes(row.uuid))
									.slice(0, 5)
									.map((row) => row.name)
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
							<ThSort {table} field="templateType">Type</ThSort>
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
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openTemplate(row)}
									>
										{row.name}
									</button>
									<CellSubtitle>{row.templateType}</CellSubtitle>
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{row.templateType}</Table.Cell>
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
										<DropdownMenu.Item onclick={() => openTemplate(row)}>Open</DropdownMenu.Item>
										{#if row.name !== 'Default'}
											<DropdownMenu.Separator />
											<DropdownMenu.Item
												variant="destructive"
												onclick={() =>
													ask({
														title: `Delete ${row.name}?`,
														description: 'This template will no longer be served to clients.',
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
									No {typeLabel} templates.
									<button
										type="button"
										class="ml-1 text-[var(--app-accent-light)] hover:underline"
										onclick={() => (openCreate = true)}
									>
										Create one
									</button>
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

<TemplateCreateDialog
	bind:open={openCreate}
	{typeLabel}
	pending={createM.pending}
	oncreate={async (name) => {
		await run(() => createM.mutate({ name, templateType: type }), 'Created');
	}}
/>

<TemplateDetailSheet
	bind:template={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	ondelete={() => {
		if (!selected || selected.name === 'Default') return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'This template will no longer be served to clients.',
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
