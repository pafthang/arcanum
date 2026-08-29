<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { pretty } from '$lib/list';
	import type { NodePlugin } from '@arcanum/ts-client';

	let {
		plugin = $bindable(null),
		pending = false,
		onupdate,
		onclone,
		ondelete
	}: {
		plugin: NodePlugin | null;
		pending?: boolean;
		onupdate: (body: { uuid: string; name: string; pluginConfig: unknown }) => Promise<void>;
		onclone: () => void;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let draft = $state('{}');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');
	let parseError = $state('');

	$effect(() => {
		const next = plugin;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		draft = pretty(next.pluginConfig);
		parseError = '';
	});

	const open = $derived(plugin !== null);
	const actions = [
		{ id: 'save', label: 'Save changes' },
		{ id: 'clone', label: 'Clone' }
	];
	const menu = $derived(plugin ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!plugin) return;
		if (actionId === 'clone') {
			onclone();
			return;
		}
		try {
			const pluginConfig = JSON.parse(draft);
			parseError = '';
			await onupdate({ uuid: plugin.uuid, name: name.trim() || plugin.name, pluginConfig });
		} catch {
			parseError = 'Config must be valid JSON.';
		}
	}
</script>

<DetailSheet
	{open}
	title={plugin ? `Edit ${plugin.name}` : 'Edit plugin'}
	description="Per-node plugin config. Sync applies it to connected nodes."
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			plugin = null;
			loadedUuid = null;
		}
	}}
>
	{#if plugin}
		<div class={chrome.stack}>
			<label class={chrome.field}>
				<span class={chrome.label}>Name</span>
				<Input bind:value={name} maxlength={30} />
			</label>
			<div class={chrome.field}>
				<span class={chrome.label}>Plugin config</span>
				<JsonField bind:value={draft} rows={16} />
				{#if parseError}
					<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
				{/if}
			</div>
		</div>
	{/if}
</DetailSheet>
