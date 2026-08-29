<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import type { ConfigInbound } from '@arcanum/ts-client';

	const selectClass =
		'h-8 w-full rounded-md border border-[var(--app-border)] bg-[var(--color-bg)] px-2 text-sm text-[var(--color-text-primary)]';

	let {
		open = $bindable(false),
		pending = false,
		inbounds = [],
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		inbounds?: ConfigInbound[];
		oncreate: (body: {
			remark: string;
			address: string;
			port: number;
			inbound: { configProfileUuid: string | null; configProfileInboundUuid: string | null };
		}) => Promise<void>;
	} = $props();

	let remark = $state('');
	let address = $state('');
	let port = $state('443');
	let inboundUuid = $state('');
	let actionId = $state('create');

	function reset() {
		remark = '';
		address = '';
		port = '443';
		actionId = 'create';
	}

	async function submit() {
		if (!remark.trim() || !address.trim()) return;
		const inbound = inbounds.find((item) => item.uuid === inboundUuid);
		await oncreate({
			remark: remark.trim(),
			address: address.trim(),
			port: Number(port) || 443,
			inbound: {
				configProfileUuid: inbound?.profileUuid ?? null,
				configProfileInboundUuid: inboundUuid || null
			}
		});
		reset();
		open = false;
	}

	const actions = $derived([
		{ id: 'create', label: 'Create', disabled: pending || !remark.trim() || !address.trim() }
	]);
</script>

<AppDialog
	bind:open
	title="Create host"
	description="Hosts appear in subscriptions for the selected inbound."
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
			<span class={chrome.label}>Remark</span>
			<Input bind:value={remark} required placeholder="EU Reality" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Address</span>
			<Input bind:value={address} required placeholder="host.example.com" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Port</span>
			<Input bind:value={port} inputmode="numeric" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Inbound</span>
			<select class={selectClass} bind:value={inboundUuid}>
				<option value="">None</option>
				{#each inbounds as inbound (inbound.uuid)}
					<option value={inbound.uuid}>{inbound.tag} · {inbound.type}</option>
				{/each}
			</select>
		</label>
	</form>
</AppDialog>
