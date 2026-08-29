<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import type { ConfigProfile } from '@arcanum/ts-client';

	const selectClass =
		'h-8 w-full rounded-md border border-[var(--app-border)] bg-[var(--color-bg)] px-2 text-sm text-[var(--color-text-primary)]';

	let {
		open = $bindable(false),
		pending = false,
		profiles = [],
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		profiles?: ConfigProfile[];
		oncreate: (body: {
			name: string;
			address: string;
			port?: number;
			configProfile: { activeConfigProfileUuid: string; activeInbounds: string[] };
		}) => Promise<void>;
	} = $props();

	let name = $state('');
	let address = $state('');
	let port = $state('2222');
	let profileUuid = $state('');
	let inboundUuid = $state('');
	let actionId = $state('create');

	const selectedProfile = $derived(profiles.find((profile) => profile.uuid === profileUuid));

	$effect(() => {
		if (!open) return;
		if (!profileUuid && profiles[0]) profileUuid = profiles[0].uuid;
	});

	$effect(() => {
		const first = selectedProfile?.inbounds?.[0]?.uuid ?? '';
		if (first && !selectedProfile?.inbounds?.some((inbound) => inbound.uuid === inboundUuid)) {
			inboundUuid = first;
		}
	});

	function reset() {
		name = '';
		address = '';
		port = '2222';
		actionId = 'create';
	}

	async function submit() {
		if (!name.trim() || !address.trim() || !profileUuid) return;
		await oncreate({
			name: name.trim(),
			address: address.trim(),
			port: Number(port) || undefined,
			configProfile: {
				activeConfigProfileUuid: profileUuid,
				activeInbounds: inboundUuid ? [inboundUuid] : []
			}
		});
		reset();
		open = false;
	}

	const actions = $derived([
		{
			id: 'create',
			label: 'Create',
			disabled: pending || !name.trim() || !address.trim() || !profileUuid
		}
	]);
</script>

<AppDialog
	bind:open
	title="Create node"
	description="A config profile with inbounds is required to start."
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
			<Input bind:value={name} required placeholder="eu-1" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Address</span>
			<Input bind:value={address} required placeholder="node.example.com" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Port</span>
			<Input bind:value={port} inputmode="numeric" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Config profile</span>
			<select class={selectClass} bind:value={profileUuid}>
				{#each profiles as profile (profile.uuid)}
					<option value={profile.uuid}>{profile.name}</option>
				{/each}
			</select>
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Inbound</span>
			<select class={selectClass} bind:value={inboundUuid}>
				{#each selectedProfile?.inbounds ?? [] as inbound (inbound.uuid)}
					<option value={inbound.uuid}>{inbound.tag} ({inbound.type})</option>
				{/each}
			</select>
		</label>
	</form>
</AppDialog>
