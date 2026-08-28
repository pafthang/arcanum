<script lang="ts">
	import { onMount } from 'svelte';
	import { ChevronDown, Search } from 'lucide-svelte';
	import { Input } from '$lib/components/ui/input';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import * as Kbd from '$lib/components/ui/kbd';
	import * as Sheet from '$lib/components/ui/sheet';
	import {
		bindShortcutSheet,
		filterShortcutGroups,
		shortcutsUi
	} from '$lib/features/layout/shortcuts.svelte';

	let search = $state('');
	let collapsed = $state<string[]>([]);
	const groups = $derived(filterShortcutGroups(search));

	$effect(() => {
		if (!shortcutsUi.open) search = '';
	});

	onMount(() => bindShortcutSheet());
</script>

<Sheet.Root bind:open={shortcutsUi.open}>
	<Sheet.Content
		side="right"
		class="gap-0 overflow-y-auto border-[var(--app-border)] bg-[var(--color-bg)] p-0 sm:max-w-sm"
	>
		<Sheet.Header class="px-0">
			<Sheet.Title class="px-5">Shortcuts</Sheet.Title>
			<Sheet.Description class="sr-only">Search and browse keyboard shortcuts.</Sheet.Description>
			<div class="px-5 pt-1">
				<div class="relative">
					<Search
						size={14}
						class="pointer-events-none absolute top-1/2 left-2.5 -translate-y-1/2 text-[var(--color-text-tertiary)]"
					/>
					<Input
						class="pl-8"
						placeholder="Search shortcuts"
						autocomplete="off"
						autocorrect="off"
						spellcheck={false}
						bind:value={search}
					/>
				</div>
			</div>
		</Sheet.Header>

		<div class="pb-4">
			{#each groups as group (group.name)}
				{@const open = !collapsed.includes(group.name)}
				<div class="px-5 py-3">
					<Collapsible.Root
						{open}
						onOpenChange={(next) => {
							collapsed = next
								? collapsed.filter((name) => name !== group.name)
								: [...collapsed, group.name];
						}}
					>
						<Collapsible.Trigger
							class="group flex w-full items-center gap-2 text-sm text-[var(--color-text-primary)]"
						>
							<ChevronDown
								size={14}
								class="text-[var(--color-text-tertiary)] transition-transform group-hover:text-[var(--color-text-primary)] {open
									? ''
									: '-rotate-90'}"
							/>
							{group.name}
						</Collapsible.Trigger>
						<Collapsible.Content class="space-y-3 pt-3 text-sm text-[var(--color-text-secondary)]">
							{#each group.shortcuts as shortcut (shortcut.id)}
								<div class="flex items-center justify-between gap-2">
									<span>{shortcut.title}</span>
									<Kbd.Root class="font-mono tracking-widest">{shortcut.keys}</Kbd.Root>
								</div>
							{/each}
						</Collapsible.Content>
					</Collapsible.Root>
				</div>
			{/each}

			{#if groups.length === 0}
				<div class="flex flex-col items-center gap-1 pt-14 text-[var(--color-text-tertiary)]">
					<Search size={18} />
					<p class="text-sm">No shortcuts found</p>
				</div>
			{/if}
		</div>
	</Sheet.Content>
</Sheet.Root>
