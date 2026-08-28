<script lang="ts">
	import * as Select from '$lib/components/ui/select';

	let {
		value,
		options,
		onselect,
		embedded = false
	}: {
		value: string;
		options: { value: string; label: string }[];
		onselect: (value: string) => void;
		embedded?: boolean;
	} = $props();

	const current = $derived(options.find((option) => option.value === value)?.label ?? 'Filter');
</script>

<Select.Root
	type="single"
	{value}
	onValueChange={(next) => {
		if (!next || next === value) return;
		onselect(next);
	}}
>
	<Select.Trigger
		size="sm"
		class={embedded
			? 'h-7 max-w-28 gap-1 border-0 bg-transparent px-2 text-[11px] shadow-none hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] dark:bg-transparent dark:hover:bg-[var(--color-bg-hover)] [&_svg]:size-3 [&_svg]:opacity-50'
			: 'h-7 min-w-16 max-w-36 px-2 text-xs'}
		title="Filter"
		aria-label="Filter"
	>
		{current}
	</Select.Trigger>
	<Select.Content class="p-0.5" align="end" side={embedded ? 'top' : 'bottom'}>
		{#each options as option (option.value)}
			<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
		{/each}
	</Select.Content>
</Select.Root>
