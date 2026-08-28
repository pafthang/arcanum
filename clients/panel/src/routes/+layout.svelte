<script lang="ts">
	import '../app.css';
	import { Toaster } from 'svelte-sonner';
	import { onMount } from 'svelte';
	import { session } from '$lib/rw';
	import { preferencesState } from '$lib/features/preferences/preferences.state.svelte';

	let { children } = $props();

	onMount(() => {
		session.init();
		preferencesState.init();
	});
</script>

<svelte:head>
	<title>OptimaWave</title>
</svelte:head>

<Toaster position="bottom-right" theme={preferencesState.resolvedMode} />
{@render children()}
