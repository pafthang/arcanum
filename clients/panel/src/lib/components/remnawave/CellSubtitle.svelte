<script lang="ts">
	let {
		title,
		children
	}: {
		title?: string;
		children?: import('svelte').Snippet;
	} = $props();
</script>

<div
	class="w-0 min-w-full truncate text-[11px] leading-tight text-[var(--color-text-tertiary)]"
	{title}
>
	{@render children?.()}
</div>
