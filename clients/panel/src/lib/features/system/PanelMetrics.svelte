<script lang="ts">
	import { Cpu, HeartPulse, TrendingUp } from 'lucide-svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import MetricCards from '$lib/features/overview/MetricCards.svelte';
	import SectionHeading from '$lib/features/overview/SectionHeading.svelte';
	import { bandwidthDelta, formatBytes, formatNumber, formatUptime } from '$lib/format';
	import { rw } from '$lib/rw';

	const stats = rw.system.stats();
	const bandwidth = rw.system.bandwidth();
	const health = rw.system.health();
	const meta = rw.system.metadata();

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

	const panelCards = $derived.by(() => {
		const s = stats.data;
		if (!s) return [];
		const memPct = s.memory?.total ? Math.round((s.memory.used / s.memory.total) * 100) : null;
		return [
			{ label: 'CPU cores', value: formatNumber(s.cpu?.cores) },
			{
				label: 'Memory',
				value: formatBytes(s.memory?.used),
				sub: memPct != null ? `${memPct}% of ${formatBytes(s.memory?.total)}` : undefined
			},
			{ label: 'Uptime', value: formatUptime(s.uptime) },
			{ label: 'Online now', value: formatNumber(s.onlineStats?.onlineNow) },
			{ label: 'Nodes online', value: formatNumber(s.nodes?.totalOnline) }
		];
	});

	const bandwidthCards = $derived.by(() => {
		const b = bandwidth.data;
		if (!b) return [];
		return [
			{ label: 'Last 2 days', stat: b.bandwidthLastTwoDays },
			{ label: 'Last 7 days', stat: b.bandwidthLastSevenDays },
			{ label: 'Last 30 days', stat: b.bandwidthLast30Days },
			{ label: 'Calendar month', stat: b.bandwidthCalendarMonth },
			{ label: 'Current year', stat: b.bandwidthCurrentYear }
		].map(({ label, stat }) => ({
			label,
			value: stat?.current || '0 B',
			sub: stat?.previous ? `prev ${stat.previous}` : undefined,
			trend: bandwidthDelta(stat?.current, stat?.previous)
		}));
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

	$effect(() => {
		const tick = () => {
			if (document.hidden) return;
			void stats.refetch();
			void bandwidth.refetch();
			void health.refetch();
		};
		const id = setInterval(tick, 30_000);
		return () => clearInterval(id);
	});
</script>

{#if stats.loading && !stats.data}
	<LoadingState />
{:else if stats.error && !stats.data}
	<ErrorState message={stats.error.message} onretry={() => stats.refetch()} />
{:else}
	<div class="space-y-8">
		<section>
			<SectionHeading icon={Cpu} title="Panel" description="API host load. Refreshes every 30s." />
			<MetricCards cards={panelCards} />
		</section>
		<section>
			<SectionHeading icon={TrendingUp} title="Bandwidth" description="Traffic versus the previous window." />
			{#if bandwidth.loading && !bandwidth.data}
				<LoadingState />
			{:else if bandwidth.error && !bandwidth.data}
				<ErrorState message={bandwidth.error.message} onretry={() => bandwidth.refetch()} />
			{:else}
				<MetricCards cards={bandwidthCards} />
			{/if}
		</section>
		{#if runtimeCards.length > 0}
			<section>
				<SectionHeading icon={HeartPulse} title="Runtime" description="Process health from API instances." />
				<MetricCards cards={runtimeCards} />
			</section>
		{:else if meta.data}
			<section>
				<SectionHeading icon={Cpu} title="Build" description="Metadata for this instance." />
				<MetricCards
					cards={[
						{ label: 'Version', value: meta.data.version || '—' },
						{ label: 'Build', value: meta.data.build?.number || '—' },
						{ label: 'Branch', value: meta.data.git?.backend?.branch || '—' },
						{ label: 'Commit', value: meta.data.git?.backend?.commitSha?.slice(0, 7) || '—' }
					]}
				/>
			</section>
		{/if}
	</div>
{/if}
