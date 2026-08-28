<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import BootScreen from '$lib/components/auth/BootScreen.svelte';
	import { session } from '$lib/rw';

	onMount(() => {
		session.init();
		void goto(session.isAuthenticated ? '/dashboard/home' : '/auth/login', { replaceState: true });
	});
</script>

<BootScreen />
