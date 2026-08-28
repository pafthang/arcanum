<script lang="ts">
	import { onMount } from 'svelte';
	import * as echarts from 'echarts';
	import {
		getAnalyticsChartTheme,
		observeAnalyticsTheme,
		seriesChartColor
	} from '$lib/features/analytics/chart-theme';

	let {
		categories,
		series
	}: {
		categories: string[];
		series: { name: string; data: number[] }[];
	} = $props();

	let container: HTMLDivElement | undefined = $state();
	let chart: echarts.ECharts | null = null;

	function renderChart() {
		if (!container || !chart) return;
		if (categories.length === 0 || series.length === 0) {
			chart.clear();
			return;
		}
		const theme = getAnalyticsChartTheme();
		chart.setOption(
			{
				backgroundColor: 'transparent',
				animationDuration: 250,
				tooltip: {
					trigger: 'axis',
					backgroundColor: theme.background,
					borderColor: theme.border,
					borderWidth: 1,
					borderRadius: 8,
					padding: [6, 10],
					axisPointer: { lineStyle: { color: theme.textTertiary, type: 'dotted', opacity: 0.45 } },
					textStyle: { color: theme.textPrimary, fontSize: 11 }
				},
				legend: {
					data: series.map((item) => item.name),
					left: 12,
					top: 8,
					icon: 'circle',
					itemWidth: 8,
					itemHeight: 8,
					itemGap: 18,
					textStyle: { fontSize: 10, color: theme.textSecondary }
				},
				grid: { left: 12, right: 16, top: 44, bottom: 12, containLabel: true },
				xAxis: {
					type: 'category',
					data: categories,
					boundaryGap: false,
					axisLabel: {
						fontSize: 10,
						color: theme.textTertiary,
						rotate: categories.length > 12 ? 30 : 0,
						hideOverlap: true
					},
					axisTick: { show: false },
					axisLine: { lineStyle: { color: theme.border } }
				},
				yAxis: {
					type: 'value',
					axisLine: { show: false },
					axisTick: { show: false },
					axisLabel: { fontSize: 10, color: theme.textTertiary },
					splitLine: { lineStyle: { color: theme.border, opacity: 0.35 } }
				},
				series: series.map((item, i) => ({
					name: item.name,
					type: 'line',
					data: item.data,
					smooth: true,
					showSymbol: true,
					symbol: 'circle',
					symbolSize: 5,
					lineStyle: { width: 2, color: seriesChartColor(i, theme) },
					itemStyle: { color: seriesChartColor(i, theme) },
					areaStyle: { opacity: 0.08, color: seriesChartColor(i, theme) }
				}))
			},
			true
		);
	}

	$effect(() => {
		categories;
		series;
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
	<div bind:this={container} class="h-64 w-full"></div>
</div>
