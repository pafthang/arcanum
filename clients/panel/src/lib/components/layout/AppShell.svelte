<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import AppSidebar from './Sidebar.svelte';
	import AppHeader from './AppHeader.svelte';
	import StatusBar from './StatusBar.svelte';
	import { contentWidth } from '$lib/features/layout/content-width.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<Sidebar.Provider class="h-dvh min-h-0 overflow-hidden bg-[var(--color-bg)]">
	<AppSidebar />
	<Sidebar.Inset class="min-h-0 overflow-hidden bg-[var(--color-bg)]">
		<AppHeader />
		<div class="min-h-0 flex-1 overflow-y-auto">
			<div
				class="mx-auto w-full space-y-8 px-4 py-6 sm:px-6 {contentWidth.value === 'wide'
					? 'max-w-none'
					: 'max-w-[1440px]'}"
			>
				{@render children()}
			</div>
		</div>
		<StatusBar />
	</Sidebar.Inset>
</Sidebar.Provider>
