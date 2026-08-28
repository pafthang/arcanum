<script lang="ts">
	import { onMount } from 'svelte';
	import { Input } from '$lib/components/ui/input';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';

	const q = rw.settings.get();
	const updateM = rw.settings.update();

	let title = $state('');
	let logoUrl = $state('');
	let loaded = $state(false);

	$effect.pre(() => {
		const token = pageChrome.set({
			title: 'Branding',
			save: {
				onclick: () => void save(),
				disabled: () => updateM.pending,
				pending: () => updateM.pending
			}
		});
		return () => pageChrome.clear(token);
	});

	$effect(() => {
		const data = q.data as { brandingSettings?: { title?: string | null; logoUrl?: string | null } } | undefined;
		if (!data || loaded) return;
		title = data.brandingSettings?.title ?? '';
		logoUrl = data.brandingSettings?.logoUrl ?? '';
		loaded = true;
	});

	async function save() {
		try {
			await updateM.mutate({
				brandingSettings: {
					title: title.trim() || null,
					logoUrl: logoUrl.trim() || null
				}
			});
			appToast.success('Branding saved');
			loaded = false;
			await q.refetch();
		} catch (err) {
			appToast.apiError(err, 'Save failed');
		}
	}

	onMount(() => {
		void q.refetch();
	});
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Branding</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">Title and logo shown on the login screen.</p>

{#if q.loading && !q.data}
	<div class="mt-8"><LoadingState /></div>
{:else if q.error && !q.data}
	<div class="mt-8"><ErrorState message={q.error.message} onretry={() => q.refetch()} /></div>
{:else}
	<SettingsCard>
		<SettingsRow title="Panel title" description="Replaces OptimaWave on the sign-in page.">
			<Input class="w-[220px]" bind:value={title} placeholder="OptimaWave" />
		</SettingsRow>
		<SettingsRow title="Logo URL" description="HTTPS image used next to the title.">
			<Input class="w-[240px]" type="url" bind:value={logoUrl} placeholder="https://" />
		</SettingsRow>
	</SettingsCard>
{/if}
