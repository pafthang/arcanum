<script lang="ts">
	import BrandMark from '$lib/components/brand/BrandMark.svelte';

	let { message = 'Loading panel…' }: { message?: string } = $props();
</script>

<div class="flex min-h-dvh flex-col items-center justify-center bg-[var(--color-bg)] px-6">
	<BrandMark size="lg" />
	<p class="mt-5 text-sm text-[var(--color-text-secondary)]">{message}</p>
	<div class="mt-4 h-1 w-48 overflow-hidden rounded-full bg-[var(--color-bg-tertiary)]">
		<div class="boot-bar h-full w-1/2 rounded-full bg-[var(--app-accent)]"></div>
	</div>
</div>

<style>
	.boot-bar {
		animation: boot-slide 1.1s ease-in-out infinite;
	}
	@keyframes boot-slide {
		0% {
			transform: translateX(-120%);
		}
		100% {
			transform: translateX(220%);
		}
	}
</style>
