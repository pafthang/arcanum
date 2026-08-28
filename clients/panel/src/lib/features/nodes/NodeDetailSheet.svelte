<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import StatusBadge from '$lib/components/remnawave/StatusBadge.svelte';
	import TrafficBar from '$lib/components/remnawave/TrafficBar.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import type { ConfigProfile, Node } from '@optimawave/ts-back';
	import { nodeEndpoint, nodeProfile, nodeStatus } from './status';

	const selectClass =
		'h-8 w-full rounded-md border border-[var(--app-border)] bg-[var(--color-bg)] px-2 text-sm text-[var(--color-text-primary)]';

	let {
		node = $bindable(null),
		pending = false,
		profiles = [],
		onupdate,
		onenable,
		ondisable,
		onrestart,
		onreset,
		ondelete
	}: {
		node: Node | null;
		pending?: boolean;
		profiles?: ConfigProfile[];
		onupdate: (body: {
			uuid: string;
			name: string;
			address: string;
			port?: number;
			configProfile: { activeConfigProfileUuid: string; activeInbounds: string[] };
		}) => Promise<void>;
		onenable: () => void;
		ondisable: () => void;
		onrestart: () => void;
		onreset: () => void;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let address = $state('');
	let port = $state('');
	let profileUuid = $state('');
	let inboundUuid = $state('');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');

	const selectedProfile = $derived(profiles.find((profile) => profile.uuid === profileUuid));
	const status = $derived(node ? nodeStatus(node) : '');

	$effect(() => {
		const next = node;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		address = next.address;
		port = next.port != null ? String(next.port) : '';
		const profile = nodeProfile(next);
		profileUuid = profile.uuid ?? profiles[0]?.uuid ?? '';
		inboundUuid = profile.inbounds[0]?.uuid ?? selectedProfile?.inbounds?.[0]?.uuid ?? '';
	});

	const open = $derived(node !== null);

	const actions = $derived(
		node
			? [
					{ id: 'save', label: 'Save changes' },
					node.isDisabled ? { id: 'enable', label: 'Enable' } : { id: 'disable', label: 'Disable' },
					{ id: 'restart', label: 'Restart' },
					{ id: 'reset', label: 'Reset traffic' }
				]
			: [{ id: 'save', label: 'Save changes' }]
	);

	const menu = $derived(node ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!node) return;
		if (actionId === 'enable') {
			onenable();
			return;
		}
		if (actionId === 'disable') {
			ondisable();
			return;
		}
		if (actionId === 'restart') {
			onrestart();
			return;
		}
		if (actionId === 'reset') {
			onreset();
			return;
		}
		if (!name.trim() || !address.trim() || !profileUuid) return;
		await onupdate({
			uuid: node.uuid,
			name: name.trim(),
			address: address.trim(),
			port: Number(port) || undefined,
			configProfile: {
				activeConfigProfileUuid: profileUuid,
				activeInbounds: inboundUuid ? [inboundUuid] : []
			}
		});
	}
</script>

<DetailSheet
	{open}
	title={node ? `Edit ${node.name}` : 'Edit node'}
	description={node ? nodeEndpoint(node) : ''}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			node = null;
			loadedUuid = null;
		}
	}}
>
	{#if node}
		<div class="space-y-6">
			<section class="space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<StatusBadge value={status} />
					<span class="text-[12px] text-[var(--color-text-tertiary)]">{node.usersOnline ?? 0} online</span>
				</div>
				<div class={chrome.tile}>
					<TrafficBar used={node.trafficUsedBytes} limit={node.trafficLimitBytes} />
				</div>
			</section>
			<section class={chrome.stack}>
				<p class={chrome.section}>Record</p>
				<label class={chrome.field}>
					<span class={chrome.label}>Name</span>
					<Input bind:value={name} />
				</label>
				<label class={chrome.field}>
					<span class={chrome.label}>Address</span>
					<Input bind:value={address} />
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
			</section>
		</div>
	{/if}
</DetailSheet>
