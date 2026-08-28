<script lang="ts">
	import { Switch } from '$lib/components/ui/switch';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';

	type SubSettings = {
		uuid?: string;
		serveJsonAtBaseSubscription?: boolean;
		isShowCustomRemarks?: boolean;
		randomizeHosts?: boolean;
	};

	const q = rw.subscription.settings.get();
	const updateM = rw.subscription.settings.update();

	let serveJson = $state(false);
	let customRemarks = $state(false);
	let randomizeHosts = $state(false);
	let loaded = $state(false);
	const uuid = $derived((q.data as SubSettings | undefined)?.uuid ?? '');

	$effect.pre(() => {
		const token = pageChrome.set({
			title: 'Subscription',
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
		serveJson = Boolean(data.serveJsonAtBaseSubscription);
		customRemarks = Boolean(data.isShowCustomRemarks);
		randomizeHosts = Boolean(data.randomizeHosts);
		loaded = true;
	});

	async function save() {
		try {
			await updateM.mutate({
				uuid,
				serveJsonAtBaseSubscription: serveJson,
				isShowCustomRemarks: customRemarks,
				randomizeHosts
			});
			appToast.success('Subscription settings saved');
			loaded = false;
			await q.refetch();
		} catch (err) {
			appToast.apiError(err, 'Save failed');
		}
	}
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Subscription</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">
	How public subscription URLs behave. Response rules have their own page.
</p>

{#if q.loading && !q.data}
	<div class="mt-8"><LoadingState /></div>
{:else if q.error && !q.data}
	<div class="mt-8"><ErrorState message={q.error.message} onretry={() => q.refetch()} /></div>
{:else}
	<SettingsCard>
		<SettingsRow
			title="Serve JSON at base URL"
			description="Return Xray JSON when the client hits the subscription URL without a client suffix."
		>
			<Switch size="sm" bind:checked={serveJson} />
		</SettingsRow>
		<SettingsRow title="Custom remarks" description="Use custom remark templates in the client list.">
			<Switch size="sm" bind:checked={customRemarks} />
		</SettingsRow>
		<SettingsRow title="Randomize hosts" description="Shuffle host order in generated subscriptions.">
			<Switch size="sm" bind:checked={randomizeHosts} />
		</SettingsRow>
	</SettingsCard>
{/if}
