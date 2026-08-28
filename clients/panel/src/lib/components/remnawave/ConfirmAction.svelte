<script lang="ts">
	import AppDialog from './AppDialog.svelte';

	let {
		open = $bindable(false),
		title,
		description,
		confirmLabel = 'Confirm',
		variant = 'destructive',
		pending = false,
		onconfirm
	}: {
		open: boolean;
		title: string;
		description: string;
		confirmLabel?: string;
		variant?: 'destructive' | 'default';
		pending?: boolean;
		onconfirm: () => void | Promise<void>;
	} = $props();

	let busy = $state(false);
	let actionId = $state('confirm');

	const actions = $derived([
		{
			id: 'confirm',
			label: confirmLabel,
			variant: variant === 'destructive' ? ('destructive' as const) : ('default' as const)
		}
	]);

	async function confirm() {
		if (busy || pending) return;
		busy = true;
		try {
			await onconfirm();
			open = false;
		} finally {
			busy = false;
		}
	}
</script>

<AppDialog
	bind:open
	{title}
	{description}
	{actions}
	bind:actionId
	pending={busy || pending}
	onrun={confirm}
/>
