<script lang="ts">
	import * as Table from '$lib/components/ui/table';

	let { columns, rows = 8 }: { columns: number; rows?: number } = $props();
</script>

{#each Array.from({ length: rows }, (_, i) => i) as row (row)}
	<Table.Row class="hover:bg-transparent" aria-hidden="true">
		{#each Array.from({ length: columns }, (_, i) => i) as col (col)}
			<Table.Cell class="px-3 py-2.5">
				<div
					class="h-3.5 animate-pulse rounded bg-[var(--color-bg-tertiary)]"
					style="width: {52 + ((row * 3 + col * 7) % 40)}%"
				></div>
			</Table.Cell>
		{/each}
	</Table.Row>
{/each}
