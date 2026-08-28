<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import RadioIcon from '@lucide/svelte/icons/radio';
	import SearchIcon from '@lucide/svelte/icons/search';
	import UserIcon from '@lucide/svelte/icons/user';
	import XIcon from '@lucide/svelte/icons/x';
	import * as Command from '$lib/components/ui/command';
	import * as InputGroup from '$lib/components/ui/input-group';
	import * as Kbd from '$lib/components/ui/kbd';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { searchChord } from '$lib/features/layout/shortcuts.svelte';
	import {
		filterCommands,
		filterPages,
		parseSessionTarget,
		quickOpen,
		resolveUserBody,
		searchUsers,
		sessionHref,
		type QuickOpenCommand,
		type QuickOpenPage,
		type SessionTarget
	} from '$lib/features/layout/quick-open.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';
	import type { User } from '@optimawave/ts-back';

	type TableSearch = { value: string; set: () => void };
	type PaletteItem =
		| { kind: 'command'; key: string; command: QuickOpenCommand }
		| { kind: 'page'; key: string; page: QuickOpenPage }
		| { kind: 'session'; key: string; target: SessionTarget }
		| { kind: 'user'; key: string; user: User };

	const resolveM = rw.users.resolve();
	const spec = $derived(pageChrome.search);
	const chord = searchChord();

	let query = $state('');
	let selected = $state('');
	let open = $state(false);
	let keepTable = $state(false);
	let users = $state<User[]>([]);
	let userPending = $state(false);
	let tableSearch = $state<TableSearch | null>(null);
	let boundTable: object | null = null;

	const mode = $derived<'all' | 'commands' | 'users' | 'sessions'>(
		query.startsWith('#')
			? 'sessions'
			: query.startsWith('>')
				? 'commands'
				: query.startsWith('@')
					? 'users'
					: 'all'
	);
	const text = $derived(mode === 'all' ? query : query.slice(1).trimStart());
	const pages = $derived(
		mode === 'users' || mode === 'sessions' || mode === 'commands' ? [] : filterPages(text)
	);
	const commands = $derived(
		mode === 'commands' || (mode === 'all' && text.trim()) ? filterCommands(text) : []
	);
	const sessionTarget = $derived(mode === 'sessions' ? parseSessionTarget(text) : null);
	const items = $derived.by((): PaletteItem[] => {
		const out: PaletteItem[] = [];
		if (sessionTarget) {
			out.push({
				kind: 'session',
				key: `session:${sessionTarget.kind}:${sessionTarget.id}`,
				target: sessionTarget
			});
		}
		if (mode === 'commands' || mode === 'all') {
			for (const command of commands) out.push({ kind: 'command', key: `cmd:${command.id}`, command });
		}
		if (mode === 'all') {
			for (const page of pages) out.push({ kind: 'page', key: `page:${page.href}`, page });
		}
		if (mode !== 'commands') {
			for (const user of users) out.push({ kind: 'user', key: `user:${user.id}`, user });
		}
		return out;
	});
	const commandItems = $derived(
		items.filter((item): item is Extract<PaletteItem, { kind: 'command' }> => item.kind === 'command')
	);
	const sessionItems = $derived(
		items.filter((item): item is Extract<PaletteItem, { kind: 'session' }> => item.kind === 'session')
	);
	const pageItems = $derived(
		items.filter((item): item is Extract<PaletteItem, { kind: 'page' }> => item.kind === 'page')
	);
	const userItems = $derived(
		items.filter((item): item is Extract<PaletteItem, { kind: 'user' }> => item.kind === 'user')
	);
	const placeholder = $derived(spec?.placeholder ?? 'Search or go to...');
	const tableLabel = $derived(tableSearch?.value || spec?.value?.() || '');

	let tableTimer: ReturnType<typeof setTimeout> | undefined;
	let userTimer: ReturnType<typeof setTimeout> | undefined;
	let userGen = 0;
	let seenNonce = 0;

	$effect.pre(() => {
		const table = spec?.table ?? null;
		if (table !== boundTable) {
			boundTable = table;
			tableSearch = table
				? (
						table as { createSearch?: (fields?: unknown) => TableSearch }
					).createSearch?.(spec?.fields) ?? null
				: null;
		}
	});

	$effect(() => {
		if (!items.some((item) => item.key === selected)) selected = items[0]?.key ?? '';
	});

	$effect(() => {
		if (!open) return;
		const next = query;
		queueUsers();
		if (next.startsWith('>') || next.startsWith('@') || next.startsWith('#')) return;
		if (keepTable && !next) return;
		applyTable(next);
	});

	function applyTable(next: string) {
		if (tableSearch) {
			tableSearch.value = next;
			clearTimeout(tableTimer);
			tableTimer = setTimeout(() => tableSearch?.set(), 160);
			return;
		}
		spec?.setValue?.(next);
		spec?.oninput?.();
	}

	function queueUsers() {
		clearTimeout(userTimer);
		const skipUsers =
			mode === 'commands' ||
			(mode === 'sessions' && (sessionTarget?.kind === 'node' || text.trim().length < 2)) ||
			text.trim().length < 2;
		if (skipUsers) {
			userGen += 1;
			users = [];
			userPending = false;
			return;
		}
		userPending = true;
		userTimer = setTimeout(() => {
			void loadUsers(text);
		}, 160);
	}

	async function loadUsers(q: string) {
		const gen = ++userGen;
		try {
			const rows = await searchUsers(q);
			if (gen !== userGen) return;
			users = rows;
		} catch {
			if (gen !== userGen) return;
			users = [];
		} finally {
			if (gen === userGen) userPending = false;
		}
	}

	function showPalette(init = '', opts?: { keepTable?: boolean }) {
		query = init;
		keepTable = Boolean(opts?.keepTable) && !init;
		open = true;
	}

	function clearTable(event: MouseEvent) {
		event.preventDefault();
		event.stopPropagation();
		applyTable('');
	}

	async function go(item: PaletteItem) {
		if (item.kind === 'command') {
			if (item.command.onRun) {
				open = false;
				item.command.onRun();
				return;
			}
			if (item.command.prefix != null) {
				query = item.command.prefix;
				keepTable = true;
				return;
			}
			if (item.command.href) {
				open = false;
				await goto(item.command.href);
			}
			return;
		}
		open = false;
		if (item.kind === 'page') {
			await goto(item.page.href);
			return;
		}
		if (item.kind === 'session') {
			await goto(sessionHref(item.target));
			return;
		}
		if (mode === 'sessions') {
			await goto(sessionHref({ kind: 'user', id: String(item.user.id) }));
			return;
		}
		await goto(`/dashboard/management/users?id=${item.user.id}`);
	}

	async function resolveUser(raw: string) {
		const trimmed = raw.trim();
		if (!trimmed) return;
		try {
			const user = await resolveM.mutate(resolveUserBody(trimmed));
			open = false;
			await goto(`/dashboard/management/users?id=${user.id}`);
		} catch (err) {
			appToast.apiError(err, 'Not found');
		}
	}

	async function submitEmpty() {
		if (mode === 'sessions') {
			const target = parseSessionTarget(text);
			if (target) {
				open = false;
				await goto(sessionHref(target));
			}
			return;
		}
		if (mode !== 'commands' && text.trim()) await resolveUser(text);
	}

	function onDialogKeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || items.length) return;
		event.preventDefault();
		void submitEmpty();
	}

	function onglobal(event: KeyboardEvent) {
		const meta = event.metaKey || event.ctrlKey;
		if (meta && event.key.toLowerCase() === 'p') {
			event.preventDefault();
			showPalette(event.shiftKey ? '>' : '', { keepTable: !event.shiftKey });
			return;
		}
		const target = event.target as HTMLElement | null;
		const typing = Boolean(target?.closest('input, textarea, select, [contenteditable="true"]'));
		if (event.key === '/' && !typing && !meta && !event.altKey) {
			event.preventDefault();
			showPalette('', { keepTable: true });
		}
	}

	$effect(() => {
		const nonce = quickOpen.nonce;
		if (!nonce || nonce === seenNonce) return;
		seenNonce = nonce;
		showPalette(quickOpen.prefix, { keepTable: !quickOpen.prefix });
	});

	onMount(() => {
		window.addEventListener('keydown', onglobal);
		return () => {
			window.removeEventListener('keydown', onglobal);
			clearTimeout(tableTimer);
			clearTimeout(userTimer);
		};
	});
</script>

<div class="w-full" data-quick-open>
	<InputGroup.Root class="h-8 w-full">
		<InputGroup.Addon>
			<SearchIcon class="size-3.5" />
		</InputGroup.Addon>
		<button
			type="button"
			id="list-search"
			class="flex h-full min-w-0 flex-1 cursor-pointer items-center truncate px-2 text-left text-sm outline-none {tableLabel
				? 'text-[var(--color-text-primary)]'
				: 'text-[var(--color-text-tertiary)]'}"
			onclick={() => showPalette(tableLabel)}
		>
			{tableLabel || placeholder}
		</button>
		{#if tableLabel}
			<InputGroup.Addon align="inline-end">
				<button
					type="button"
					class="rounded p-0.5 hover:text-[var(--color-text-primary)]"
					aria-label="Clear search"
					onclick={clearTable}
				>
					<XIcon class="size-3.5" />
				</button>
			</InputGroup.Addon>
		{/if}
		<InputGroup.Addon align="inline-end">
			<Kbd.Root class="text-[10px] text-[var(--color-text-tertiary)]">{chord}</Kbd.Root>
		</InputGroup.Addon>
	</InputGroup.Root>
</div>

<Command.Dialog
	bind:open
	bind:value={selected}
	shouldFilter={false}
	loop
	title="Search"
	description="Search pages, users and commands."
	class="sm:max-w-lg"
>
	<Command.Input
		placeholder="Type a command or search..."
		bind:value={query}
		onkeydown={onDialogKeydown}
	/>
	<Command.List>
		{#if userPending || resolveM.pending}
			<Command.Loading class="px-2.5 py-2 text-xs text-[var(--color-text-tertiary)]">
				Searching…
			</Command.Loading>
		{/if}
		<Command.Empty class="text-[var(--color-text-tertiary)]">
			{#if mode === 'users'}
				Type a username, id or short UUID
			{:else if mode === 'sessions'}
				Type a user id or node uuid
			{:else if mode === 'commands'}
				No matching commands
			{:else if text.trim()}
				No results found.
			{:else}
				Type to filter · @ user · # sessions · &gt; command
			{/if}
		</Command.Empty>
		{#if commandItems.length}
			<Command.Group heading="Commands">
				{#each commandItems as item (item.key)}
					{@const Icon = item.command.icon}
					<Command.Item
						value={item.key}
						keywords={[item.command.label, item.command.hint]}
						onSelect={() => void go(item)}
					>
						<Icon class="size-4" />
						<span>{item.command.label}</span>
						<Command.Shortcut>{item.command.hint}</Command.Shortcut>
					</Command.Item>
				{/each}
			</Command.Group>
		{/if}
		{#if sessionItems.length}
			<Command.Group heading="Sessions">
				{#each sessionItems as item (item.key)}
					<Command.Item value={item.key} onSelect={() => void go(item)}>
						<RadioIcon class="size-4" />
						<span>
							Inspect {item.target.kind === 'user' ? 'user' : 'node'}
							{item.target.id}
						</span>
						<Command.Shortcut>Start job</Command.Shortcut>
					</Command.Item>
				{/each}
			</Command.Group>
		{/if}
		{#if pageItems.length}
			<Command.Group heading="Pages">
				{#each pageItems as item (item.key)}
					{@const Icon = item.page.icon}
					<Command.Item
						value={item.key}
						keywords={[item.page.label, item.page.hint, item.page.href]}
						onSelect={() => void go(item)}
					>
						<Icon class="size-4" />
						<span>{item.page.label}</span>
						<Command.Shortcut>{item.page.hint}</Command.Shortcut>
					</Command.Item>
				{/each}
			</Command.Group>
		{/if}
		{#if userItems.length}
			<Command.Group heading="Users">
				{#each userItems as item (item.key)}
					<Command.Item
						value={item.key}
						keywords={[item.user.username, String(item.user.id)]}
						onSelect={() => void go(item)}
					>
						<UserIcon class="size-4" />
						<span>{item.user.username}</span>
						<Command.Shortcut>#{item.user.id}</Command.Shortcut>
					</Command.Item>
				{/each}
			</Command.Group>
		{/if}
	</Command.List>
</Command.Dialog>
