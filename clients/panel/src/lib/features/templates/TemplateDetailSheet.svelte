<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Textarea } from '$lib/components/ui/textarea';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { pretty } from '$lib/list';

	const YAML_TYPES = new Set(['CLASH', 'MIHOMO', 'STASH']);

	export type TemplateRecord = {
		uuid: string;
		name: string;
		templateType: string;
		templateJson?: unknown;
		encodedTemplateYaml?: string | null;
	};

	let {
		template = $bindable(null),
		pending = false,
		onupdate,
		ondelete
	}: {
		template: TemplateRecord | null;
		pending?: boolean;
		onupdate: (body: {
			uuid: string;
			name: string;
			templateJson?: unknown;
			encodedTemplateYaml?: string;
		}) => Promise<void>;
		ondelete: () => void;
	} = $props();

	let name = $state('');
	let draft = $state('');
	let loadedUuid = $state<string | null>(null);
	let actionId = $state('save');
	let parseError = $state('');

	const yaml = $derived(template ? YAML_TYPES.has(template.templateType) : false);

	function decodeYaml(encoded: string | null | undefined): string {
		if (!encoded) return '';
		try {
			const bin = atob(encoded);
			return new TextDecoder().decode(Uint8Array.from(bin, (ch) => ch.charCodeAt(0)));
		} catch {
			return encoded;
		}
	}

	function encodeYaml(text: string): string {
		const bytes = new TextEncoder().encode(text);
		let bin = '';
		bytes.forEach((byte) => (bin += String.fromCharCode(byte)));
		return btoa(bin);
	}

	$effect(() => {
		const next = template;
		if (!next || loadedUuid === next.uuid) return;
		loadedUuid = next.uuid;
		actionId = 'save';
		name = next.name;
		parseError = '';
		draft = YAML_TYPES.has(next.templateType) ? decodeYaml(next.encodedTemplateYaml) : pretty(next.templateJson ?? {});
	});

	const open = $derived(template !== null);
	const actions = [{ id: 'save', label: 'Save changes' }];
	const menu = $derived(
		template && template.name !== 'Default'
			? [{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }]
			: []
	);

	async function runAction() {
		if (!template) return;
		if (yaml) {
			await onupdate({
				uuid: template.uuid,
				name: name.trim() || template.name,
				encodedTemplateYaml: encodeYaml(draft)
			});
			return;
		}
		try {
			const templateJson = JSON.parse(draft || '{}');
			parseError = '';
			await onupdate({ uuid: template.uuid, name: name.trim() || template.name, templateJson });
		} catch {
			parseError = 'Template must be valid JSON.';
		}
	}
</script>

<DetailSheet
	{open}
	title={template ? `Edit ${template.name}` : 'Edit template'}
	description={template?.templateType ?? ''}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			template = null;
			loadedUuid = null;
		}
	}}
>
	{#if template}
		<div class={chrome.stack}>
			<label class={chrome.field}>
				<span class={chrome.label}>Name</span>
				<Input bind:value={name} disabled={template.name === 'Default'} />
			</label>
			<div class={chrome.field}>
				<span class={chrome.label}>{yaml ? 'YAML' : 'JSON'}</span>
				{#if yaml}
					<Textarea bind:value={draft} rows={16} class="font-mono text-xs" />
				{:else}
					<JsonField bind:value={draft} rows={16} />
				{/if}
				{#if parseError}
					<p class="text-[12px] text-[var(--color-error)]">{parseError}</p>
				{/if}
			</div>
		</div>
	{/if}
</DetailSheet>
