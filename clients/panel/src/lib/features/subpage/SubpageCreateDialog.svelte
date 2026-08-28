<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';

	let {
		open = $bindable(false),
		pending = false,
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		oncreate: (name: string) => Promise<void>;
	} = $props();

	let name = $state('');
	let actionId = $state('create');

	async function submit() {
		if (!name.trim()) return;
		await oncreate(name.trim());
		name = '';
		actionId = 'create';
		open = false;
	}

	const actions = $derived([{ id: 'create', label: 'Create', disabled: pending || !name.trim() }]);
</script>

<AppDialog
	bind:open
	title="Create page config"
	description="Name must be 2–30 characters: letters, numbers, spaces, _ or -."
	{actions}
	bind:actionId
	{pending}
	onrun={submit}
>
	<form
		class={chrome.stack}
		onsubmit={(event) => {
			event.preventDefault();
			void submit();
		}}
	>
		<label class={chrome.field}>
			<span class={chrome.label}>Name</span>
			<Input bind:value={name} required placeholder="Default" maxlength={30} />
		</label>
	</form>
</AppDialog>
