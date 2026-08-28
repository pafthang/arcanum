<script lang="ts">
	import { Checkbox } from '$lib/components/ui/checkbox';

	type Inbound = { uuid: string; tag: string; type?: string; port?: number | null };

	let {
		inbounds,
		selected = $bindable([]),
		empty = 'No inbounds yet. Create a config profile inbound first.'
	}: {
		inbounds: Inbound[];
		selected: string[];
		empty?: string;
	} = $props();

	function toggle(uuid: string, checked: boolean) {
		if (checked) {
			if (!selected.includes(uuid)) selected = [...selected, uuid];
		} else {
			selected = selected.filter((id) => id !== uuid);
		}
	}
</script>

{#if inbounds.length === 0}
	<p class="text-xs text-[var(--color-text-tertiary)]">{empty}</p>
{:else}
	<div class="max-h-52 space-y-0.5 overflow-y-auto rounded-lg border border-[var(--app-border)] p-1.5">
		{#each inbounds as inbound (inbound.uuid)}
			<label class="flex cursor-pointer items-center gap-2.5 rounded-md px-2 py-1.5 text-sm hover:bg-[var(--color-bg-hover)]/50">
				<Checkbox
					checked={selected.includes(inbound.uuid)}
					onCheckedChange={(value) => toggle(inbound.uuid, Boolean(value))}
				/>
				<span class="min-w-0 truncate font-medium">{inbound.tag}</span>
				<span class="ml-auto shrink-0 text-[11px] text-[var(--color-text-tertiary)]">
					{inbound.type ?? 'inbound'}{inbound.port != null ? ` :${inbound.port}` : ''}
				</span>
			</label>
		{/each}
	</div>
{/if}
