<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';

	let { value }: { value: string | boolean } = $props();

	const label = $derived(typeof value === 'boolean' ? (value ? 'yes' : 'no') : value);
	const variant = $derived.by(() => {
		const v = String(label).toUpperCase();
		if (['ACTIVE', 'CONNECTED', 'YES', 'TRUE', 'ONLINE'].includes(v)) return 'secondary' as const;
		if (['DISABLED', 'EXPIRED', 'NO', 'FALSE', 'OFFLINE'].includes(v)) return 'destructive' as const;
		if (['LIMITED', 'CONNECTING'].includes(v)) return 'outline' as const;
		return 'outline' as const;
	});
</script>

<Badge {variant} class="pointer-events-none">{label}</Badge>
