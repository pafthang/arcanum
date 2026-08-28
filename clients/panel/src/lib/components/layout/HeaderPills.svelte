<script lang="ts">
	import type { ComponentType } from 'svelte';

	let {
		value,
		options,
		onselect
	}: {
		value: string;
		options: { value: string; label?: string; icon?: ComponentType; title?: string }[];
		onselect: (value: string) => void;
	} = $props();
</script>

<div class="flex h-8 w-fit items-center rounded-lg border border-[var(--app-border)] bg-[var(--color-bg)] p-0.5">
	{#each options as option}
		{@const Icon = option.icon}
		<button
			type="button"
			title={option.title ?? option.label}
			aria-pressed={value === option.value}
			onclick={() => onselect(option.value)}
			class="inline-flex h-6 items-center gap-1.5 rounded-md text-xs transition-colors {option.label
				? 'px-2.5'
				: 'px-1.5'} {value === option.value
				? 'bg-[var(--color-bg-hover)] text-[var(--color-text-primary)]'
				: 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]'}"
		>
			{#if Icon}
				<Icon size={13} />
			{/if}
			{#if option.label}
				<span class={option.icon ? 'hidden sm:inline' : ''}>{option.label}</span>
			{/if}
		</button>
	{/each}
</div>
