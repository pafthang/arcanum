<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { quickOpen } from '$lib/features/layout/quick-open.svelte';

	onMount(() => {
		quickOpen.open();
		void goto('/dashboard/home', { replaceState: true });
	});
</script>
