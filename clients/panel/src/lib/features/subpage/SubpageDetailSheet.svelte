<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { pretty } from '$lib/list';

	export type PageConfig = { uuid: string; name: string; config?: unknown };

	let {
		config = $bindable(null),
		pending = false,
		onupdate,
		onclone,
		ondelete
	}: {
		config: PageConfig | null;
		pending?: boolean;
		onupdate: (body: { uuid: string; name: string; config: unknown }) => Promise<void>;
		onclone: () => void;
		ondelete: () => void;
	} = $props();

	const reserved = '00000000-0000-0000-0000-000000000000';

	let name = $state('');
	let draft = $state('{}');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');
	let parseError = $state('');

	$effect(() => {
		const next = config;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		draft = pretty(next.config ?? {});
		parseError = '';
	});

	const open = $derived(config !== null);
	const actions = [
		{ id: 'save', label: 'Save changes' },
		{ id: 'clone', label: 'Clone' }
	];
	const menu = $derived(
		config && config.uuid !== reserved
			? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }]
			: []
	);

	async function runAction() {
		if (!config) return;
		if (actionId === 'clone') {
			onclone();
			return;
		}
		try {
			const parsed = JSON.parse(draft);
			parseError = '';
			await onupdate({ uuid: config.uuid, name: name.trim() || config.name, config: parsed });
		} catch {
			parseError = 'Config must be valid JSON.';
		}
	}
</script>

<DetailSheet
	{open}
	title={config ? `Edit ${config.name}` : 'Edit page config'}
	description="Branding and locales for the hosted subscription page."
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			config = null;
			loadedUuid = null;
		}
	}}
>
	{#if config}
		<div class={chrome.stack}>
			<label class={chrome.field}>
				<span class={chrome.label}>Name</span>
				<Input bind:value={name} maxlength={30} />
			</label>
			<div class={chrome.field}>
				<span class={chrome.label}>Config JSON</span>
				<JsonField bind:value={draft} rows={16} />
				{#if parseError}
					<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
				{/if}
			</div>
		</div>
	{/if}
</DetailSheet>
