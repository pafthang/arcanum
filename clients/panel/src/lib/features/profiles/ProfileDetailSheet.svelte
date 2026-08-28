<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { pretty } from '$lib/list';
	import { errorMessage, parseJsonObject, profileHasInbound } from '$lib/json-edit';
	import type { ConfigProfile } from '@optimawave/ts-back';

	let {
		profile = $bindable(null),
		pending = false,
		onupdate,
		ondelete
	}: {
		profile: ConfigProfile | null;
		pending?: boolean;
		onupdate: (body: { uuid: string; name: string; config: unknown }) => Promise<void>;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let draft = $state('{}');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');
	let parseError = $state('');

	$effect(() => {
		const next = profile;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		draft = pretty(next.config);
		parseError = '';
	});

	const open = $derived(profile !== null);
	const inboundTags = $derived((profile?.inbounds ?? []).map((inbound) => inbound.tag).join(' · '));
	const actions = [{ id: 'save', label: 'Save changes' }];
	const menu = $derived(profile ? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }] : []);

	async function runAction() {
		if (!profile) return;
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
			await onupdate({ uuid: profile.uuid, name: name.trim() || profile.name, config: parsed.value });
		} catch (err) {
			parseError = errorMessage(err, 'Save failed');
		}
	}
</script>

<DetailSheet
	{open}
	title={profile ? `Edit ${profile.name}` : 'Edit profile'}
	description={inboundTags || 'No inbounds yet'}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			profile = null;
			loadedUuid = null;
		}
	}}
>
	{#if profile}
		<div class={chrome.stack}>
			<section class="grid grid-cols-2 gap-3">
				<div class={chrome.tile}>
					<p class={chrome.hint}>Inbounds</p>
					<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{profile.inbounds?.length ?? 0}</p>
				</div>
				<div class={chrome.tile}>
					<p class={chrome.hint}>Nodes</p>
					<p class="mt-1 text-sm font-medium text-[var(--color-text-primary)]">{profile.nodes?.length ?? 0}</p>
				</div>
			</section>
			<label class={chrome.field}>
				<span class={chrome.label}>Name</span>
				<Input bind:value={name} />
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
