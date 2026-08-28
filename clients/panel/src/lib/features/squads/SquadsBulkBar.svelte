<script lang="ts">
	import { Button } from '$lib/components/ui/button';

	let {
		count,
		onaddall,
		onremoveall,
		ondelete,
		onclear
	}: {
		count: number;
		onaddall: () => void;
		onremoveall: () => void;
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
		<Button size="sm" variant="outline" onclick={onaddall}>Add all users</Button>
		<Button size="sm" variant="outline" onclick={onremoveall}>Remove all users</Button>
		<Button size="sm" variant="destructive" onclick={ondelete}>Delete</Button>
		<Button size="sm" variant="ghost" onclick={onclear}>Clear</Button>
	</div>
</div>
