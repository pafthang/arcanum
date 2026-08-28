<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import StatusBadge from '$lib/components/remnawave/StatusBadge.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import type { ConfigInbound, Host } from '@optimawave/ts-back';

	const selectClass =
		'h-8 w-full rounded-md border border-[var(--app-border)] bg-[var(--color-bg)] px-2 text-sm text-[var(--color-text-primary)]';

	let {
		host = $bindable(null),
		pending = false,
		inbounds = [],
		onupdate,
		onenable,
		ondisable,
		ondelete
	}: {
		host: Host | null;
		pending?: boolean;
		inbounds?: ConfigInbound[];
		onupdate: (body: {
			uuid: string;
			remark: string;
			address: string;
			port: number;
			inbound: { configProfileUuid: string | null; configProfileInboundUuid: string | null };
		}) => Promise<void>;
		onenable: () => void;
		ondisable: () => void;
		ondelete: () => void;
	} = $props();

	let remark = $state('');
	let address = $state('');
	let port = $state('');
	let inboundUuid = $state('');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');

	$effect(() => {
		const next = host;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		remark = next.remark;
		address = next.address;
		port = String(next.port ?? 443);
		inboundUuid = next.inbound?.configProfileInboundUuid ?? '';
	});

	const open = $derived(host !== null);
	const actions = $derived(
		host
			? [
					{ id: 'save', label: 'Save changes' },
					host.isDisabled ? { id: 'enable', label: 'Enable' } : { id: 'disable', label: 'Disable' }
				]
			: [{ id: 'save', label: 'Save changes' }]
	);
	const menu = $derived(host ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!host) return;
		if (actionId === 'enable') {
			onenable();
			return;
		}
		if (actionId === 'disable') {
			ondisable();
			return;
		}
		const inbound = inbounds.find((item) => item.uuid === inboundUuid);
		await onupdate({
			uuid: host.uuid,
			remark: remark.trim() || host.remark,
			address: address.trim() || host.address,
			port: Number(port) || host.port,
			inbound: {
				configProfileUuid: inbound?.profileUuid ?? host.inbound?.configProfileUuid ?? null,
				configProfileInboundUuid: inboundUuid || null
			}
		});
	}
</script>

<DetailSheet
	{open}
	title={host ? `Edit ${host.remark}` : 'Edit host'}
	description={host ? `${host.address}:${host.port}` : ''}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			host = null;
			loadedUuid = null;
		}
	}}
>
	{#if host}
		<div class={chrome.stack}>
			<StatusBadge value={host.isDisabled ? 'DISABLED' : 'ACTIVE'} />
			<label class={chrome.field}>
				<span class={chrome.label}>Remark</span>
				<Input bind:value={remark} />
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
				<span class={chrome.label}>Inbound</span>
				<select class={selectClass} bind:value={inboundUuid}>
					<option value="">None</option>
					{#each inbounds as inbound (inbound.uuid)}
						<option value={inbound.uuid}>{inbound.tag} · {inbound.type}</option>
					{/each}
				</select>
			</label>
		</div>
	{/if}
</DetailSheet>
