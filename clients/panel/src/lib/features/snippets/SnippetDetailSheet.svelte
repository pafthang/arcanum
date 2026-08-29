<script lang="ts">
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { pretty } from '$lib/list';
	import { errorMessage, parseSnippetArray } from '$lib/json-edit';
	import type { Snippet } from '@arcanum/ts-client';

	let {
		snippet = $bindable(null),
		pending = false,
		onupdate,
		onsync,
		ondelete
	}: {
		snippet: Snippet | null;
		pending?: boolean;
		onupdate: (body: { name: string; snippet: unknown }) => Promise<void>;
		onsync: () => void;
		ondelete: () => void;
	} = $props();

	let draft = $state('[]');
	let loadedName = $state<string | null>(null);
	let actionId = $state('save');
	let parseError = $state('');

	$effect(() => {
		const next = snippet;
		if (!next || loadedName === next.name) return;
		loadedName = next.name;
		actionId = 'save';
		draft = pretty(next.snippet);
		parseError = '';
	});

	const open = $derived(snippet !== null);
	const actions = [
		{ id: 'save', label: 'Save changes' },
		{ id: 'sync', label: 'Sync' }
	];
	const menu = $derived(snippet ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!snippet) return;
		if (actionId === 'sync') {
			onsync();
			return;
		}
		const parsed = parseSnippetArray(draft);
		if (!parsed.ok) {
			parseError = parsed.error;
			return;
		}
		parseError = '';
		try {
			await onupdate({ name: snippet.name, snippet: parsed.value });
		} catch (err) {
			parseError = errorMessage(err, 'Save failed');
		}
	}
</script>

<DetailSheet
	{open}
	title={snippet ? `Edit ${snippet.name}` : 'Edit snippet'}
	description="Must be a non-empty JSON array of objects. Empty arrays or objects are rejected."
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			snippet = null;
			loadedName = null;
		}
	}}
>
	{#if snippet}
		<div class={chrome.field}>
			<span class={chrome.label}>JSON</span>
			<JsonField bind:value={draft} rows={16} />
			{#if parseError}
				<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
			{/if}
		</div>
	{/if}
</DetailSheet>
