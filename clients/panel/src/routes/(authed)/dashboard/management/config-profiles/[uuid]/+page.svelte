<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	$effect(() => {
		const uuid = page.params.uuid ?? '';
		void goto(
			uuid
				? `/dashboard/management/config-profiles?id=${encodeURIComponent(uuid)}`
				: '/dashboard/management/config-profiles',
			{ replaceState: true }
		);
	});
</script>
