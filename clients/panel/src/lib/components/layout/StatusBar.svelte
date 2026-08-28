<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Activity,
		AlignHorizontalSpaceAround,
		Server,
		UnfoldHorizontal,
		Users,
		Zap
	} from 'lucide-svelte';
	import ShortcutsSheet from './ShortcutsSheet.svelte';
	import { formatNumber, formatUptime } from '$lib/format';
	import { contentWidth } from '$lib/features/layout/content-width.svelte';
	import { shortcutsChord, shortcutsUi } from '$lib/features/layout/shortcuts.svelte';
	import { rw } from '$lib/rw';

	const stats = rw.system.stats();
	const health = rw.system.health();
	const meta = rw.system.metadata();

	const healthy = $derived(!health.error && Boolean(health.data || stats.data));
	const healthLabel = $derived(
		health.error ? 'API unreachable' : health.loading && !health.data ? 'Connecting' : 'Healthy'
	);
	const version = $derived(meta.data?.version || '—');
	const branch = $derived(meta.data?.git?.backend?.branch || '');
	const users = $derived(stats.data?.users?.totalUsers);
	const online = $derived(stats.data?.onlineStats?.onlineNow);
	const nodes = $derived(stats.data?.nodes?.totalOnline);
	const uptime = $derived(stats.data?.uptime);

	onMount(() => {
		const tick = () => {
			if (document.hidden) return;
			void stats.refetch();
			void health.refetch();
		};
		const id = setInterval(tick, 30_000);
		return () => clearInterval(id);
	});
</script>

<footer
	data-slot="status-bar"
	class="flex h-6 shrink-0 items-stretch justify-between gap-2 overflow-hidden border-t border-[var(--app-border)] bg-[var(--color-bg-secondary)] text-[11px] text-[var(--color-text-secondary)]"
>
	<div class="flex min-w-0 items-stretch">
		<span
			class="inline-flex items-center gap-1.5 px-2.5"
			title={health.error?.message || healthLabel}
		>
			<span
				class="size-1.5 shrink-0 rounded-full {healthy
					? 'bg-[var(--color-success)]'
					: health.loading
						? 'bg-[var(--color-warning)]'
						: 'bg-[var(--color-error)]'}"
			></span>
			<span class="truncate">{healthLabel}</span>
		</span>
		<a
			href="/dashboard/home"
			title="Panel version"
			class="inline-flex items-center px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
		>
			{version}
		</a>
		{#if branch}
			<span class="hidden items-center px-2 sm:inline-flex" title="Backend branch">{branch}</span>
		{/if}
	</div>
	<div class="flex shrink-0 items-stretch">
		<a
			href="/dashboard/management/users"
			title="Total users"
			class="inline-flex items-center gap-1 px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
		>
			<Users size={12} />
			{users == null ? '—' : formatNumber(users)}
			<span class="hidden sm:inline">users</span>
		</a>
		<a
			href="/dashboard/management/users"
			title="Online now"
			class="inline-flex items-center gap-1 px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
		>
			<Activity size={12} />
			{online == null ? '—' : formatNumber(online)}
			<span class="hidden sm:inline">online</span>
		</a>
		<a
			href="/dashboard/management/nodes"
			title="Nodes online"
			class="inline-flex items-center gap-1 px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
		>
			<Server size={12} />
			{nodes == null ? '—' : formatNumber(nodes)}
			<span class="hidden sm:inline">nodes</span>
		</a>
		<span class="hidden items-center px-2.5 sm:inline-flex" title="Panel uptime">
			{formatUptime(uptime)}
		</span>
		<span class="mx-1 w-px self-stretch bg-[var(--app-border)]" aria-hidden="true"></span>
		<button
			type="button"
			title="Centered content"
			aria-pressed={contentWidth.value === 'contained'}
			onclick={() => contentWidth.set('contained')}
			class="inline-flex items-center px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] {contentWidth.value ===
			'contained'
				? 'text-[var(--color-text-primary)]'
				: ''}"
		>
			<AlignHorizontalSpaceAround size={12} />
		</button>
		<button
			type="button"
			title="Full-width content"
			aria-pressed={contentWidth.value === 'wide'}
			onclick={() => contentWidth.set('wide')}
			class="inline-flex items-center px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] {contentWidth.value ===
			'wide'
				? 'text-[var(--color-text-primary)]'
				: ''}"
		>
			<UnfoldHorizontal size={12} />
		</button>
		<span class="mx-1 w-px self-stretch bg-[var(--app-border)]" aria-hidden="true"></span>
		<button
			type="button"
			title="Shortcuts ({shortcutsChord()})"
			aria-label="Shortcuts"
			aria-pressed={shortcutsUi.open}
			onclick={() => shortcutsUi.toggle()}
			class="inline-flex items-center gap-1 px-2 hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] {shortcutsUi.open
				? 'text-[var(--color-text-primary)]'
				: ''}"
		>
			<Zap size={12} />
			<span class="hidden sm:inline">Shortcuts</span>
		</button>
	</div>
</footer>
<ShortcutsSheet />
