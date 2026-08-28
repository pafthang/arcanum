<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';

	export type Provider = {
		uuid: string;
		name: string;
		loginUrl?: string | null;
		billingHistory?: { totalAmount?: number; totalBills?: number };
		billingNodes?: unknown[];
	};

	let {
		provider = $bindable(null),
		pending = false,
		onupdate,
		ondelete
	}: {
		provider: Provider | null;
		pending?: boolean;
		onupdate: (body: { uuid: string; name: string; loginUrl?: string }) => Promise<void>;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let loginUrl = $state('');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');

	$effect(() => {
		const next = provider;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		loginUrl = next.loginUrl ?? '';
	});

	const open = $derived(provider !== null);
	const actions = [{ id: 'save', label: 'Save changes' }];
	const menu = $derived(provider ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!provider) return;
		await onupdate({
			uuid: provider.uuid,
			name: name.trim() || provider.name,
			loginUrl: loginUrl.trim() || undefined
		});
	}
</script>

<DetailSheet
	{open}
	title={provider ? `Edit ${provider.name}` : 'Edit provider'}
	description="Billing provider used by infra nodes."
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			provider = null;
			loadedUuid = null;
		}
	}}
>
	{#if provider}
		<div class={chrome.stack}>
			<section class="grid grid-cols-2 gap-3">
				<div class={chrome.tile}>
					<p class={chrome.hint}>Bills</p>
					<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">
						{provider.billingHistory?.totalBills ?? 0}
					</p>
				</div>
				<div class={chrome.tile}>
					<p class={chrome.hint}>Spent</p>
					<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">
						{provider.billingHistory?.totalAmount ?? 0}
					</p>
				</div>
			</section>
			<label class={chrome.field}>
				<span class={chrome.label}>Name</span>
				<Input bind:value={name} />
			</label>
			<label class={chrome.field}>
				<span class={chrome.label}>Login URL</span>
				<Input bind:value={loginUrl} />
			</label>
		</div>
	{/if}
</DetailSheet>
