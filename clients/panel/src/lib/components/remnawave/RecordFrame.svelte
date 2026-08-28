<script lang="ts">
	import type { Snippet } from 'svelte';
	import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { chrome } from './chrome';
	import SplitAction, { type SplitActionItem } from './SplitAction.svelte';

	export type RecordMenuItem = {
		label: string;
		onclick: () => void;
		variant?: 'destructive' | 'default';
	};

	let {
		title,
		description,
		menu = [],
		actions = [],
		actionId = $bindable(''),
		pending = false,
		closeLabel = 'Close',
		onclose,
		onrun,
		hasBody = true,
		children
	}: {
		title: string;
		description?: string;
		menu?: RecordMenuItem[];
		actions?: SplitActionItem[];
		actionId?: string;
		pending?: boolean;
		closeLabel?: string;
		onclose: () => void;
		onrun: () => void | Promise<void>;
		hasBody?: boolean;
		children?: Snippet;
	} = $props();

	const resolvedAction = $derived(actionId || actions[0]?.id || '');
</script>

<header class="flex shrink-0 items-start justify-between gap-3 border-b border-[var(--app-border)] px-6 py-4">
	<div class="min-w-0 pt-0.5">
		<h2 class={chrome.title}>{title}</h2>
		{#if description}
			<p class={chrome.description}>{description}</p>
		{/if}
	</div>
	{#if menu.length > 0}
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Button
						{...props}
						variant="ghost"
						size="icon-sm"
						class="shrink-0 rounded-full bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]"
						aria-label="More actions"
					>
						<EllipsisIcon class="size-4" />
					</Button>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content align="end" class="min-w-44">
				{#each menu as item (item.label)}
					<DropdownMenu.Item variant={item.variant} onclick={item.onclick}>{item.label}</DropdownMenu.Item>
				{/each}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	{/if}
</header>

{#if hasBody && children}
	<div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
		{@render children()}
	</div>
{/if}

<footer class="flex shrink-0 items-center justify-between gap-3 border-t border-[var(--app-border)] bg-[var(--color-bg)] px-6 py-3.5">
	<Button variant="ghost" class="px-2 font-medium" onclick={onclose}>{closeLabel}</Button>
	{#if actions.length > 0}
		<SplitAction
			items={actions}
			value={resolvedAction}
			{pending}
			onselect={(id) => (actionId = id)}
			onrun={onrun}
		/>
	{/if}
</footer>
