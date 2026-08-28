<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	$effect(() => {
		const uuid = page.params.uuid ?? '';
		void goto(uuid ? `/dashboard/subpage?id=${encodeURIComponent(uuid)}` : '/dashboard/subpage', {
			replaceState: true
		});
	});
</script>
