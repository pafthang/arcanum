<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	$effect(() => {
		const id = page.params.id ?? '';
		void goto(id ? `/dashboard/management/users?id=${encodeURIComponent(id)}` : '/dashboard/management/users', {
			replaceState: true
		});
	});
</script>
