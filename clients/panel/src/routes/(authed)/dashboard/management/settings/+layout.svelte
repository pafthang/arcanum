<script lang="ts">
	import type { Snippet } from 'svelte';
	import { contentWidth } from '$lib/features/layout/content-width.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div
	class="w-full {contentWidth.value === 'wide' ? 'max-w-none' : 'mx-auto max-w-2xl'}"
>
	{@render children()}
</div>
