<script lang="ts">
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import { Save } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';

	export type SplitActionItem = {
		id: string;
		label: string;
		disabled?: boolean;
		variant?: 'default' | 'destructive';
	};

	let {
		items,
		value,
		pending = false,
		onselect,
		onrun
	}: {
		items: SplitActionItem[];
		value: string;
		pending?: boolean;
		onselect: (id: string) => void;
		onrun: () => void | Promise<void>;
	} = $props();

	const current = $derived(items.find((item) => item.id === value) ?? items[0]);
	const split = $derived(items.length > 1);
	const saveAction = $derived(current?.id === 'save');
	const iconOnly = $derived(saveAction && !pending);
	const leftClass = $derived(
		[
			'h-8',
			iconOnly ? 'w-8 px-0' : '',
			split ? 'rounded-none rounded-l-lg' : ''
		]
			.filter(Boolean)
			.join(' ')
	);
</script>

{#if current}
	<div class="inline-flex h-8 overflow-hidden rounded-lg">
		<Button
			class={leftClass}
			size={iconOnly ? 'icon-sm' : 'default'}
			variant={current.variant === 'destructive' ? 'destructive' : 'default'}
			disabled={pending || current.disabled}
			title={saveAction ? 'Save' : current.label}
			aria-label={saveAction ? 'Save' : current.label}
			onclick={() => void onrun()}
		>
			{#if pending}
				Working…
			{:else if saveAction}
				<Save class="size-4" />
			{:else}
				{current.label}
			{/if}
		</Button>
		{#if split}
			<DropdownMenu.Root>
				<DropdownMenu.Trigger>
					{#snippet child({ props })}
						<Button
							{...props}
							size="icon-sm"
							variant={current.variant === 'destructive' ? 'destructive' : 'default'}
							class="h-8 w-8 rounded-none rounded-r-lg border-l border-[color-mix(in_srgb,white_18%,transparent)] px-0"
							disabled={pending}
							aria-label="Change action"
						>
							<ChevronUpIcon class="size-4" />
						</Button>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content align="end" side="top" class="min-w-44">
					{#each items as item (item.id)}
						<DropdownMenu.Item
							disabled={item.disabled}
							onclick={() => onselect(item.id)}
							class={item.id === current.id ? 'font-medium' : ''}
						>
							{item.label}
						</DropdownMenu.Item>
					{/each}
				</DropdownMenu.Content>
			</DropdownMenu.Root>
		{/if}
	</div>
{/if}
