<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	let {
		count,
		onenable,
		ondisable,
		ondelete,
		onclear
	}: {
		count: number;
		onenable: () => void;
		ondisable: () => void;
		ondelete: () => void;
		onclear: () => void;
	} = $props();
</script>

<div
	class="flex flex-wrap items-center gap-2 rounded-md border border-[var(--app-border)] bg-[var(--color-bg-secondary)]/60 px-3 py-2"
>
	<p class="text-xs text-[var(--color-text-secondary)]">
		<span class="font-medium text-[var(--color-text-primary)]">{count}</span> selected
	</p>
	<div class="ml-auto flex flex-wrap gap-1.5">
		<Button size="sm" variant="outline" onclick={onenable}>Enable</Button>
		<Button size="sm" variant="outline" onclick={ondisable}>Disable</Button>
		<Button size="sm" variant="destructive" onclick={ondelete}>Delete</Button>
		<Button size="sm" variant="ghost" onclick={onclear}>Clear</Button>
	</div>
</div>
