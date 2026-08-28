<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import InboundPicker from './InboundPicker.svelte';

	type Inbound = { uuid: string; tag: string; type?: string; port?: number | null };

	let {
		open = $bindable(false),
		kind,
		pending = false,
		inbounds = [],
		oncreate
	}: {
		open: boolean;
		kind: 'internal' | 'external';
		pending?: boolean;
		inbounds?: Inbound[];
		oncreate: (body: { name: string; inbounds?: string[] }) => Promise<void>;
	} = $props();

	let name = $state('');
	let selected = $state<string[]>([]);
	let actionId = $state('create');

	async function submit() {
		if (!name.trim()) return;
		await oncreate({
			name: name.trim(),
			inbounds: kind === 'internal' ? selected : undefined
		});
		name = '';
		selected = [];
		actionId = 'create';
		open = false;
	}

	const actions = $derived([{ id: 'create', label: 'Create', disabled: pending || !name.trim() }]);
</script>

<AppDialog
	bind:open
	title="Create {kind} squad"
	description={kind === 'internal'
		? 'Group users by inbounds. Pick the inbounds this squad should include.'
		: 'Group users to override hosts on the subscription.'}
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
			<Input bind:value={name} required placeholder="Premium" />
		</label>
		{#if kind === 'internal'}
			<div class={chrome.field}>
				<span class={chrome.label}>Inbounds</span>
				<InboundPicker {inbounds} bind:selected />
			</div>
		{/if}
	</form>
</AppDialog>
