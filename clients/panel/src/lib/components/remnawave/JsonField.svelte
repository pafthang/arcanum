<script lang="ts">
	import { Textarea } from '$lib/components/ui/textarea';

	let { value = $bindable(''), rows = 14, class: className = '' }: { value?: string; rows?: number; class?: string } =
		$props();
</script>

<Textarea bind:value {rows} class="font-mono text-xs {className}" />
