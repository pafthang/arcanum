<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import AppDialog from '$lib/components/remnawave/AppDialog.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import { dateInputToExpireIso, plusDaysIso, toDateInput, type TrafficUnit } from '$lib/format';

	let {
		open = $bindable(false),
		pending = false,
		oncreate
	}: {
		open: boolean;
		pending?: boolean;
		oncreate: (body: {
			username: string;
			expireAt: string;
			trafficLimitBytes: number;
			tag?: string;
		}) => Promise<void>;
	} = $props();

	let username = $state('');
	let expireDate = $state(toDateInput(plusDaysIso(30)));
	let amount = $state('0');
	let unit = $state<TrafficUnit>('GB');
	let tag = $state('');
	let actionId = $state('create');

	function reset() {
		username = '';
		expireDate = toDateInput(plusDaysIso(30));
		amount = '0';
		unit = 'GB';
		tag = '';
		actionId = 'create';
	}

	async function submit() {
		if (!username.trim() || !expireDate) return;
		const n = Number(amount);
		const mul = { MB: 1024 ** 2, GB: 1024 ** 3, TB: 1024 ** 4 };
		await oncreate({
			username: username.trim(),
			expireAt: dateInputToExpireIso(expireDate),
			trafficLimitBytes: !Number.isFinite(n) || n <= 0 ? 0 : Math.round(n * mul[unit]),
			tag: tag.trim() || undefined
		});
		reset();
		open = false;
	}

	const actions = $derived([
		{ id: 'create', label: 'Create', disabled: pending || !username.trim() || !expireDate }
	]);
</script>

<AppDialog
	bind:open
	title="Create user"
	description="Username and expiration are required. Secrets are generated."
	{actions}
	bind:actionId
	{pending}
	onrun={submit}
>
	<form
		class={chrome.stack}
		onsubmit={(event) => {
			event.preventDefault();
			void submit();
		}}
	>
		<label class={chrome.field}>
			<span class={chrome.label}>Username</span>
			<Input bind:value={username} required autocomplete="off" placeholder="alice" />
		</label>
		<label class={chrome.field}>
			<span class={chrome.label}>Expires on</span>
			<Input type="date" bind:value={expireDate} required />
		</label>
		<div class={chrome.field}>
			<span class={chrome.label}>Traffic limit</span>
			<div class="flex gap-2">
				<Input bind:value={amount} inputmode="decimal" placeholder="0 = unlimited" />
				<Select.Root type="single" value={unit} onValueChange={(value) => value && (unit = value as TrafficUnit)}>
					<Select.Trigger class="w-24" size="sm">{unit}</Select.Trigger>
					<Select.Content>
						<Select.Item value="MB">MB</Select.Item>
						<Select.Item value="GB">GB</Select.Item>
						<Select.Item value="TB">TB</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>
			<p class={chrome.hint}>0 means unlimited.</p>
		</div>
		<label class={chrome.field}>
			<span class={chrome.label}>Tag</span>
			<Input bind:value={tag} placeholder="optional" />
		</label>
	</form>
</AppDialog>
