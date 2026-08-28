<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as Sheet from '$lib/components/ui/sheet';
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
		onOpenChange,
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
		onOpenChange?: (open: boolean) => void;
		onrun: () => void | Promise<void>;
	} = $props();

	function close() {
		open = false;
		onOpenChange?.(false);
	}
</script>

<Sheet.Root
	{open}
	onOpenChange={(value) => {
		open = value;
		onOpenChange?.(value);
	}}
>
	<Sheet.Content
		side="right"
		showCloseButton={false}
		class="gap-0 border-[var(--app-border)] bg-[var(--color-bg)] p-0 shadow-[-24px_0_80px_rgba(0,0,0,0.35)] sm:max-w-[420px]"
	>
		<RecordFrame
			{title}
			{description}
			{menu}
			{actions}
			bind:actionId
			{pending}
			onclose={close}
			{onrun}
		>
			{#if children}
				{@render children()}
			{/if}
		</RecordFrame>
	</Sheet.Content>
</Sheet.Root>
