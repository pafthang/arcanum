<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { DEFAULT_PROFILE_CONFIG, errorMessage, parseJsonObject, profileHasInbound } from '$lib/json-edit';

	let {
		open = $bindable(false),
		pending = false,
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		oncreate: (body: { name: string; config: unknown }) => Promise<void>;
	} = $props();

	let name = $state('');
	let draft = $state(DEFAULT_PROFILE_CONFIG);
	let actionId = $state('create');
	let parseError = $state('');

	function reset() {
		name = '';
		draft = DEFAULT_PROFILE_CONFIG;
		actionId = 'create';
		parseError = '';
	}

	async function submit() {
		if (!name.trim()) return;
		const parsed = parseJsonObject(draft);
		if (!parsed.ok) {
			parseError = parsed.error;
			return;
		}
		if (!profileHasInbound(parsed.value)) {
			parseError = 'Config must include at least one inbound (inbounds cannot be empty).';
			return;
		}
		parseError = '';
		try {
			await oncreate({ name: name.trim(), config: parsed.value });
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
	title="Create config profile"
	description="Xray JSON is split into inbounds after save. At least one inbound is required."
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
			<Input bind:value={name} required placeholder="Main" />
		</label>
		<div class={chrome.field}>
			<span class={chrome.label}>Config JSON</span>
			<JsonField bind:value={draft} rows={10} />
			<p class={chrome.hint}>inbounds must contain at least one tagged inbound (vless, shadowsocks, …).</p>
			{#if parseError}
				<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
			{/if}
		</div>
	</form>
</AppDialog>
