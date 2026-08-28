<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { chrome } from './chrome';
	import { cn } from '$lib/utils.js';
	import RecordFrame, { type RecordMenuItem } from './RecordFrame.svelte';
	import type { SplitActionItem } from './SplitAction.svelte';

	let {
		open = $bindable(false),
		title,
		description,
		menu = [],
		actions = [],
		actionId = $bindable(''),
		pending = false,
		children,
		onrun
	}: {
		open: boolean;
		title: string;
		description?: string;
		menu?: RecordMenuItem[];
		actions?: SplitActionItem[];
		actionId?: string;
		pending?: boolean;
		children?: Snippet;
		onrun?: () => void | Promise<void>;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Content
		showCloseButton={false}
		class={cn('flex max-h-[min(90vh,720px)] flex-col gap-0 overflow-hidden p-0 sm:max-w-[420px]', chrome.panel, 'rounded-xl')}
	>
		<RecordFrame
			{title}
			{description}
			{menu}
			{actions}
			bind:actionId
			{pending}
			hasBody={Boolean(children)}
			onclose={() => (open = false)}
			onrun={() => onrun?.()}
		>
			{#if children}
				{@render children()}
			{/if}
		</RecordFrame>
	</Dialog.Content>
</Dialog.Root>
