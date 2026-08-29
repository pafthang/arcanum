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
		ThLabel,
		ThSort,
		TableSkeletonRows,
		TableStatusBar,
		syncTableRows
	} from '$lib/components/ui/data-table';
	import CellSubtitle from '$lib/components/remnawave/CellSubtitle.svelte';
	import StatusBadge from '$lib/components/remnawave/StatusBadge.svelte';
	import ExpireLabel from '$lib/components/remnawave/ExpireLabel.svelte';
	import TrafficBar from '$lib/components/remnawave/TrafficBar.svelte';
	import ConfirmAction from '$lib/components/remnawave/ConfirmAction.svelte';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import UserCreateDialog from '$lib/features/users/UserCreateDialog.svelte';
	import UserDetailSheet from '$lib/features/users/UserDetailSheet.svelte';
	import UsersBulkBar from '$lib/features/users/UsersBulkBar.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { listHotkeys } from '$lib/features/layout/list-hotkeys';
	import { columnVisible, loadColumnPrefs, loadPageSize, saveColumnPrefs, savePageSize } from '$lib/features/layout/table-prefs';
	import { appToast } from '$lib/features/toast/toast';
	import { copyText } from '$lib/copy';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';
	import type { User } from '@arcanum/ts-client';

	let start = $state(0);
	let size = $state(loadPageSize('users', 20));
	let search = $state('');
	let searchDraft = $state('');
	let statusFilter = $state('all');

	const list = rw.users.list(() => ({
		start,
		size,
		search: search || undefined,
		status: statusFilter === 'all' ? undefined : statusFilter
	}));
	const createM = rw.users.create();
	const updateM = rw.users.update();
	const removeM = rw.users.remove();
	const enableM = rw.users.enable();
	const disableM = rw.users.disable();
	const resetM = rw.users.resetTraffic();
	const revokeM = rw.users.revoke();
	const extendM = rw.users.extend();
	const bulkDelete = rw.users.bulkDelete();
	const bulkReset = rw.users.bulkResetTraffic();
	const bulkExtend = rw.users.bulkExtend();
	const bulkUpdate = rw.users.bulkUpdate();

	const colPrefs = loadColumnPrefs('users');
	const handler = new TableHandler<User>([], { rowsPerPage: size, selectBy: 'id' });
	const table = handler as TableHandlerInterface<User>;
	const view = handler.createView([
		{ index: 1, name: 'Username', isVisible: columnVisible(colPrefs, 'Username') },
		{ index: 2, name: 'Status', isVisible: columnVisible(colPrefs, 'Status') },
		{ index: 3, name: 'Expire', isVisible: columnVisible(colPrefs, 'Expire') },
		{ index: 4, name: 'Traffic', isVisible: columnVisible(colPrefs, 'Traffic') },
		{ index: 5, name: 'Tag', isVisible: columnVisible(colPrefs, 'Tag') }
	]);

	let openCreate = $state(false);
	let selected = $state<User | null>(null);
	let confirmOpen = $state(false);
	let confirmSpec = $state({
		title: '',
		description: '',
		confirmLabel: 'Confirm',
		variant: 'destructive' as 'destructive' | 'default',
		run: async () => {}
	});

	const source = $derived(list.data?.users ?? []);
	const total = $derived(list.data?.total ?? 0);
	const filtered = $derived(Boolean(search || statusFilter !== 'all'));
	const statusOptions = [
		{ value: 'all', label: 'All' },
		{ value: 'ACTIVE', label: 'Active' },
		{ value: 'DISABLED', label: 'Disabled' },
		{ value: 'LIMITED', label: 'Limited' },
		{ value: 'EXPIRED', label: 'Expired' }
	];

	syncTableRows(handler, list.data?.users ?? [], { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, source, { ignoreEmpty: list.loading });
	});
	$effect(() => {
		const next = handler.rowsPerPage;
		if (next !== size) {
			size = next;
			start = 0;
			savePageSize('users', next);
		}
	});
	$effect(() => {
		if (list.loading) return;
		if (total === 0) {
			if (start !== 0) start = 0;
			return;
		}
		if (start >= total) start = Math.max(0, (Math.ceil(total / size) - 1) * size);
	});

	let searchTimer: ReturnType<typeof setTimeout> | undefined;
	function onsearch() {
		clearTimeout(searchTimer);
		searchTimer = setTimeout(() => {
			search = searchDraft.trim();
			start = 0;
		}, 200);
	}

	$effect.pre(() => {
		const token = pageChrome.set({
			create: { label: 'Create user', onclick: () => (openCreate = true) },
			search: {
				placeholder: 'Search users...',
				value: () => searchDraft,
				setValue: (value) => {
					searchDraft = value;
				},
				oninput: onsearch
			},
			toolbar: {
				table,
				view,
				onviewchange: () => saveColumnPrefs('users', view),
				filter: {
					value: () => statusFilter,
					options: statusOptions,
					onselect: (value) => {
						statusFilter = value;
						start = 0;
					}
				}
			}
		});
		return () => pageChrome.clear(token);
	});

	onMount(() => listHotkeys({ oncreate: () => (openCreate = true) }));

	const visibleColumnCount = $derived(view.columns.filter((column) => column.isVisible).length);
	const emptyColspan = $derived(Math.max(visibleColumnCount + 2, 1));
	const someRowsSelected = $derived(handler.rowCount.selected > 0 && !handler.isAllSelected);
	const selectedIds = $derived(handler.selected.map((id) => Number(id)).filter((id) => Number.isFinite(id)));

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

	const openId = $derived(page.url.searchParams.get('id') ?? '');
	const peeked = rw.users.byId(() => openId);

	$effect(() => {
		if (!openId) return;
		const found = source.find((user) => String(user.id) === openId);
		if (found) {
			selected = found;
			return;
		}
		if (peeked.data && String(peeked.data.id) === openId) selected = peeked.data;
	});

	$effect(() => {
		if (selected !== null || !openId || peeked.loading || peeked.data) return;
		const url = new URL(page.url.href);
		url.searchParams.delete('id');
		void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
	});

	async function run(fn: () => Promise<unknown>, ok: string) {
		try {
			await fn();
			appToast.success(ok);
			await list.refetch();
			if (selected) {
				selected = (list.data?.users ?? []).find((user) => user.id === selected?.id) ?? selected;
			}
		} catch (err) {
			appToast.apiError(err, 'Request failed');
			throw err;
		}
	}

	function selectedUsers(): User[] {
		const ids = new Set(selectedIds);
		return source.filter((user) => ids.has(user.id));
	}

	function openUser(user: User) {
		selected = user;
		const url = new URL(page.url.href);
		if (url.searchParams.get('id') !== String(user.id)) {
			url.searchParams.set('id', String(user.id));
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
	}

	function openRow(user: User, event: MouseEvent) {
		if ((event.target as HTMLElement).closest('button, a, input, [role="checkbox"]')) return;
		openUser(user);
	}
</script>

{#if list.error && !list.data}
	<ErrorState message={list.error.message} onretry={() => list.refetch()} />
{:else if !list.loading && total === 0 && !filtered}
	<EmptyState
		title="No users yet"
		description="Create the first account to start managing access, traffic and subscriptions."
		action={{ label: 'Create user', onclick: () => (openCreate = true) }}
	/>
{:else}
	<DataTable {table} headless>
		{#snippet header()}
			{#if handler.rowCount.selected > 0}
				<div class="mb-3 w-full">
					<UsersBulkBar
						count={handler.rowCount.selected}
						onenable={() =>
							void run(
								() => bulkUpdate.mutate({ userIds: selectedIds, fields: { status: 'ACTIVE' } }),
								'Enabled'
							)}
						ondisable={() =>
							ask({
								title: `Disable ${handler.rowCount.selected} users?`,
								description: 'Disabled accounts cannot connect until they are enabled again.',
								confirmLabel: 'Disable',
								run: () =>
									run(
										() => bulkUpdate.mutate({ userIds: selectedIds, fields: { status: 'DISABLED' } }),
										'Disabled'
									)
							})}
						onextend={() =>
							void run(
								() => bulkExtend.mutate({ userIds: selectedIds, extendDays: 30 }),
								'Extended 30d'
							)}
						onreset={() =>
							ask({
								title: `Reset traffic for ${handler.rowCount.selected} users?`,
								description: 'Used traffic counters will be set to zero.',
								confirmLabel: 'Reset',
								run: () => run(() => bulkReset.mutate({ userIds: selectedIds }), 'Traffic reset')
							})}
						ondelete={() =>
							ask({
								title: `Delete ${handler.rowCount.selected} users?`,
								description: selectedUsers()
									.slice(0, 5)
									.map((user) => user.username)
									.join(', ')
									.concat(handler.rowCount.selected > 5 ? '…' : ''),
								confirmLabel: 'Delete',
								run: async () => {
									await run(() => bulkDelete.mutate({ userIds: selectedIds }), 'Deleted');
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
							<ThSort {table} field="username" class="sticky left-10 z-20 bg-[var(--color-bg-secondary)]">User</ThSort>
						{/if}
						{#if view.columns[1]?.isVisible}
							<ThSort {table} field="status">Status</ThSort>
						{/if}
						{#if view.columns[2]?.isVisible}
							<ThSort {table} field="expireAt">Expire</ThSort>
						{/if}
						{#if view.columns[3]?.isVisible}
							<ThSort
								{table}
								field={(row) => row.userTraffic?.usedTrafficBytes ?? 0}
								class="text-right"
								btnClass="justify-end"
							>
								Traffic
							</ThSort>
						{/if}
						{#if view.columns[4]?.isVisible}
							<ThSort {table} field="tag">Tag</ThSort>
						{/if}
						<ThLabel class="w-12 text-right"><span class="sr-only">Actions</span></ThLabel>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as user (user.id)}
						<Table.Row
							data-state={handler.selected.includes(user.id) ? 'selected' : undefined}
							class="group cursor-pointer border-[var(--app-border)] transition-none hover:!bg-[var(--color-bg-hover)] data-[state=selected]:!bg-[var(--color-bg-hover)]"
							onclick={(event) => openRow(user, event)}
						>
							<Table.Cell class="sticky left-0 z-10 w-10 bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
								<Checkbox
									aria-label={`Select ${user.username}`}
									checked={handler.selected.includes(user.id)}
									onCheckedChange={() => handler.select(user.id)}
								/>
							</Table.Cell>
							{#if view.columns[0]?.isVisible}
								<Table.Cell class="sticky left-10 z-10 max-w-[18rem] bg-[var(--color-bg)] px-3 group-hover:!bg-[var(--color-bg-hover)] group-data-[state=selected]:!bg-[var(--color-bg-hover)]">
									<button
										type="button"
										class="block max-w-full truncate text-sm font-medium text-[var(--color-text-primary)] hover:underline"
										onclick={() => openUser(user)}
									>
										{user.username}
									</button>
									<CellSubtitle title={`#${user.id} · ${user.shortUuid}`}>#{user.id} · {user.shortUuid}</CellSubtitle>
								</Table.Cell>
							{/if}
							{#if view.columns[1]?.isVisible}
								<Table.Cell class="px-3"><StatusBadge value={user.status} /></Table.Cell>
							{/if}
							{#if view.columns[2]?.isVisible}
								<Table.Cell class="px-3"><ExpireLabel value={user.expireAt} /></Table.Cell>
							{/if}
							{#if view.columns[3]?.isVisible}
								<Table.Cell class="px-3">
									<TrafficBar compact used={user.userTraffic?.usedTrafficBytes} limit={user.trafficLimitBytes} />
								</Table.Cell>
							{/if}
							{#if view.columns[4]?.isVisible}
								<Table.Cell class="px-3 text-[var(--color-text-secondary)]">{user.tag ?? '—'}</Table.Cell>
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
										<DropdownMenu.Item onclick={() => openUser(user)}>Open</DropdownMenu.Item>
										<DropdownMenu.Item
											onclick={async () => {
												if (await copyText(user.subscriptionUrl)) appToast.success('Subscription URL copied');
											}}
										>
											Copy sub
										</DropdownMenu.Item>
										<DropdownMenu.Item onclick={() => run(() => extendM.mutate({ userId: user.id, days: 30 }), 'Extended 30d')}>
											+30 days
										</DropdownMenu.Item>
										{#if user.status === 'DISABLED'}
											<DropdownMenu.Item onclick={() => run(() => enableM.mutate(user.id), 'Enabled')}>
												Enable
											</DropdownMenu.Item>
										{:else}
											<DropdownMenu.Item
												onclick={() =>
													ask({
														title: `Disable ${user.username}?`,
														description: 'This account will not be able to connect until you enable it again.',
														confirmLabel: 'Disable',
														run: () => run(() => disableM.mutate(user.id), 'Disabled')
													})}
											>
												Disable
											</DropdownMenu.Item>
										{/if}
										<DropdownMenu.Item
											onclick={() =>
												ask({
													title: `Reset traffic for ${user.username}?`,
													description: 'Used traffic will be set to zero.',
													confirmLabel: 'Reset',
													run: () => run(() => resetM.mutate(user.id), 'Traffic reset')
												})}
										>
											Reset traffic
										</DropdownMenu.Item>
										<DropdownMenu.Separator />
										<DropdownMenu.Item
											variant="destructive"
											onclick={() =>
												ask({
													title: `Delete ${user.username}?`,
													description: 'This cannot be undone. Subscription and credentials will stop working.',
													confirmLabel: 'Delete',
													run: () => run(() => removeM.mutate(user.id), 'Deleted')
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
									No users match the current table state.
									<button
										type="button"
										class="ml-1 text-[var(--app-accent-light)] hover:underline"
										onclick={() => {
											statusFilter = 'all';
											searchDraft = '';
											search = '';
											start = 0;
										}}
									>
										Clear filters
									</button>
								</Table.Cell>
							</Table.Row>
						{/if}
					{/each}
				</Table.Body>
			</Table.Root>
		</div>

		{#snippet footer()}
			<TableStatusBar {total} {start} {size} onstart={(next) => (start = next)} />
		{/snippet}
	</DataTable>
{/if}

<UserCreateDialog
	bind:open={openCreate}
	pending={createM.pending}
	oncreate={async (body) => {
		await run(() => createM.mutate(body), 'User created');
	}}
/>

<UserDetailSheet
	bind:user={selected}
	pending={updateM.pending}
	onupdate={async (body) => {
		await run(() => updateM.mutate(body), 'Saved');
	}}
	onenable={() => {
		const user = selected;
		if (user) void run(() => enableM.mutate(user.id), 'Enabled');
	}}
	ondisable={() => {
		if (!selected) return;
		ask({
			title: `Disable ${selected.username}?`,
			description: 'This account will not be able to connect until you enable it again.',
			confirmLabel: 'Disable',
			run: () => run(() => disableM.mutate(selected!.id), 'Disabled')
		});
	}}
	onextend={() => {
		const user = selected;
		if (user) void run(() => extendM.mutate({ userId: user.id, days: 30 }), 'Extended 30d');
	}}
	onreset={() => {
		if (!selected) return;
		ask({
			title: `Reset traffic for ${selected.username}?`,
			description: 'Used traffic will be set to zero.',
			confirmLabel: 'Reset',
			run: () => run(() => resetM.mutate(selected!.id), 'Traffic reset')
		});
	}}
	onrevoke={() => {
		if (!selected) return;
		ask({
			title: `Revoke subscription for ${selected.username}?`,
			description: 'Current subscription URL and secrets will be rotated.',
			confirmLabel: 'Revoke',
			run: () => run(() => revokeM.mutate({ userId: selected!.id }), 'Subscription revoked')
		});
	}}
	ondelete={() => {
		if (!selected) return;
		ask({
			title: `Delete ${selected.username}?`,
			description: 'This cannot be undone. Subscription and credentials will stop working.',
			confirmLabel: 'Delete',
			run: async () => {
				await run(() => removeM.mutate(selected!.id), 'Deleted');
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
