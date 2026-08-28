<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import { Switch } from '$lib/components/ui/switch';
	import SettingsCard from '$lib/components/remnawave/SettingsCard.svelte';
	import SettingsRow from '$lib/components/remnawave/SettingsRow.svelte';
	import ErrorState from '$lib/components/shared/ErrorState.svelte';
	import LoadingState from '$lib/components/shared/LoadingState.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { rw } from '$lib/rw';

	type OauthProvider = {
		enabled?: boolean;
		clientId?: string | null;
		clientSecret?: string | null;
		allowedEmails?: string[];
		allowedIds?: string[];
		frontendDomain?: string | null;
		plainDomain?: string | null;
		realm?: string | null;
		keycloakDomain?: string | null;
		authorizationUrl?: string | null;
		tokenUrl?: string | null;
		withPkce?: boolean;
	};

	type Settings = {
		passwordSettings?: { enabled?: boolean };
		passkeySettings?: { enabled?: boolean; rpId?: string | null; origin?: string | null };
		oauth2Settings?: Record<string, OauthProvider>;
	};

	const PROVIDERS = [
		{ key: 'github', label: 'GitHub', hint: 'client id, secret, allowed emails' },
		{ key: 'yandex', label: 'Yandex', hint: 'client id, secret, allowed emails' },
		{ key: 'pocketid', label: 'Pocket ID', hint: 'client id, secret, domains' },
		{ key: 'keycloak', label: 'Keycloak', hint: 'realm, client id, secret, domains' },
		{ key: 'telegram', label: 'Telegram', hint: 'bot token, allowed telegram ids' },
		{ key: 'generic', label: 'Generic OIDC', hint: 'authorization/token URLs, PKCE' }
	] as const;

	const q = rw.settings.get();
	const updateM = rw.settings.update();

	let passwordEnabled = $state(true);
	let passkeyEnabled = $state(false);
	let rpId = $state('');
	let origin = $state('');
	let oauth = $state<Record<string, OauthProvider>>({});
	let loaded = $state(false);

	$effect.pre(() => {
		const token = pageChrome.set({
			title: 'Authentication',
			save: {
				onclick: () => void save(),
				disabled: () => updateM.pending,
				pending: () => updateM.pending
			}
		});
		return () => pageChrome.clear(token);
	});

	$effect(() => {
		const data = q.data as Settings | undefined;
		if (!data || loaded) return;
		passwordEnabled = data.passwordSettings?.enabled ?? true;
		passkeyEnabled = data.passkeySettings?.enabled ?? false;
		rpId = data.passkeySettings?.rpId ?? '';
		origin = data.passkeySettings?.origin ?? '';
		oauth = structuredClone(data.oauth2Settings ?? {});
		loaded = true;
	});

	function provider(key: string): OauthProvider {
		return oauth[key] ?? {};
	}

	function setProvider(key: string, patch: Partial<OauthProvider>) {
		oauth = { ...oauth, [key]: { ...provider(key), ...patch } };
	}

	function emailsText(key: string): string {
		return (provider(key).allowedEmails ?? []).join(', ');
	}

	function idsText(key: string): string {
		return (provider(key).allowedIds ?? []).join(', ');
	}

	async function save() {
		try {
			await updateM.mutate({
				passwordSettings: { enabled: passwordEnabled },
				passkeySettings: {
					enabled: passkeyEnabled,
					rpId: rpId.trim() || null,
					origin: origin.trim() || null
				},
				oauth2Settings: oauth
			});
			appToast.success('Authentication saved');
			loaded = false;
			await q.refetch();
		} catch (err) {
			appToast.apiError(err, 'Save failed');
		}
	}
</script>

<h1 class="text-2xl font-semibold text-[var(--color-text-primary)]">Authentication</h1>
<p class="mt-1 text-sm text-[var(--color-text-tertiary)]">
	At least one method must stay enabled. OAuth providers need client id and secret when turned on.
</p>

{#if q.loading && !q.data}
	<div class="mt-8"><LoadingState /></div>
{:else if q.error && !q.data}
	<div class="mt-8"><ErrorState message={q.error.message} onretry={() => q.refetch()} /></div>
{:else}
	<SettingsCard title="Password">
		<SettingsRow title="Password login" description="Username and password on the sign-in page.">
			<Switch size="sm" bind:checked={passwordEnabled} />
		</SettingsRow>
	</SettingsCard>

	<SettingsCard title="Passkey">
		<SettingsRow title="Passkeys" description="WebAuthn. RP ID and origin are required when enabled.">
			<Switch size="sm" bind:checked={passkeyEnabled} />
		</SettingsRow>
		{#if passkeyEnabled}
			<SettingsRow title="RP ID" description="Usually the panel hostname.">
				<Input class="w-[220px]" bind:value={rpId} placeholder="panel.example.com" />
			</SettingsRow>
			<SettingsRow title="Origin" description="Full origin, including https://">
				<Input class="w-[240px]" bind:value={origin} placeholder="https://panel.example.com" />
			</SettingsRow>
		{/if}
	</SettingsCard>

	{#each PROVIDERS as item}
		<SettingsCard title={item.label} description={item.hint}>
			<SettingsRow title="Enabled">
				<Switch
					size="sm"
					checked={Boolean(provider(item.key).enabled)}
					onCheckedChange={(value) => setProvider(item.key, { enabled: value })}
				/>
			</SettingsRow>
			{#if provider(item.key).enabled}
				<SettingsRow title="Client ID">
					<Input
						class="w-[220px]"
						value={provider(item.key).clientId ?? ''}
						oninput={(event) => setProvider(item.key, { clientId: event.currentTarget.value })}
					/>
				</SettingsRow>
				<SettingsRow title="Client secret">
					<Input
						class="w-[220px]"
						type="password"
						value={provider(item.key).clientSecret ?? ''}
						oninput={(event) => setProvider(item.key, { clientSecret: event.currentTarget.value })}
					/>
				</SettingsRow>
				{#if item.key === 'telegram'}
					<SettingsRow title="Allowed IDs" description="Comma-separated Telegram user ids.">
						<Input
							class="w-[240px]"
							value={idsText(item.key)}
							oninput={(event) =>
								setProvider(item.key, {
									allowedIds: event.currentTarget.value
										.split(',')
										.map((part) => part.trim())
										.filter(Boolean)
								})}
						/>
					</SettingsRow>
				{:else}
					<SettingsRow title="Allowed emails" description="Comma-separated. Required for GitHub and Yandex.">
						<Input
							class="w-[240px]"
							value={emailsText(item.key)}
							oninput={(event) =>
								setProvider(item.key, {
									allowedEmails: event.currentTarget.value
										.split(',')
										.map((part) => part.trim())
										.filter(Boolean)
								})}
						/>
					</SettingsRow>
				{/if}
			{/if}
		</SettingsCard>
	{/each}

{/if}
