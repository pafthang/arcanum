<script lang="ts" generics="Row extends import('@vincjo/datatables/server').Row">
	import SearchIcon from '@lucide/svelte/icons/search';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import XIcon from '@lucide/svelte/icons/x';
	import * as InputGroup from '$lib/components/ui/input-group';
	import { cn } from '$lib/utils.js';
	import type { TableHandlerInterface } from '@vincjo/datatables/server';

	type Props = {
		table: TableHandlerInterface<Row>;
		placeholder?: string;
		class?: string;
		id?: string;
		fields?: Array<keyof Row | ((row: Row) => unknown)>;
	};
	let { table, placeholder = 'Search...', class: className, id = 'list-search', fields }: Props = $props();

	// svelte-ignore state_referenced_locally
	const search = (
		table as TableHandlerInterface<Row> & {
			createSearch: (scope?: Array<keyof Row | ((row: Row) => unknown)>) => ReturnType<
				TableHandlerInterface<Row>['createSearch']
			>;
		}
	).createSearch(fields);
	const loading = $derived('isLoading' in table && Boolean((table as { isLoading?: boolean }).isLoading));

	let timer: ReturnType<typeof setTimeout> | undefined;

	function apply() {
		search.set();
	}

	function oninput() {
		clearTimeout(timer);
		timer = setTimeout(apply, 160);
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			clearTimeout(timer);
			apply();
		}
	}

	function clear() {
		clearTimeout(timer);
		search.value = '';
		search.set();
	}
</script>

<InputGroup.Root class={cn('h-8 max-w-xs', className)}>
	<InputGroup.Addon>
		{#if loading && search.value.length > 0}
			<Loader2Icon class="size-3.5 animate-spin" />
		{:else}
			<SearchIcon class="size-3.5" />
		{/if}
	</InputGroup.Addon>
	<InputGroup.Input {id} type="search" {placeholder} bind:value={search.value} {oninput} {onkeydown} />
	{#if search.value}
		<InputGroup.Addon align="inline-end">
			<button type="button" class="rounded p-0.5 hover:text-[var(--color-text-primary)]" aria-label="Clear search" onclick={clear}>
				<XIcon class="size-3.5" />
			</button>
		</InputGroup.Addon>
	{/if}
</InputGroup.Root>
