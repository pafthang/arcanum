<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { copyText } from '$lib/copy';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';

	const keygen = rw.settings.keygen();

	$effect.pre(() => {
		const token = pageChrome.set({ title: 'Node keys' });
		return () => pageChrome.clear(token);
	});

	async function copy(label: string, value: string) {
		if (await copyText(value)) appToast.success(`${label} copied`);
		else appToast.error(`Could not copy ${label.toLowerCase()}`);
	}
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Node keys</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">
	SECRET_KEY used by remnanode. Paste the secret into the node environment.
</p>

{#if keygen.loading && !keygen.data}
	<div class="mt-8"><LoadingState /></div>
{:else if keygen.error && !keygen.data}
	<div class="mt-8"><ErrorState message={keygen.error.message} onretry={() => keygen.refetch()} /></div>
{:else}
	<SettingsCard>
		<SettingsRow title="Public key" description="Safe to share with nodes that verify the panel.">
			<Button
				size="sm"
				variant="outline"
				onclick={() => copy('Public key', keygen.data?.pubKey ?? '')}
				disabled={!keygen.data?.pubKey}
			>
				Copy
			</Button>
		</SettingsRow>
		<SettingsRow title="Secret key" description="Keep this private. Required on every node.">
			<Button
				size="sm"
				variant="outline"
				onclick={() => copy('Secret key', keygen.data?.secretKey ?? '')}
				disabled={!keygen.data?.secretKey}
			>
				Copy
			</Button>
		</SettingsRow>
	</SettingsCard>
	{#if keygen.data?.pubKey}
		<pre class="mt-4 overflow-auto rounded-lg border border-[var(--app-border)] bg-[var(--color-bg-secondary)] p-4 font-mono text-[11px] text-[var(--color-text-secondary)]">{keygen.data.pubKey}</pre>
	{/if}
{/if}
