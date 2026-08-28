<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import Columns2Icon from '@lucide/svelte/icons/columns-2';
	import { Button } from '$lib/components/ui/button';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import { RowsPerPage, TableSearch } from './index';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';

	let {
		table,
		view,
		placeholder = 'Search...',
		fields,
		hideSearch = true,
		embedded = false,
		onviewchange
	}: {
		table: TableHandlerInterface<Row>;
		view?: {
			columns: { index: number; name?: string; isVisible?: boolean; toggle?: () => void }[];
		};
		placeholder?: string;
		fields?: Array<keyof Row | ((row: Row) => unknown)>;
		hideSearch?: boolean;
		embedded?: boolean;
		onviewchange?: () => void;
	} = $props();

	const visibleColumnCount = $derived(view?.columns.filter((column) => column.isVisible).length ?? 0);
</script>

{#if !hideSearch}
	<div class="flex min-w-0 flex-1">
		<TableSearch {table} {placeholder} {fields} class="w-full sm:max-w-xs" />
	</div>
{/if}
<div class="flex shrink-0 flex-nowrap items-center {embedded ? 'gap-0' : 'gap-1.5'}">
	<RowsPerPage {table} {embedded} />
	{#if view}
		<DropdownMenu.Root>
			<DropdownMenu.Trigger>
				{#snippet child({ props })}
					<Button
						{...props}
						variant={embedded ? 'ghost' : 'outline'}
						size="icon-sm"
						class={embedded
							? 'size-7 text-[var(--color-text-tertiary)] hover:text-[var(--color-text-primary)] [&_svg]:size-3.5'
							: ''}
						title="Columns"
						aria-label="Toggle columns"
					>
						<Columns2Icon strokeWidth={1.6} />
					</Button>
				{/snippet}
			</DropdownMenu.Trigger>
			<DropdownMenu.Content class="w-44" align="end" side={embedded ? 'top' : 'bottom'}>
				<DropdownMenu.Group>
					<DropdownMenu.Label>Columns</DropdownMenu.Label>
					<DropdownMenu.Separator />
					{#each view.columns as column (column.index)}
						<DropdownMenu.CheckboxItem
							closeOnSelect={false}
							checked={column.isVisible}
							disabled={visibleColumnCount === 1 && column.isVisible}
							onCheckedChange={() => {
								column.toggle?.();
								onviewchange?.();
							}}
						>
							{column.name}
						</DropdownMenu.CheckboxItem>
					{/each}
				</DropdownMenu.Group>
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	{/if}
</div>
