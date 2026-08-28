<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import {
		getAnalyticsChartTheme,
		observeAnalyticsTheme,
		seriesChartColor
	} from '$lib/features/analytics/chart-theme';

	let {
		items
	}: {
		items: { name: string; value: number; color?: string }[];
	} = $props();

	let container: HTMLDivElement | undefined = $state();
	let chart: echarts.ECharts | null = null;

	function tooltipTheme(theme: ReturnType<typeof getAnalyticsChartTheme>) {
		return {
			trigger: 'axis' as const,
			backgroundColor: theme.background,
			borderColor: theme.border,
			borderWidth: 1,
			borderRadius: 8,
			padding: [6, 10] as [number, number],
			axisPointer: { lineStyle: { color: theme.textTertiary, type: 'dotted' as const, opacity: 0.45 } },
			textStyle: { color: theme.textPrimary, fontSize: 11 }
		};
	}

	function renderChart() {
		if (!container || !chart) return;
		if (items.length === 0) {
			chart.clear();
			return;
		}
		const theme = getAnalyticsChartTheme();
		const names = items.map((item) => item.name);
		const counts = items.map((item) => item.value);
		const colors = items.map((item, i) => item.color ?? seriesChartColor(i, theme));
		chart.setOption(
			{
				backgroundColor: 'transparent',
				animationDuration: 250,
				tooltip: tooltipTheme(theme),
				grid: { left: 12, right: 12, top: 16, bottom: 12, containLabel: true },
				xAxis: {
					type: 'category',
					data: names,
					axisLabel: {
						fontSize: 10,
						color: theme.textTertiary,
						interval: 0,
						rotate: names.length > 7 ? 25 : 0
					},
					axisTick: { show: false },
					axisLine: { lineStyle: { color: theme.border } }
				},
				yAxis: {
					type: 'value',
					minInterval: 1,
					axisLine: { show: false },
					axisTick: { show: false },
					axisLabel: { fontSize: 10, color: theme.textTertiary },
					splitLine: { lineStyle: { color: theme.border, opacity: 0.35 } }
				},
				series: [
					{
						type: 'bar',
						data: counts.map((val, i) => ({
							value: val,
							itemStyle: { color: colors[i], borderRadius: [4, 4, 0, 0] }
						})),
						barMaxWidth: 36
					}
				]
			},
			true
		);
	}

	$effect(() => {
		items;
		renderChart();
	});

	onMount(() => {
		if (!container) return;
		chart = echarts.init(container, undefined, { renderer: 'canvas' });
		renderChart();
		const resizeObserver = new ResizeObserver(() => chart?.resize());
		resizeObserver.observe(container);
		const stopTheme = observeAnalyticsTheme(renderChart);
		return () => {
			resizeObserver.disconnect();
			stopTheme();
			chart?.dispose();
			chart = null;
		};
	});
</script>

<div class="overflow-hidden rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/40">
	<div bind:this={container} class="h-60 w-full"></div>
</div>
