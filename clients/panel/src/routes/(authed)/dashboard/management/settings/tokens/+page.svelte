<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ConfirmAction from '$lib/components/remnawave/ConfirmAction.svelte';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import EmptyState from '$lib/components/shared/EmptyState.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { copyText } from '$lib/copy';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { formatDate } from '$lib/format';
	import { appToast } from '$lib/features/toast/toast';
	import { asArray } from '$lib/list';
	import { rw } from '$lib/rw';
	import type { ApiToken } from '@arcanum/ts-client';

	const list = rw.tokens.list();
	const createM = rw.tokens.create();
	const removeM = rw.tokens.remove();

	let openCreate = $state(false);
	let name = $state('');
	let days = $state('365');
	let createdToken = $state('');
	let confirmOpen = $state(false);
	let pendingUuid = $state('');

	const tokens = $derived(asArray<ApiToken>(list.data, ['tokens']));

	$effect.pre(() => {
		const token = pageChrome.set({
			title: 'API tokens',
			create: { label: 'Create token', onclick: () => (openCreate = true) }
		});
		return () => pageChrome.clear(token);
	});

	const actions = $derived([{ id: 'create', label: 'Create', disabled: createM.pending || !name.trim() }]);

	async function create() {
		if (!name.trim()) return;
		try {
			const token = await createM.mutate({
				name: name.trim(),
				expiresInDays: Number(days) || 365,
				scopes: ['*']
			});
			createdToken = token.token ?? '';
			name = '';
			openCreate = false;
			appToast.success('Token created');
			await list.refetch();
		} catch (err) {
			appToast.apiError(err, 'Create failed');
		}
	}
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">API tokens</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">
	Bearer tokens for automation. The secret is shown once at create time.
</p>

{#if createdToken}
	<div class="mt-6 rounded-lg border border-[var(--app-accent)]/40 bg-[var(--color-bg-secondary)] px-5 py-4">
		<p class="text-sm font-medium text-[var(--color-text-primary)]">Copy this token now</p>
		<p class="mt-1 text-xs text-[var(--color-text-tertiary)]">It will not be shown again.</p>
		<pre class="mt-3 overflow-auto font-mono text-xs text-[var(--color-text-secondary)]">{createdToken}</pre>
		<Button
			class="mt-3"
			size="sm"
			variant="outline"
			onclick={async () => {
				if (await copyText(createdToken)) appToast.success('Token copied');
			}}
		>
			Copy token
		</Button>
	</div>
{/if}

{#if list.loading && !list.data}
	<div class="mt-8"><LoadingState /></div>
{:else if list.error && !list.data}
	<div class="mt-8"><ErrorState message={list.error.message} onretry={() => list.refetch()} /></div>
{:else if tokens.length === 0}
	<div class="mt-8">
		<EmptyState title="No API tokens" action={{ label: 'Create token', onclick: () => (openCreate = true) }} />
	</div>
{:else}
	<SettingsCard>
		{#each tokens as token (token.uuid)}
			<SettingsRow title={token.name} description={`Expires ${formatDate(token.expireAt)}`}>
				<Button
					size="sm"
					variant="destructive"
					onclick={() => {
						pendingUuid = token.uuid;
						confirmOpen = true;
					}}
				>
					Revoke
				</Button>
			</SettingsRow>
		{/each}
	</SettingsCard>
{/if}

<AppDialog
	bind:open={openCreate}
	title="Create API token"
	description="Full-access token (*). Store the secret after create."
	{actions}
	pending={createM.pending}
	onrun={create}
>
	<div class={chrome.stack}>
		<label class={chrome.field}>
			<span class={chrome.label}>Name</span>
			<Input bind:value={name} placeholder="ci-bot" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Expires in days</span>
			<Input bind:value={days} inputmode="numeric" />
		</label>
	</div>
</AppDialog>

<ConfirmAction
	bind:open={confirmOpen}
	title="Revoke this token?"
	description="Scripts using it will fail on the next request."
	confirmLabel="Revoke"
	onconfirm={async () => {
		await removeM.mutate(pendingUuid);
		appToast.success('Revoked');
		await list.refetch();
	}}
/>
