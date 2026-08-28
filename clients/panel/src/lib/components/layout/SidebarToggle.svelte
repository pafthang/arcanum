<script lang="ts">
	import { PanelLeft } from 'lucide-svelte';
	import { useSidebar } from '$lib/components/ui/sidebar';

	const sidebar = useSidebar();
	const visible = $derived(sidebar.isMobile || sidebar.state === 'collapsed');
</script>

{#if visible}
	<button
		type="button"
		onclick={() => (sidebar.isMobile ? sidebar.setOpenMobile(true) : sidebar.setOpen(true))}
		class="shrink-0 rounded-md p-1 text-[var(--color-text-tertiary)] hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]"
		title="Expand sidebar"
	>
		<PanelLeft size={16} />
	</button>
{/if}
