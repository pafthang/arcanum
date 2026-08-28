<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import InboundPicker from './InboundPicker.svelte';

	type Inbound = { uuid: string; tag: string; type?: string; port?: number | null };
	type Squad = {
		uuid: string;
		name: string;
		info?: { membersCount?: number; inboundsCount?: number };
		inbounds?: { uuid: string }[];
	};

	let {
		squad = $bindable(null),
		kind,
		pending = false,
		inbounds = [],
		onsave,
		onaddall,
		onremoveall,
		ondelete
	}: {
		squad: Squad | null;
		kind: 'internal' | 'external';
		pending?: boolean;
		inbounds?: Inbound[];
		onsave: (body: { uuid: string; name: string; inbounds?: string[] }) => Promise<void>;
		onaddall: () => void;
		onremoveall: () => void;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let selected = $state<string[]>([]);
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');

	$effect(() => {
		const next = squad;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		selected = (next.inbounds ?? []).map((item) => item.uuid);
	});

	const open = $derived(squad !== null);
	const description =
		kind === 'internal'
			? 'Users in this squad share the selected inbounds.'
			: 'Users in this squad can override hosts.';

	const actions = [
		{ id: 'save', label: 'Save changes' },
		{ id: 'add', label: 'Add all users' },
		{ id: 'remove', label: 'Remove all users' }
	];

	const menu = [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }];

	async function runAction() {
		if (!squad) return;
		if (actionId === 'add') {
			onaddall();
			return;
		}
		if (actionId === 'remove') {
			onremoveall();
			return;
		}
		await onsave({
			uuid: squad.uuid,
			name: name.trim(),
			inbounds: kind === 'internal' ? selected : undefined
		});
	}
</script>

<DetailSheet
	{open}
	title={squad ? `Edit ${squad.name}` : 'Edit squad'}
	{description}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			squad = null;
			loadedUuid = null;
		}
	}}
>
	{#if squad}
		<div class="space-y-6">
			<section class="grid grid-cols-2 gap-3">
				<div class={chrome.tile}>
					<p class={chrome.hint}>Members</p>
					<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{squad.info?.membersCount ?? 0}</p>
				</div>
				{#if kind === 'internal'}
					<div class={chrome.tile}>
						<p class={chrome.hint}>Inbounds</p>
						<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">
							{squad.info?.inboundsCount ?? selected.length}
						</p>
					</div>
				{:else}
					<div class={chrome.tile}>
						<p class={chrome.hint}>Type</p>
						<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">Host overrides</p>
					</div>
				{/if}
			</section>

			<section class={chrome.stack}>
				<p class={chrome.section}>Record</p>
				<label class={chrome.field}>
					<span class={chrome.label}>Name</span>
					<Input bind:value={name} />
				</label>
				{#if kind === 'internal'}
					<div class={chrome.field}>
						<span class={chrome.label}>Inbounds</span>
						<InboundPicker {inbounds} bind:selected />
					</div>
				{/if}
			</section>
		</div>
	{/if}
</DetailSheet>
