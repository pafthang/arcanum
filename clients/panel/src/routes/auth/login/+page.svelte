<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { LogIn } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Password } from '$lib/components/ui/password';
	import { Separator } from '$lib/components/ui/separator';
	import AuthShell from '$lib/components/auth/AuthShell.svelte';
	import BootScreen from '$lib/components/auth/BootScreen.svelte';
	import { appToast } from '$lib/features/toast/toast';
	import { APP_NAME } from '$lib/brand';
	import { rw, session } from '$lib/rw';

	const statusQ = rw.auth.status();
	const loginM = rw.auth.login();
	const oauthM = rw.auth.oauth2Authorize();
	const passkeyM = rw.auth.passkeyAuthenticationVerify();
	const passkeyOpts = rw.auth.passkeyAuthenticationOptions();

	let username = $state('');
	let password = $state('');

	const status = $derived(statusQ.data);
	const passwordEnabled = $derived(status?.authentication?.password?.enabled ?? true);
	const registerAllowed = $derived(status?.isRegisterAllowed ?? false);
	const loginAllowed = $derived(status?.isLoginAllowed ?? true);
	const passkeyEnabled = $derived(status?.authentication?.passkey?.enabled ?? false);
	const oauth = $derived(status?.authentication?.oauth2?.providers);
	const title = $derived(status?.branding?.title || APP_NAME);
	const oauthProviders = $derived(
		oauth ? Object.entries(oauth).filter(([, enabled]) => Boolean(enabled)).map(([name]) => name) : []
	);
	const hasAlt = $derived(passkeyEnabled || oauthProviders.length > 0);
	const firstAdmin = $derived(Boolean(status && !loginAllowed && registerAllowed));

	onMount(() => {
		session.init();
		if (session.isAuthenticated) void goto('/dashboard/home', { replaceState: true });
	});

	$effect(() => {
		if (firstAdmin) void goto('/auth/register', { replaceState: true });
	});

	async function submit(e: Event) {
		e.preventDefault();
		try {
			await loginM.mutate({ username, password });
			void goto('/dashboard/home', { replaceState: true });
		} catch (err) {
			appToast.apiError(err, 'Sign in failed');
		}
	}

	async function oauth2(provider: string) {
		try {
			const res = await oauthM.mutate({ provider });
			if (res.authorizationUrl) window.location.assign(res.authorizationUrl);
		} catch (err) {
			appToast.apiError(err, 'OAuth2 failed');
		}
	}

	async function passkey() {
		try {
			const options = (await passkeyOpts.mutate()) as
				| { publicKey?: PublicKeyCredentialRequestOptions }
				| PublicKeyCredentialRequestOptions;
			const publicKey = (
				options && 'publicKey' in options ? options.publicKey : options
			) as PublicKeyCredentialRequestOptions | undefined;
			if (!publicKey) throw new Error('No passkey challenge');
			const cred = (await navigator.credentials.get({ publicKey })) as PublicKeyCredential | null;
			if (!cred) return;
			await passkeyM.mutate({ response: cred.toJSON ? cred.toJSON() : cred });
			void goto('/dashboard/home', { replaceState: true });
		} catch (err) {
			appToast.apiError(err, 'Passkey failed');
		}
	}
</script>

{#if (statusQ.loading && !status) || firstAdmin}
	<BootScreen message={firstAdmin ? 'Opening setup…' : 'Checking authentication…'} />
{:else}
	<AuthShell {title} subtitle="Sign in to the panel">
		{#if statusQ.error}
			<div
				class="rounded-lg border border-[var(--color-error)]/30 bg-[var(--color-error)]/10 px-3 py-2 text-center text-sm text-[var(--color-error)]"
			>
				Server is not responding. Check logs.
			</div>
		{:else if !loginAllowed && !registerAllowed}
			<p class="text-center text-sm text-[var(--color-text-secondary)]">Sign in is disabled on this panel.</p>
		{:else}
			{#if passwordEnabled && loginAllowed}
				<form class="space-y-3.5" onsubmit={submit}>
					<div>
						<label class="text-[11px] font-medium text-[var(--color-text-tertiary)]" for="username">Username</label>
						<Input
							id="username"
							bind:value={username}
							required
							autocomplete="username"
							placeholder="admin"
							class="mt-1"
						/>
					</div>
					<div>
						<label class="text-[11px] font-medium text-[var(--color-text-tertiary)]" for="password">Password</label>
						<Password
							id="password"
							bind:value={password}
							required
							autocomplete="current-password"
							placeholder="Your password"
							class="mt-1"
						/>
					</div>
					<Button class="mt-2 w-full" type="submit" disabled={loginM.pending}>
						<LogIn class="size-4" />
						{loginM.pending ? 'Signing in…' : 'Sign in'}
					</Button>
				</form>
			{/if}

			{#if passwordEnabled && loginAllowed && hasAlt}
				<div class="my-5 flex items-center gap-3">
					<Separator class="flex-1 bg-[var(--app-border)]" />
					<span class="text-[11px] tracking-wide text-[var(--color-text-tertiary)] uppercase">Or</span>
					<Separator class="flex-1 bg-[var(--app-border)]" />
				</div>
			{/if}

			{#if hasAlt}
				<div class="grid gap-2">
					{#if passkeyEnabled}
						<Button class="w-full" variant="outline" onclick={passkey} disabled={passkeyM.pending}>Passkey</Button>
					{/if}
					{#each oauthProviders as provider}
						<Button variant="outline" class="w-full capitalize" onclick={() => oauth2(provider)}>
							Continue with {provider}
						</Button>
					{/each}
				</div>
			{/if}

			{#if registerAllowed && passwordEnabled}
				<p class="mt-5 text-center text-xs text-[var(--color-text-tertiary)]">
					First-time setup?
					<a href="/auth/register" class="text-[var(--app-accent-light)] hover:underline">Create the admin account</a>
				</p>
			{/if}
		{/if}
	</AuthShell>
{/if}
