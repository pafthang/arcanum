<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import JsonField from '$lib/components/remnawave/JsonField.svelte';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { pretty } from '$lib/list';
	import { errorMessage, parseJsonObject } from '$lib/json-edit';
	import { rw } from '$lib/rw';

	type SubSettings = { uuid?: string; responseRules?: unknown };

	const q = rw.subscription.settings.get();
	const updateM = rw.subscription.settings.update();
	const matcher = rw.system.srrMatcher();

	let draft = $state('{}');
	let loaded = $state(false);
	let parseError = $state('');
	let userAgent = $state('v2rayN/6.0');
	let result = $state<unknown>(null);
	let testing = $state(false);
	const uuid = $derived((q.data as SubSettings | undefined)?.uuid ?? '');

	$effect.pre(() => {
		const token = pageChrome.set({
			title: 'Response rules',
			save: {
				onclick: () => void save(),
				disabled: () => updateM.pending || !uuid,
				pending: () => updateM.pending
			}
		});
		return () => pageChrome.clear(token);
	});

	$effect(() => {
		const data = q.data as SubSettings | undefined;
		if (!data || loaded) return;
		draft = pretty(data.responseRules ?? {});
		loaded = true;
	});

	async function save() {
		try {
			const parsed = parseJsonObject(draft);
			if (!parsed.ok) {
				parseError = parsed.error;
				return;
			}
			parseError = '';
			await updateM.mutate({ uuid, responseRules: parsed.value });
			appToast.success('Response rules saved');
			loaded = false;
			await q.refetch();
		} catch (err) {
			parseError = errorMessage(err, 'Save failed');
			appToast.apiError(err, 'Save failed');
		}
	}

	async function test() {
		testing = true;
		try {
			const parsed = parseJsonObject(draft);
			if (!parsed.ok) {
				parseError = parsed.error;
				return;
			}
			parseError = '';
			result = await matcher.mutate({ responseRules: parsed.value, userAgent });
		} catch (err) {
			appToast.apiError(err, 'Test failed');
		} finally {
			testing = false;
		}
	}
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Response rules</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">
	SRR maps a client User-Agent to a subscription format. Test against the draft JSON below.
</p>

{#if q.loading && !q.data}
	<div class="mt-8"><LoadingState /></div>
{:else if q.error && !q.data}
	<div class="mt-8"><ErrorState message={q.error.message} onretry={() => q.refetch()} /></div>
{:else}
	<div class="mt-8">
		<p class="text-sm font-medium text-[var(--color-text-secondary)]">Rules JSON</p>
		<div class="mt-3">
			<JsonField bind:value={draft} rows={16} />
		</div>
		{#if parseError}
			<p class="mt-2 text-[12px] text-[var(--color-error)]">{parseError}</p>
		{/if}
	</div>

	<SettingsCard title="Matcher">
		<SettingsRow title="User-Agent" description="Sent with the rules JSON. Does not save.">
			<div class="flex gap-2">
				<Input class="w-[220px]" bind:value={userAgent} />
				<Button size="sm" variant="outline" onclick={test} disabled={testing}>Test</Button>
			</div>
		</SettingsRow>
	</SettingsCard>

	{#if result}
		<pre class="mt-4 overflow-auto rounded-md border border-[var(--app-border)] p-3 font-mono text-xs text-[var(--color-text-secondary)]">{pretty(result)}</pre>
	{/if}
{/if}
