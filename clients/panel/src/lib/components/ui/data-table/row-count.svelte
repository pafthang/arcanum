<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import type { TableHandlerInterface } from '@vincjo/datatables/server';

	let {
		table,
		selection = false,
		loading = false
	}: { table: TableHandlerInterface<Row>; selection?: boolean; loading?: boolean } = $props();
	const { start, end, total, selected } = $derived(table.rowCount);
</script>

<p class="text-[11px] leading-8 text-[var(--color-text-tertiary)]">
	{#if loading && total === 0}
		Loading…
	{:else if selection && selected > 0}
		<span class="text-[var(--color-text-secondary)]">{selected}</span>
		of {total} selected
		<button type="button" class="ml-1 hover:text-[var(--color-text-primary)] hover:underline" onclick={() => table.clearSelection()}>
			Clear
		</button>
	{:else if total > 0}
		{start}–{end} of {total}
	{:else}
		{table.i18n.noRows ?? 'No entries'}
	{/if}
</p>
