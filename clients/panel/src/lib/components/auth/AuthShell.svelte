<script lang="ts">
	import type { Snippet } from 'svelte';
	import BrandMark from '$lib/components/brand/BrandMark.svelte';
	import { APP_NAME } from '$lib/brand';

	let {
		title = APP_NAME,
		subtitle,
		children
	}: {
		title?: string;
		subtitle?: string;
		children: Snippet;
	} = $props();
</script>

<div
	class="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-[var(--color-bg)] px-4 py-10"
>
	<div
		class="pointer-events-none absolute inset-0 opacity-40"
		style="background:
			radial-gradient(ellipse 70% 50% at 50% -10%, color-mix(in srgb, var(--app-accent) 22%, transparent), transparent 60%),
			radial-gradient(ellipse 40% 30% at 80% 110%, color-mix(in srgb, var(--app-accent) 12%, transparent), transparent 55%);"
	></div>
	<div
		class="pointer-events-none absolute inset-0 opacity-[0.07]"
		style="background-image: linear-gradient(var(--app-border) 1px, transparent 1px), linear-gradient(90deg, var(--app-border) 1px, transparent 1px); background-size: 48px 48px;"
	></div>

	<div class="relative z-10 flex w-full max-w-[420px] flex-col items-center">
		<BrandMark size="lg" {title} />
		{#if subtitle}
			<p class="mt-3 text-center text-sm text-[var(--color-text-secondary)]">{subtitle}</p>
		{/if}
		<div
			class="mt-8 w-full rounded-xl border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/60 p-6 shadow-[0_24px_80px_rgba(0,0,0,0.35)] backdrop-blur-sm"
		>
			{@render children()}
		</div>
	</div>
</div>
