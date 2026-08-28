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
	import ProviderCreateDialog from '$lib/features/infra/ProviderCreateDialog.svelte';
	import ProviderDetailSheet, { type Provider } from '$lib/features/infra/ProviderDetailSheet.svelte';
	import ProfilesBulkBar from '$lib/features/profiles/ProfilesBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { formatDate } from '$lib/format';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type BillingNode = {
		uuid: string;
		name?: string | null;
		providerUuid: string;
		provider?: { name?: string };
		node?: { name?: string; countryCode?: string } | null;
		nextBillingAt?: string;
	};
	type HistoryRow = {
		uuid: string;
		amount: number;
		billedAt?: string;
		provider?: { name?: string };
	};

	const providersQ = rw.infra.providers.list();
	const nodesQ = rw.infra.nodes.list();
	const historyQ = rw.infra.history.list();
	const createP = rw.infra.providers.create();
	const updateP = rw.infra.providers.update();
	const removeP = rw.infra.providers.remove();
	const removeNode = rw.infra.nodes.remove();
	const removeHistory = rw.infra.history.remove();

	const colPrefs = loadColumnPrefs('infra-providers');
	const handler = new TableHandler<Provider>([], {
		rowsPerPage: loadPageSize('infra-providers'),
		selectBy: 'uuid'
	});
	const table = handler as TableHandlerInterface<Provider>;
	const view = handler.createView([
		{ index: 1, name: 'Name', isVisible: columnVisible(colPrefs, 'Name') },
		{ index: 2, name: 'Bills', isVisible: columnVisible(colPrefs, 'Bills') },
		{ index: 3, name: 'Spent', isVisible: columnVisible(colPrefs, 'Spent') }
	]);

	const nodeHandler = new TableHandler<BillingNode>([], {
		rowsPerPage: loadPageSize('infra-nodes', 10),
		selectBy: 'uuid'
	});
	const nodeTable = nodeHandler as TableHandlerInterface<BillingNode>;

	const historyHandler = new TableHandler<HistoryRow>([], {
		rowsPerPage: loadPageSize('infra-history', 10),
		selectBy: 'uuid'
	});
	const historyTable = historyHandler as TableHandlerInterface<HistoryRow>;

	let openCreate = $state(false);
	let selected = $state<Provider | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const providers = $derived(asArray<Provider>(providersQ.data, ['providers']));
	const billed = $derived(asArray<BillingNode>(nodesQ.data, ['billingNodes', 'nodes']));
	const history = $derived(asArray<HistoryRow>(historyQ.data, ['records', 'history']));

	syncTableRows(handler, providers, { ignoreEmpty: true });
	syncTableRows(nodeHandler, billed, { ignoreEmpty: true });
	syncTableRows(historyHandler, history, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, providers, { ignoreEmpty: providersQ.loading });
	});
	$effect(() => {
		syncTableRows(nodeHandler, billed, { ignoreEmpty: nodesQ.loading });
	});
	$effect(() => {
		syncTableRows(historyHandler, history, { ignoreEmpty: historyQ.loading });
	});
	$effect(() => {
		savePageSize('infra-providers', handler.rowsPerPage);
		savePageSize('infra-nodes', nodeHandler.rowsPerPage);
		savePageSize('infra-history', historyHandler.rowsPerPage);
	});

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Add provider', onclick: () => (openCreate = true) },
			search: { table, placeholder: 'Search providers...', fields: ['name'] },
			toolbar: { table, view, onviewchange: () => saveColumnPrefs('infra-providers', view) }
		});
		return () => pageChrome.clear(token);
	});

	onMount(() => listHotkeys({ oncreate: () => (openCreate = true) }));

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount + 2, 1));
	const someRowsSelected = $derived(handler.rowCount.selected > 0 && !handler.isAllSelected);
	const selectedUuids = $derived(handler.selected.map(String));
	const nodeSelected = $derived(nodeHandler.rowCount.selected > 0 && !nodeHandler.isAllSelected);
	const historySelected = $derived(historyHandler.rowCount.selected > 0 && !historyHandler.isAllSelected);

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
			await Promise.all([providersQ.refetch(), nodesQ.refetch(), historyQ.refetch()]);
			if (selected) selected = providers.find((row) => row.uuid === selected?.uuid) ?? selected;
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	async function forSelected(fn: (uuid: string) => Promise<unknown>, ids: string[]) {
		for (const uuid of ids) await fn(uuid);
	}

	function openRow(row: Provider, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		selected = row;
	}
</script>

{#if providersQ.error && !providersQ.data}
	<ErrorState message={providersQ.error.message} onretry={() => providersQ.refetch()} />
{:else if !providersQ.loading && providers.length === 0 && billed.length === 0 && history.length === 0}
	<EmptyState
		title="No billing data yet"
		description="Add a provider to track infra invoices and billed nodes."
		action={{ label: 'Add provider', onclick: () => (openCreate = true) }}
	/>
{:else}
	<div class="flex min-h-0 flex-col gap-8">
		<section>
			<h2 class="mb-3 text-[13px] font-medium text-[var(--color-text-secondary)]">Providers</h2>
			<DataTable {table} headless>
				{#snippet header()}
					{#if handler.rowCount.selected > 0}
						<div class="mb-3 w-full">
							<ProfilesBulkBar
								count={handler.rowCount.selected}
								ondelete={() =>
									ask({
										title: `Delete ${handler.rowCount.selected} providers?`,
										description: providers
											.filter((row) => selectedUuids.includes(row.uuid))
											.slice(0, 5)
											.map((row) => row.name)
											.join(', '),
										confirmLabel: 'Delete',
										run: async () => {
											await run(
												() => forSelected((uuid) => removeP.mutate(uuid), [...selectedUuids]),
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

				<div>
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
									<ThSort {table} field="name">Name</ThSort>
								{/if}
								{#if view.columns[1]?.isVisible}
									<ThSort {table} field={(row) => row.billingHistory?.totalBills ?? 0} class="text-right" btnClass="justify-end">
										Bills
									</ThSort>
								{/if}
								{#if view.columns[2]?.isVisible}
									<ThSort {table} field={(row) => row.billingHistory?.totalAmount ?? 0} class="text-right" btnClass="justify-end">
										Spent
									</ThSort>
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
									<Table.Cell class="w-10 px-3">
										<Checkbox
											aria-label={`Select ${row.name}`}
											checked={handler.selected.includes(row.uuid)}
											onCheckedChange={() => handler.select(row.uuid)}
										/>
									</Table.Cell>
									{#if view.columns[0]?.isVisible}
										<Table.Cell class="max-w-[18rem] px-3">
											<button
												type="button"
												class="block max-w-full truncate font-medium text-[var(--color-text-primary)] hover:underline"
												onclick={() => (selected = row)}
											>
												{row.name}
											</button>
											{#if row.loginUrl}
												<CellSubtitle title={row.loginUrl}>{row.loginUrl}</CellSubtitle>
											{/if}
										</Table.Cell>
									{/if}
									{#if view.columns[1]?.isVisible}
										<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
											{row.billingHistory?.totalBills ?? 0}
										</Table.Cell>
									{/if}
									{#if view.columns[2]?.isVisible}
										<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">
											{row.billingHistory?.totalAmount ?? 0}
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
												<DropdownMenu.Separator />
												<DropdownMenu.Item
													variant="destructive"
													onclick={() =>
														ask({
															title: `Delete ${row.name}?`,
															description: 'Billing history for this provider stays unless deleted separately.',
															confirmLabel: 'Delete',
															run: () => run(() => removeP.mutate(row.uuid), 'Deleted')
														})}
												>
													Delete
												</DropdownMenu.Item>
											</DropdownMenu.Content>
										</DropdownMenu.Root>
									</Table.Cell>
								</Table.Row>
							{:else}
								{#if providersQ.loading}
									<TableSkeletonRows columns={emptyColspan} />
								{:else}
									<Table.Row class="hover:bg-transparent">
										<Table.Cell colspan={emptyColspan} class="py-8 text-center text-sm text-[var(--color-text-tertiary)]">
											No providers yet.
										</Table.Cell>
									</Table.Row>
								{/if}
							{/each}
						</Table.Body>
					</Table.Root>
				</div>

				{#snippet footer()}
			<TableStatusBar {table} loading={providersQ.loading} selection />
		{/snippet}
			</DataTable>
		</section>

		<section>
			<h2 class="mb-3 text-[13px] font-medium text-[var(--color-text-secondary)]">Billed nodes</h2>
			<DataTable table={nodeTable} headless>
				{#snippet header()}
					<div class="flex w-full min-w-0 flex-col gap-2">
						{#if nodeHandler.rowCount.selected > 0}
							<ProfilesBulkBar
								count={nodeHandler.rowCount.selected}
								ondelete={() =>
									ask({
										title: `Delete ${nodeHandler.rowCount.selected} billed nodes?`,
										description: 'They will no longer be tracked for this provider.',
										confirmLabel: 'Delete',
										run: async () => {
											await run(
												() => forSelected((uuid) => removeNode.mutate(uuid), nodeHandler.selected.map(String)),
												'Deleted'
											);
											nodeHandler.clearSelection();
										}
									})}
								onclear={() => nodeHandler.clearSelection()}
							/>
						{/if}
					</div>
				{/snippet}
				<div>
					<Table.Root class="min-w-full">
						<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
							<Table.Row class="hover:!bg-transparent">
								<ThLabel class="w-10">
									<Checkbox
										aria-label="Select all billed nodes"
										checked={nodeHandler.isAllSelected}
										indeterminate={nodeSelected}
										onCheckedChange={() => nodeHandler.selectAll()}
									/>
								</ThLabel>
								<ThSort table={nodeTable} field={(row) => row.node?.name ?? row.name ?? ''}>Node</ThSort>
								<ThSort table={nodeTable} field={(row) => row.provider?.name ?? ''}>Provider</ThSort>
								<ThSort table={nodeTable} field="nextBillingAt">Next bill</ThSort>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each nodeHandler.rows as row (row.uuid)}
								<Table.Row
									data-state={nodeHandler.selected.includes(row.uuid) ? 'selected' : undefined}
									class="group border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
								>
									<Table.Cell class="w-10 px-3">
										<Checkbox
											aria-label={`Select ${row.node?.name ?? row.name ?? row.uuid}`}
											checked={nodeHandler.selected.includes(row.uuid)}
											onCheckedChange={() => nodeHandler.select(row.uuid)}
										/>
									</Table.Cell>
									<Table.Cell class="px-3">
										<span class="font-medium">{row.node?.name ?? row.name ?? '—'}</span>
										{#if row.node?.countryCode}
											<CellSubtitle>{row.node.countryCode}</CellSubtitle>
										{/if}
									</Table.Cell>
									<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{row.provider?.name ?? '—'}</Table.Cell>
									<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{formatDate(row.nextBillingAt)}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row class="hover:bg-transparent">
									<Table.Cell colspan={4} class="py-8 text-center text-sm text-[var(--color-text-tertiary)]">
										No billed nodes.
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				{#snippet footer()}
			<TableStatusBar table={nodeTable} loading={nodesQ.loading} selection />
		{/snippet}
			</DataTable>
		</section>

		<section>
			<h2 class="mb-3 text-[13px] font-medium text-[var(--color-text-secondary)]">History</h2>
			<DataTable table={historyTable} headless>
				{#snippet header()}
					<div class="flex w-full min-w-0 flex-col gap-2">
						{#if historyHandler.rowCount.selected > 0}
							<ProfilesBulkBar
								count={historyHandler.rowCount.selected}
								ondelete={() =>
									ask({
										title: `Delete ${historyHandler.rowCount.selected} history records?`,
										description: 'This cannot be undone.',
										confirmLabel: 'Delete',
										run: async () => {
											await run(
												() =>
													forSelected((uuid) => removeHistory.mutate(uuid), historyHandler.selected.map(String)),
												'Deleted'
											);
											historyHandler.clearSelection();
										}
									})}
								onclear={() => historyHandler.clearSelection()}
							/>
						{/if}
					</div>
				{/snippet}
				<div>
					<Table.Root class="min-w-full">
						<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
							<Table.Row class="hover:!bg-transparent">
								<ThLabel class="w-10">
									<Checkbox
										aria-label="Select all history rows"
										checked={historyHandler.isAllSelected}
										indeterminate={historySelected}
										onCheckedChange={() => historyHandler.selectAll()}
									/>
								</ThLabel>
								<ThSort table={historyTable} field="billedAt">Billed</ThSort>
								<ThSort table={historyTable} field={(row) => row.provider?.name ?? ''}>Provider</ThSort>
								<ThSort table={historyTable} field="amount" class="text-right" btnClass="justify-end">Amount</ThSort>
							</Table.Row>
						</Table.Header>
						<Table.Body>
							{#each historyHandler.rows as row (row.uuid)}
								<Table.Row
									data-state={historyHandler.selected.includes(row.uuid) ? 'selected' : undefined}
									class="group border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
								>
									<Table.Cell class="w-10 px-3">
										<Checkbox
											aria-label={`Select bill ${row.uuid}`}
											checked={historyHandler.selected.includes(row.uuid)}
											onCheckedChange={() => historyHandler.select(row.uuid)}
										/>
									</Table.Cell>
									<Table.Cell class="px-3">{formatDate(row.billedAt)}</Table.Cell>
									<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{row.provider?.name ?? '—'}</Table.Cell>
									<Table.Cell class="px-3 text-right text-[var(--color-text-secondary)]">{row.amount}</Table.Cell>
								</Table.Row>
							{:else}
								<Table.Row class="hover:bg-transparent">
									<Table.Cell colspan={4} class="py-8 text-center text-sm text-[var(--color-text-tertiary)]">
										No billing history.
									</Table.Cell>
								</Table.Row>
							{/each}
						</Table.Body>
					</Table.Root>
				</div>
				{#snippet footer()}
			<TableStatusBar table={historyTable} loading={historyQ.loading} selection />
		{/snippet}
			</DataTable>
		</section>
	</div>
{/if}

<ProviderCreateDialog
	bind:open={openCreate}
	pending={createP.pending}
	oncreate={async (body) => {
		await run(() => createP.mutate(body), 'Created');
	}}
/>

<ProviderDetailSheet
	bind:provider={selected}
	pending={updateP.pending}
	onupdate={async (body) => {
		await run(() => updateP.mutate(body), 'Saved');
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.name}?`,
			description: 'Billing history for this provider stays unless deleted separately.',
			confirmLabel: 'Delete',
			run: async () => {
				await run(() => removeP.mutate(selected!.uuid), 'Deleted');
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
