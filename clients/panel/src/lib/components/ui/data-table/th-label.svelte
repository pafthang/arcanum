<script lang="ts">
	import { cn, type WithElementRef } from '$lib/utils.js';
	import type { HTMLThAttributes } from 'svelte/elements';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: WithElementRef<HTMLThAttributes> = $props();
</script>

<th
	bind:this={ref}
	scope="col"
	class={cn(
		'h-9 px-3 py-2 text-left align-middle text-[11px] font-medium tracking-wide text-[var(--color-text-tertiary)] uppercase',
		className
	)}
	{...restProps}
>
	{@render children?.()}
</th>
