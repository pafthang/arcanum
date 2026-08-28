<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import type { Snippet } from 'svelte';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';
	import { RowsPerPage, TableSearch } from './index';
	import TableStatusBar from './table-status-bar.svelte';
	import { cn } from '$lib/utils.js';

	type Props = {
		table: TableHandlerInterface<Row>;
		children: Snippet;
		basic?: boolean;
		headless?: boolean;
		header?: Snippet;
		footer?: Snippet;
	};

	let { table, children, basic = false, headless = false, header, footer }: Props = $props();

	// svelte-ignore state_referenced_locally
	const bound = table as TableHandlerInterface<Row> & { __scrollBound?: boolean };
	if (!bound.__scrollBound) {
		bound.__scrollBound = true;
		// svelte-ignore state_referenced_locally
		const host = table;
		host.on('change', () => {
			if (host.element) host.element.scrollTop = 0;
		});
	}
</script>

<section
	bind:clientWidth={table.clientWidth}
	class={cn('flex h-full min-h-0 flex-col bg-inherit', !headless && 'overflow-hidden')}
>
	{#if header || basic}
		<header class="flex w-full flex-wrap items-center justify-between gap-2">
			{#if header}
				{@render header()}
			{:else}
				<TableSearch {table} />
				<RowsPerPage {table} />
			{/if}
		</header>
	{/if}

	<div
		class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-md border border-[var(--app-border)] bg-[var(--color-bg)]"
	>
		<article bind:this={table.element} class="min-h-0 flex-1 overflow-auto bg-inherit">
			{@render children()}
		</article>
		{#if footer}
			{@render footer()}
		{:else if basic}
			<TableStatusBar {table} />
		{/if}
	</div>
</section>
