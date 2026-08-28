<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import * as Select from '$lib/components/ui/select';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';
	import { cn } from '$lib/utils.js';

	type Props = {
		table: TableHandlerInterface<Row>;
		options?: number[];
		class?: string;
		embedded?: boolean;
	};

	let { table, options = [10, 20, 50, 100], class: className = '', embedded = false }: Props = $props();

	function handleValueChange(value: string) {
		const rowsPerPage = Number(value);
		if (Number.isNaN(rowsPerPage) || rowsPerPage === table.rowsPerPage) return;
		const handler = table as TableHandlerInterface<Row> & { setRowsPerPage?: (value: number) => void };
		if (handler.setRowsPerPage) handler.setRowsPerPage(rowsPerPage);
		else handler.rowsPerPage = rowsPerPage;
		table.setPage(1);
	}
</script>

<aside class={cn('flex items-center', !embedded && 'h-7 text-[11px] text-[var(--color-text-tertiary)]', className)}>
	<Select.Root type="single" value={String(table.rowsPerPage)} onValueChange={handleValueChange}>
		<Select.Trigger
			size="sm"
			class={embedded
				? 'h-7 gap-1 border-0 bg-transparent px-2 text-[11px] shadow-none hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)] dark:bg-transparent dark:hover:bg-[var(--color-bg-hover)] [&_svg]:size-3 [&_svg]:opacity-50'
				: 'h-7 min-w-14 px-2 text-xs'}
			title="Rows per page"
			aria-label="Rows per page"
		>
			{table.rowsPerPage}
			{#if embedded}
				<span class="text-[var(--color-text-tertiary)]">rows</span>
			{/if}
		</Select.Trigger>
		<Select.Content class="p-0.5" align="end" side="top">
			{#each options as option (option)}
				<Select.Item value={String(option)} label={String(option)}>{option}</Select.Item>
			{/each}
		</Select.Content>
	</Select.Root>
</aside>
