<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { Shuffle, UserPlus } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Password } from '$lib/components/ui/password';
	import AuthShell from '$lib/components/auth/AuthShell.svelte';
	import BootScreen from '$lib/components/auth/BootScreen.svelte';
	import { generatePassword, PASSWORD_POLICY, passwordLooksValid } from '$lib/components/auth/password';
	import { appToast } from '$lib/features/toast/toast';
	import { APP_NAME } from '$lib/brand';
	import { rw, session } from '$lib/rw';

	const statusQ = rw.auth.status();
	const registerM = rw.auth.register();

	let username = $state('');
	let password = $state('');
	let confirm = $state('');

	const status = $derived(statusQ.data);
	const registerAllowed = $derived(status?.isRegisterAllowed ?? false);
	const title = $derived(status?.branding?.title || APP_NAME);
	const mismatch = $derived(confirm.length > 0 && confirm !== password);
	const canSubmit = $derived(username.length > 0 && passwordLooksValid(password) && password === confirm);

	onMount(() => {
		session.init();
		if (session.isAuthenticated) void goto('/dashboard/home', { replaceState: true });
	});

	$effect(() => {
		if (status && !registerAllowed) void goto('/auth/login', { replaceState: true });
	});

	async function submit(e: Event) {
		e.preventDefault();
		if (!canSubmit) return;
		try {
			await registerM.mutate({ username, password });
			void goto('/dashboard/home', { replaceState: true });
		} catch (err) {
			appToast.apiError(err, 'Registration failed');
		}
	}

	async function generate() {
		const next = generatePassword();
		password = next;
		confirm = next;
		try {
			await navigator.clipboard.writeText(next);
			appToast.success('Password generated and copied');
		} catch {
			appToast.success('Password generated');
		}
	}
</script>

{#if (statusQ.loading && !status) || (status && !registerAllowed)}
	<BootScreen message={status && !registerAllowed ? 'Redirecting…' : 'Checking authentication…'} />
{:else}
	<AuthShell {title} subtitle="Create the first administrator">
		{#if statusQ.error}
			<div
				class="rounded-lg border border-[var(--color-error)]/30 bg-[var(--color-error)]/10 px-3 py-2 text-center text-sm text-[var(--color-error)]"
			>
				Server is not responding. Check logs.
			</div>
		{:else}
			<p class="mb-5 text-center text-sm text-[var(--color-text-secondary)]">
				This panel has no admin yet. Set a username and a strong password to finish setup.
			</p>
			<form class="space-y-3.5" onsubmit={submit}>
				<div>
					<label class="text-[11px] font-medium text-[var(--color-text-tertiary)]" for="username">Username</label>
					<Input
						id="username"
						bind:value={username}
						required
						autocomplete="username"
						placeholder="IamSuperAdmin"
						class="mt-1"
					/>
				</div>
				<div>
					<label class="text-[11px] font-medium text-[var(--color-text-tertiary)]" for="password">Password</label>
					<Password id="password" bind:value={password} required autocomplete="new-password" class="mt-1" />
					<p class="mt-1 text-[11px] text-[var(--color-text-tertiary)]">{PASSWORD_POLICY}</p>
				</div>
				<div>
					<label class="text-[11px] font-medium text-[var(--color-text-tertiary)]" for="confirm">Confirm password</label>
					<Password id="confirm" bind:value={confirm} required autocomplete="new-password" class="mt-1" />
					{#if mismatch}
						<p class="mt-1 text-[11px] text-[var(--color-error)]">Passwords do not match.</p>
					{/if}
				</div>
				<Button class="w-full" type="button" variant="outline" onclick={generate}>
					<Shuffle class="size-4" />
					Generate password
				</Button>
				<Button class="w-full" type="submit" disabled={registerM.pending || !canSubmit}>
					<UserPlus class="size-4" />
					{registerM.pending ? 'Creating account…' : 'Create admin'}
				</Button>
			</form>
		{/if}
	</AuthShell>
{/if}
