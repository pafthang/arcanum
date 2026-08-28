<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';

	$effect(() => {
		const type = page.params.type ?? 'XRAY_JSON';
		const uuid = page.params.uuid ?? '';
		void goto(
			uuid
				? `/dashboard/templates/${type}?id=${encodeURIComponent(uuid)}`
				: `/dashboard/templates/${type}`,
			{ replaceState: true }
		);
	});
</script>
