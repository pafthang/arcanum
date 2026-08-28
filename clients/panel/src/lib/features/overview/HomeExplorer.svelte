<script lang="ts">
	import {
		Activity,
		Globe,
		Hash,
		Radio,
		Server,
		SlidersHorizontal,
		Users
	} from 'lucide-svelte';
	import * as Select from '$lib/components/ui/select';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import MetricCards from './MetricCards.svelte';
	import SectionHeading from './SectionHeading.svelte';
	import StatBarChart from './StatBarChart.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { formatBytes, formatNumber, parsePrettyBytes } from '$lib/format';
	import { rw } from '$lib/rw';

	type Measure = 'users' | 'online' | 'bandwidth' | 'nodes' | 'recap' | 'http';
	type BandwidthWindow = { current?: string } | undefined;

	let {
		stats,
		bandwidth,
		nodeDays = []
	}: {
		stats?: {
			users?: { statusCounts?: Record<string, number> };
			onlineStats?: { onlineNow?: number; lastDay?: number; lastWeek?: number; neverOnline?: number };
		} | null;
		bandwidth?: {
			bandwidthLastTwoDays?: BandwidthWindow;
			bandwidthLastSevenDays?: BandwidthWindow;
			bandwidthLast30Days?: BandwidthWindow;
			bandwidthCalendarMonth?: BandwidthWindow;
			bandwidthCurrentYear?: BandwidthWindow;
		} | null;
		nodeDays?: { nodeName: string; date: string; totalBytes: string }[];
	} = $props();

	const MEASURES: { value: Measure; label: string; icon: typeof Users }[] = [
		{ value: 'users', label: 'Users', icon: Users },
		{ value: 'online', label: 'Online', icon: Radio },
		{ value: 'bandwidth', label: 'Bandwidth', icon: Activity },
		{ value: 'nodes', label: 'Nodes', icon: Server },
		{ value: 'recap', label: 'Recap', icon: Hash },
		{ value: 'http', label: 'HTTP routes', icon: Globe }
	];

	let measure = $state<Measure>('users');
	const selected = $derived(MEASURES.find((item) => item.value === measure) ?? MEASURES[0]);
	const MeasureIcon = $derived(selected.icon);

	const recapQ = rw.system.recap();
	const httpQ = rw.system.httpStats();

	const recap = $derived(
		recapQ.data as {
			thisMonth?: { users?: number; traffic?: string };
			total?: {
				users?: number;
				nodes?: number;
				traffic?: string;
				nodesRam?: string;
				nodesCpuCores?: number;
				distinctCountries?: number;
			};
			version?: string;
		} | null
	);
	const httpRoutes = $derived(
		((httpQ.data as { routes?: { method?: string; route?: string; count?: number }[]; total?: number } | null)
			?.routes ?? []) as { method?: string; route?: string; count?: number }[]
	);

	const chartItems = $derived.by(() => {
		if (measure === 'users') {
			return Object.entries(stats?.users?.statusCounts ?? {})
				.map(([name, value]) => ({ name, value: Number(value) || 0 }))
				.filter((item) => item.value > 0)
				.sort((a, b) => b.value - a.value);
		}
		if (measure === 'online') {
			const o = stats?.onlineStats;
			return [
				{ name: 'Now', value: o?.onlineNow ?? 0 },
				{ name: '24h', value: o?.lastDay ?? 0 },
				{ name: '7d', value: o?.lastWeek ?? 0 },
				{ name: 'Never', value: o?.neverOnline ?? 0 }
			];
		}
		if (measure === 'bandwidth') {
			const b = bandwidth;
			return [
				{ name: '2 days', value: parsePrettyBytes(b?.bandwidthLastTwoDays?.current) },
				{ name: '7 days', value: parsePrettyBytes(b?.bandwidthLastSevenDays?.current) },
				{ name: '30 days', value: parsePrettyBytes(b?.bandwidthLast30Days?.current) },
				{ name: 'Month', value: parsePrettyBytes(b?.bandwidthCalendarMonth?.current) },
				{ name: 'Year', value: parsePrettyBytes(b?.bandwidthCurrentYear?.current) }
			];
		}
		if (measure === 'nodes') {
			const totals = new Map<string, number>();
			for (const row of nodeDays) {
				totals.set(row.nodeName, (totals.get(row.nodeName) ?? 0) + (Number(row.totalBytes) || 0));
			}
			return [...totals.entries()]
				.map(([name, value]) => ({ name, value }))
				.sort((a, b) => b.value - a.value)
				.slice(0, 12);
		}
		if (measure === 'http') {
			return httpRoutes
				.map((row) => ({
					name: `${row.method ?? 'GET'} ${row.route ?? ''}`.trim(),
					value: row.count ?? 0
				}))
				.sort((a, b) => b.value - a.value)
				.slice(0, 12);
		}
		return [];
	});

	const recapCards = $derived([
		{ label: 'Users this month', value: formatNumber(recap?.thisMonth?.users) },
		{ label: 'Traffic this month', value: recap?.thisMonth?.traffic || '—' },
		{ label: 'Users total', value: formatNumber(recap?.total?.users) },
		{ label: 'Nodes', value: formatNumber(recap?.total?.nodes) },
		{ label: 'Lifetime traffic', value: recap?.total?.traffic || '—' },
		{ label: 'Node RAM', value: recap?.total?.nodesRam || '—' },
		{ label: 'Node CPU cores', value: formatNumber(recap?.total?.nodesCpuCores) },
		{ label: 'Countries', value: formatNumber(recap?.total?.distinctCountries) }
	]);

	const loading = $derived(
		(measure === 'recap' && recapQ.loading && !recapQ.data) ||
			(measure === 'http' && httpQ.loading && !httpQ.data)
	);
	const error = $derived(
		measure === 'recap' ? recapQ.error : measure === 'http' ? httpQ.error : null
	);

	const groupedBy = $derived(
		measure === 'users'
			? 'status'
			: measure === 'online'
				? 'window'
				: measure === 'bandwidth'
					? 'period'
					: measure === 'nodes'
						? 'node'
						: measure === 'http'
							? 'route'
							: 'totals'
	);
</script>

<section>
	<SectionHeading
		icon={SlidersHorizontal}
		title="Build an insight"
		description="Pick a measure and see how it breaks down across this panel."
	/>

	<div class="mb-4 flex flex-wrap items-end gap-4 {chrome.tile} px-4 py-3.5">
		<div class={chrome.field}>
			<span class={chrome.label}>Measure</span>
			<Select.Root type="single" value={measure} onValueChange={(value) => value && (measure = value as Measure)}>
				<Select.Trigger size="sm" aria-label="Measure" class="w-[170px] bg-[var(--color-bg)]">
					<MeasureIcon size={13} />
					{selected.label}
				</Select.Trigger>
				<Select.Content>
					{#each MEASURES as item}
						{@const Icon = item.icon}
						<Select.Item value={item.value}><Icon size={13} />{item.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>
		<p class="pb-0.5 text-[13px] leading-5 text-[var(--color-text-tertiary)]">
			Grouped by <span class="text-[var(--color-text-secondary)]">{groupedBy}</span>
		</p>
	</div>

	{#if loading}
		<LoadingState />
	{:else if error}
		<ErrorState
			message={error.message}
			onretry={() => (measure === 'recap' ? recapQ.refetch() : httpQ.refetch())}
		/>
	{:else if measure === 'recap'}
		<MetricCards cards={recapCards} />
	{:else if chartItems.length === 0}
		<p class="text-sm text-[var(--color-text-tertiary)]">No data for this measure.</p>
	{:else}
		<div class="mb-3 flex gap-4 text-xs text-[var(--color-text-secondary)]">
			<span>
				Series
				<strong class="text-[var(--color-text-primary)]">{formatNumber(chartItems.length)}</strong>
			</span>
			{#if measure === 'nodes'}
				<span>
					Traffic
					<strong class="text-[var(--color-text-primary)]"
						>{formatBytes(chartItems.reduce((sum, item) => sum + item.value, 0))}</strong
					>
				</span>
			{/if}
		</div>
		<StatBarChart items={chartItems} />
	{/if}
</section>
