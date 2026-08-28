<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import * as Table from '$lib/components/ui/table';
	import {
		DataTable,
		TableStatusBar,
		ThSort,
		syncTableRows
	} from '$lib/components/ui/data-table';
	import CellSubtitle from '$lib/components/remnawave/CellSubtitle.svelte';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import ConfirmAction from '$lib/components/remnawave/ConfirmAction.svelte';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { resolveUserBody } from '$lib/features/layout/quick-open.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray, pretty } from '$lib/list';
	import { rw } from '$lib/rw';
	import { TableHandler, type TableHandlerInterface } from '@vincjo/datatables';

	type Conn = Record<string, unknown> & { key: string };

	const startByUser = rw.connections.startByUser();
	const startByNode = rw.connections.startByNode();
	const drop = rw.connections.drop();
	const resolveM = rw.users.resolve();

	let jobId = $state('');
	let kind = $state<'user' | 'node'>('user');
	let confirmDrop = $state(false);
	let selected = $state<Conn | null>(null);
	let seenLaunch = $state('');

	const resultUser = rw.connections.resultByUser(() => jobId);
	const resultNode = rw.connections.resultByNode(() => jobId);
	const result = $derived(kind === 'user' ? resultUser : resultNode);

	const handler = new TableHandler<Conn>([], { rowsPerPage: 20 });
	const table = handler as TableHandlerInterface<Conn>;

	const rows = $derived.by(() => {
		const data = result.data;
		const list = asArray<Record<string, unknown>>(data, ['connections', 'sessions', 'items']);
		return list.map((row, index) => ({
			...row,
			key: String(row.uuid ?? row.id ?? index)
		})) as Conn[];
	});

	syncTableRows(handler, rows, { ignoreEmpty: true });
	$effect(() => {
		syncTableRows(handler, rows, { ignoreEmpty: result.loading });
	});
	$effect.pre(() => {
		const token = pageChrome.set({
			search: jobId ? { table, placeholder: 'Search sessions...' } : null,
			toolbar: jobId ? { table } : null,
			action: jobId
				? {
						label: 'Refresh sessions',
						icon: 'sync',
						onclick: () => result.refetch(),
						pending: () => result.loading
					}
				: null
		});
		return () => pageChrome.clear(token);
	});

	$effect(() => {
		const params = page.url.searchParams;
		if (params.get('drop') === '1') {
			confirmDrop = true;
			const url = new URL(page.url.href);
			url.searchParams.delete('drop');
			void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
		}
		const nextKind = params.get('kind') === 'node' ? 'node' : params.get('kind') === 'user' ? 'user' : null;
		const id = params.get('id')?.trim() ?? '';
		if (!nextKind || !id) return;
		const key = `${nextKind}:${id}`;
		if (seenLaunch === key) return;
		seenLaunch = key;
		kind = nextKind;
		void start(nextKind, id);
	});

	async function start(nextKind: 'user' | 'node', target: string) {
		try {
			let userKey = target;
			if (nextKind === 'user' && !/^\d+$/.test(target)) {
				const user = await resolveM.mutate(resolveUserBody(target));
				userKey = String(user.id);
			}
			const res =
				nextKind === 'user'
					? await startByUser.mutate({ userId: userKey })
					: await startByNode.mutate({ uuid: target });
			const rec = res as Record<string, unknown>;
			jobId = String(rec.jobId ?? rec.id ?? '');
			appToast.success('Job started');
		} catch (err) {
			appToast.apiError(err, 'Failed to start');
		}
	}
</script>

{#if !jobId}
	<EmptyState
		title="No session snapshot"
		description="Start a job from search: #42 for a user, #node uuid for a node, or > Sessions."
	/>
{:else if result.error}
	<p class="text-sm text-[var(--color-error)]">{result.error.message}</p>
{:else if rows.length > 0}
	<DataTable {table} headless>
		<div>
			<Table.Root class="min-w-full">
				<Table.Header class="sticky top-0 z-10 bg-[var(--color-bg-secondary)]">
					<Table.Row class="hover:!bg-transparent">
						<ThSort {table} field={(row) => row.username ?? row.userId ?? row.uuid ?? ''}>Session</ThSort>
						<ThSort {table} field={(row) => row.nodeName ?? row.nodeUuid ?? ''}>Node</ThSort>
						<ThSort {table} field={(row) => row.ip ?? row.remoteIp ?? ''}>IP</ThSort>
					</Table.Row>
				</Table.Header>
				<Table.Body>
					{#each handler.rows as row (row.key)}
						<Table.Row
							class="cursor-pointer border-[var(--app-border)] hover:!bg-[var(--color-bg-hover)]"
							onclick={() => (selected = row)}
						>
							<Table.Cell class="max-w-[16rem] px-3">
								<span class="block truncate font-medium">{String(row.username ?? row.userId ?? row.uuid ?? row.key)}</span>
								<CellSubtitle>{String(row.protocol ?? row.inboundTag ?? '')}</CellSubtitle>
							</Table.Cell>
							<Table.Cell class="max-w-[14rem] px-3 text-[var(--color-text-secondary)]">
								<span class="block truncate">{String(row.nodeName ?? row.nodeUuid ?? '—')}</span>
							</Table.Cell>
							<Table.Cell class="px-3 font-mono text-xs">{String(row.ip ?? row.remoteIp ?? '—')}</Table.Cell>
						</Table.Row>
					{/each}
				</Table.Body>
			</Table.Root>
		</div>
		{#snippet footer()}
			<TableStatusBar {table} loading={result.loading} />
		{/snippet}
	</DataTable>
{:else}
	<pre class="rounded-md border border-[var(--app-border)] p-3 font-mono text-xs text-[var(--color-text-secondary)]">{pretty(result.data ?? { jobId, status: result.loading ? 'loading' : 'empty' })}</pre>
{/if}

<DetailSheet
	open={selected !== null}
	title="Session"
	description={selected ? String(selected.key) : ''}
	actions={[]}
	onrun={() => {}}
	onOpenChange={(value) => {
		if (!value) selected = null;
	}}
>
	{#if selected}
		<pre class="overflow-auto font-mono text-xs text-[var(--color-text-secondary)]">{pretty(selected)}</pre>
	{/if}
</DetailSheet>

<ConfirmAction
	bind:open={confirmDrop}
	title="Drop all connections?"
	description="Active sessions will be disconnected."
	confirmLabel="Drop"
	onconfirm={async () => {
		await drop.mutate({});
		appToast.success('Drop queued');
	}}
/>
