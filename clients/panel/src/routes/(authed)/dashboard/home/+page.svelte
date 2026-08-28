<script lang="ts">
	import {
		Activity,
		ChartNoAxesCombined,
		Cpu,
		HeartPulse,
		TrendingUp
	} from 'lucide-svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import { page } from '$app/state';
	import MetricCards from '$lib/features/overview/MetricCards.svelte';
	import SectionHeading from '$lib/features/overview/SectionHeading.svelte';
	import StatBarChart from '$lib/features/overview/StatBarChart.svelte';
	import StatLineChart from '$lib/features/overview/StatLineChart.svelte';
	import HomeExplorer from '$lib/features/overview/HomeExplorer.svelte';
	import { bandwidthDelta, formatBytes, formatNumber, formatUptime } from '$lib/format';
	import { rw } from '$lib/rw';

	const stats = rw.system.stats();
	const bandwidth = rw.system.bandwidth();
	const health = rw.system.health();
	const meta = rw.system.metadata();
	const nodesStats = rw.system.nodesStatistics();

	type RuntimeMetric = {
		rss?: number;
		heapUsed?: number;
		heapTotal?: number;
		eventLoopDelayMs?: number;
		uptime?: number;
		pid?: number;
		instanceType?: string;
		instanceId?: string;
	};

	const runtime = $derived(
		((health.data as { runtimeMetrics?: RuntimeMetric[] } | null)?.runtimeMetrics ?? []) as RuntimeMetric[]
	);

	const tab = $derived(page.url.searchParams.get('tab') === 'explore' ? 'explore' : 'overview');

	const primaryCards = $derived.by(() => {
		const s = stats.data;
		if (!s) return [];
		return [
			{ label: 'Users', value: formatNumber(s.users?.totalUsers), href: '/dashboard/management/users' },
			{ label: 'Online now', value: formatNumber(s.onlineStats?.onlineNow), href: '/dashboard/management/users' },
			{ label: 'Online 24h', value: formatNumber(s.onlineStats?.lastDay) },
			{ label: 'Nodes online', value: formatNumber(s.nodes?.totalOnline), href: '/dashboard/management/nodes' },
			{ label: 'Lifetime traffic', value: formatBytes(s.nodes?.totalBytesLifetime) }
		];
	});

	const secondaryCards = $derived.by(() => {
		const s = stats.data;
		if (!s) return [];
		const memPct = s.memory?.total ? Math.round((s.memory.used / s.memory.total) * 100) : null;
		return [
			{ label: 'Online 7d', value: formatNumber(s.onlineStats?.lastWeek) },
			{ label: 'Never online', value: formatNumber(s.onlineStats?.neverOnline) },
			{ label: 'CPU cores', value: formatNumber(s.cpu?.cores) },
			{
				label: 'Memory used',
				value: formatBytes(s.memory?.used),
				sub: memPct != null ? `${memPct}% of ${formatBytes(s.memory?.total)}` : undefined
			},
			{ label: 'Uptime', value: formatUptime(s.uptime) }
		];
	});

	const statusItems = $derived.by(() => {
		const counts = stats.data?.users?.statusCounts ?? {};
		return Object.entries(counts)
			.map(([name, value]) => ({ name, value: Number(value) || 0 }))
			.filter((item) => item.value > 0)
			.sort((a, b) => b.value - a.value);
	});

	const bandwidthCards = $derived.by(() => {
		const b = bandwidth.data;
		if (!b) return [];
		const rows = [
			{ label: 'Last 2 days', stat: b.bandwidthLastTwoDays },
			{ label: 'Last 7 days', stat: b.bandwidthLastSevenDays },
			{ label: 'Last 30 days', stat: b.bandwidthLast30Days },
			{ label: 'Calendar month', stat: b.bandwidthCalendarMonth },
			{ label: 'Current year', stat: b.bandwidthCurrentYear }
		];
		return rows.map(({ label, stat }) => ({
			label,
			value: stat?.current || '0 B',
			sub: stat?.previous ? `prev ${stat.previous}` : undefined,
			trend: bandwidthDelta(stat?.current, stat?.previous)
		}));
	});

	type NodeDay = { nodeName: string; date: string; totalBytes: string };
	const nodeDays = $derived(
		((nodesStats.data as { lastSevenDays?: NodeDay[] } | null)?.lastSevenDays ?? []) as NodeDay[]
	);

	const nodeTrend = $derived.by(() => {
		const days = nodeDays;
		if (days.length === 0) return { categories: [] as string[], series: [] as { name: string; data: number[] }[] };
		const dates = [...new Set(days.map((d) => d.date))].sort();
		const totals = dates.map((date) =>
			days
				.filter((d) => d.date === date)
				.reduce((sum, d) => sum + (Number(d.totalBytes) || 0), 0)
		);
		return {
			categories: dates.map((date) => {
				const parsed = new Date(`${date}T00:00:00`);
				return Number.isNaN(parsed.getTime())
					? date
					: parsed.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
			}),
			series: [{ name: 'Node traffic', data: totals }]
		};
	});

	$effect(() => {
		const tick = () => {
			if (typeof document !== 'undefined' && document.hidden) return;
			void stats.refetch();
			void bandwidth.refetch();
			void nodesStats.refetch();
			void health.refetch();
		};
		const id = setInterval(tick, 30_000);
		return () => clearInterval(id);
	});

	const runtimeCards = $derived(
		runtime.map((metric) => ({
			label: metric.instanceType || metric.instanceId || `PID ${metric.pid ?? '—'}`,
			value: formatBytes(metric.heapUsed),
			sub: [
				metric.uptime != null ? `up ${formatUptime(metric.uptime)}` : null,
				metric.eventLoopDelayMs != null ? `loop ${metric.eventLoopDelayMs.toFixed(1)}ms` : null
			]
				.filter(Boolean)
				.join(' · ')
		}))
	);
</script>

{#if tab === 'explore'}
	<HomeExplorer stats={stats.data ?? null} bandwidth={bandwidth.data ?? null} {nodeDays} />
{:else if stats.loading && !stats.data}
	<div class="space-y-8">
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
			{#each Array.from({ length: 5 }, (_, i) => i) as card (card)}
				<div class="h-[88px] animate-pulse rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/40"></div>
			{/each}
		</div>
		<div class="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
			{#each Array.from({ length: 5 }, (_, i) => i) as card (card)}
				<div class="h-[72px] animate-pulse rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/40"></div>
			{/each}
		</div>
	</div>
{:else if stats.error && !stats.data}
	<ErrorState message={stats.error.message} onretry={() => stats.refetch()} />
{:else}
	<div class="space-y-8">
		<section>
		<SectionHeading
			icon={Activity}
			title="Current state"
			description="Live users, nodes and traffic. Cards with links open the matching list."
		>
			{#snippet actions()}
				<p class="text-[11px] text-[var(--color-text-tertiary)]">Live · refresh 30s</p>
			{/snippet}
		</SectionHeading>
		<MetricCards cards={primaryCards} />
		<div class="mt-3">
			<MetricCards cards={secondaryCards} compact />
		</div>
	</section>

	{#if statusItems.length > 0}
		<section>
			<SectionHeading
				icon={ChartNoAxesCombined}
				title="User distribution"
				description="Accounts grouped by current status."
			/>
			<StatBarChart items={statusItems} />
		</section>
	{/if}

	<section>
		<SectionHeading
			icon={TrendingUp}
			title="Bandwidth"
			description="Traffic compared with the previous window of the same length."
		/>
		{#if bandwidth.loading && !bandwidth.data}
			<LoadingState />
		{:else if bandwidth.error && !bandwidth.data}
			<ErrorState message={bandwidth.error.message} onretry={() => bandwidth.refetch()} />
		{:else}
			<MetricCards cards={bandwidthCards} />
		{/if}
	</section>

	{#if nodeTrend.categories.length > 0}
		<section>
			<SectionHeading
				icon={TrendingUp}
				title="Node traffic, last 7 days"
				description="Aggregated bytes across nodes, same shape as the insights delivery trend."
			/>
			<StatLineChart categories={nodeTrend.categories} series={nodeTrend.series} />
		</section>
	{/if}

	{#if runtimeCards.length > 0}
		<section>
			<SectionHeading
				icon={HeartPulse}
				title="Runtime"
				description="Process health reported by the API instances."
			/>
			<MetricCards cards={runtimeCards} />
		</section>
	{:else if meta.data}
		<section>
			<SectionHeading icon={Cpu} title="Runtime" description="Build metadata for this instance." />
			<MetricCards
				cards={[
					{ label: 'Version', value: meta.data.version || '—' },
					{ label: 'Build', value: meta.data.build?.number || '—' },
					{ label: 'Branch', value: meta.data.git?.backend?.branch || '—' },
					{
						label: 'Commit',
						value: meta.data.git?.backend?.commitSha?.slice(0, 7) || '—'
					}
				]}
			/>
		</section>
	{/if}
	</div>
{/if}
