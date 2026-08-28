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
		oncreate: (body: { name: string; loginUrl?: string }) => Promise<void>;
	} = $props();

	let name = $state('');
	let loginUrl = $state('');
	let actionId = $state('create');

	async function submit() {
		if (!name.trim()) return;
		await oncreate({ name: name.trim(), loginUrl: loginUrl.trim() || undefined });
		name = '';
		loginUrl = '';
		actionId = 'create';
		open = false;
	}

	const actions = $derived([{ id: 'create', label: 'Create', disabled: pending || !name.trim() }]);
</script>

<AppDialog
	bind:open
	title="Add provider"
	description="Infra providers used for node billing."
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
			<Input bind:value={name} required placeholder="Hetzner" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Login URL</span>
			<Input bind:value={loginUrl} placeholder="https://…" />
		</label>
	</form>
</AppDialog>
