export interface AnalyticsChartTheme {
	accent: string;
	accentLight: string;
	background: string;
	border: string;
	error: string;
	success: string;
	textPrimary: string;
	textSecondary: string;
	textTertiary: string;
	warning: string;
}

function cssColor(styles: CSSStyleDeclaration, name: string, fallback: string): string {
	return styles.getPropertyValue(name).trim() || fallback;
}

export function getAnalyticsChartTheme(): AnalyticsChartTheme {
	const styles = getComputedStyle(document.documentElement);
	return {
		accent: cssColor(styles, '--app-accent', '#6650eb'),
		accentLight: cssColor(styles, '--app-accent-light', '#9585f5'),
		background: cssColor(styles, '--color-bg-secondary', '#242424'),
		border: cssColor(styles, '--app-border', '#3a3a3a'),
		error: cssColor(styles, '--color-error', '#ef4444'),
		success: cssColor(styles, '--color-success', '#22c55e'),
		textPrimary: cssColor(styles, '--color-text-primary', '#f5f5f5'),
		textSecondary: cssColor(styles, '--color-text-secondary', '#b8b8b8'),
		textTertiary: cssColor(styles, '--color-text-tertiary', '#858585'),
		warning: cssColor(styles, '--color-warning', '#f59e0b')
	};
}

export function seriesChartColor(index: number, theme: AnalyticsChartTheme): string {
	return [
		theme.accentLight,
		theme.warning,
		theme.success,
		theme.error,
		'#06b6d4',
		'#ec4899',
		theme.textSecondary
	][index % 7];
}

export function observeAnalyticsTheme(onchange: () => void): () => void {
	const observer = new MutationObserver(onchange);
	observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class', 'style'] });
	return () => observer.disconnect();
}
