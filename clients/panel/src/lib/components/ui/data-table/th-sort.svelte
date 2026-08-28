<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import ArrowUpDownIcon from '@lucide/svelte/icons/arrow-up-down';
	import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';
	import ChevronUpIcon from '@lucide/svelte/icons/chevron-up';
	import { cn, type WithoutChildren } from '$lib/utils.js';
	import type { Snippet } from 'svelte';
	import type { HTMLThAttributes } from 'svelte/elements';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';

	type Field<R extends import('@vincjo/datatables/server').Row> = Parameters<
		TableHandlerInterface<R>['createSort']
	>[0];

	type Props = WithoutChildren<HTMLThAttributes> & {
		table: TableHandlerInterface<Row>;
		field: Field<Row>;
		/** Apply this sort on mount. Omit to keep the column unsorted. */
		defaultDirection?: 'asc' | 'desc';
		children: Snippet;
		btnClass?: string;
	};

	let {
		table,
		field,
		defaultDirection,
		children,
		class: className,
		btnClass = '',
		...restProps
	}: Props = $props();

	// svelte-ignore state_referenced_locally
	const sort = table.createSort(field);
	// svelte-ignore state_referenced_locally
	const initialDirection = defaultDirection;
	if (initialDirection) sort.init(initialDirection);
</script>

<th
	scope="col"
	aria-sort={sort.isActive ? (sort.direction === 'asc' ? 'ascending' : 'descending') : 'none'}
	class={cn('h-9 border-b border-[var(--app-border)] bg-inherit p-0 text-left align-middle', className)}
	{...restProps}
>
	<button
		type="button"
		onclick={() => sort.set()}
		class={cn(
			'flex w-full cursor-pointer items-center gap-1 px-3 py-2 text-left text-[11px] font-medium tracking-wide text-[var(--color-text-tertiary)] uppercase outline-none select-none hover:text-[var(--color-text-secondary)] focus-visible:ring-2 focus-visible:ring-ring/50',
			sort.isActive && 'text-[var(--color-text-secondary)]',
			btnClass
		)}
	>
		<span class="truncate">
			{@render children()}
		</span>
		{#if sort.isActive && sort.direction === 'asc'}
			<ChevronUpIcon class="size-3 shrink-0 text-[var(--color-text-secondary)]" />
		{:else if sort.isActive && sort.direction === 'desc'}
			<ChevronDownIcon class="size-3 shrink-0 text-[var(--color-text-secondary)]" />
		{:else}
			<ArrowUpDownIcon class="size-3 shrink-0 opacity-40" />
		{/if}
	</button>
</th>
