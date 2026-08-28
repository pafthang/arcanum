<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { DEFAULT_SNIPPET, errorMessage, parseSnippetArray } from '$lib/json-edit';

	let {
		open = $bindable(false),
		pending = false,
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		oncreate: (body: { name: string; snippet: unknown }) => Promise<void>;
	} = $props();

	let name = $state('');
	let draft = $state(DEFAULT_SNIPPET);
	let actionId = $state('create');
	let parseError = $state('');

	function reset() {
		name = '';
		draft = DEFAULT_SNIPPET;
		actionId = 'create';
		parseError = '';
	}

	async function submit() {
		if (!name.trim()) return;
		const parsed = parseSnippetArray(draft);
		if (!parsed.ok) {
			parseError = parsed.error;
			return;
		}
		parseError = '';
		try {
			await oncreate({ name: name.trim(), snippet: parsed.value });
			reset();
			open = false;
		} catch (err) {
			parseError = errorMessage(err, 'Create failed');
		}
	}

	const actions = $derived([{ id: 'create', label: 'Create', disabled: pending || !name.trim() }]);
</script>

<AppDialog
	bind:open
	title="Create snippet"
	description="Named JSON fragment reused by config profiles. Must be a non-empty array of objects."
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
			<Input bind:value={name} required placeholder="reality-shortids" />
		</label>
		<div class={chrome.field}>
			<span class={chrome.label}>JSON array</span>
			<JsonField bind:value={draft} rows={10} />
			<p class={chrome.hint}>Use a JSON array of objects. Empty [] or empty objects are rejected.</p>
			{#if parseError}
				<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
			{/if}
		</div>
	</form>
</AppDialog>
