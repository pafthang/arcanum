<script lang="ts">
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import DetailSheet from '$lib/components/remnawave/DetailSheet.svelte';
	import { chrome } from '$lib/components/remnawave/chrome';
	import StatusBadge from '$lib/components/remnawave/StatusBadge.svelte';
	import ExpireLabel from '$lib/components/remnawave/ExpireLabel.svelte';
	import TrafficBar from '$lib/components/remnawave/TrafficBar.svelte';
	import { copyText } from '$lib/copy';
	import { appToast } from '$lib/features/toast/toast';
	import {
		dateInputToExpireIso,
		splitTraffic,
		toDateInput,
		trafficToBytes,
		type TrafficUnit
	} from '$lib/format';
	import type { User } from '@arcanum/ts-client';

	let {
		user = $bindable(null),
		pending = false,
		onupdate,
		onenable,
		ondisable,
		onextend,
		onreset,
		onrevoke,
		ondelete
	}: {
		user: User | null;
		pending?: boolean;
		onupdate: (body: { id: number; expireAt: string; tag?: string; trafficLimitBytes: number }) => Promise<void>;
		onenable: () => void;
		ondisable: () => void;
		onextend: () => void;
		onreset: () => void;
		onrevoke: () => void;
		ondelete: () => void;
	} = $props();

	let expireDate = $state('');
	let tag = $state('');
	let amount = $state('0');
	let unit = $state<TrafficUnit>('GB');
	let loadedId = $state<number | null>(null);
	let actionId = $state('save');

	$effect(() => {
		const next = user;
		if (!next) return;
		if (loadedId === next.id) return;
		loadedId = next.id;
		actionId = 'save';
		expireDate = toDateInput(next.expireAt);
		tag = next.tag ?? '';
		const split = splitTraffic(next.trafficLimitBytes ?? 0);
		amount = split.amount;
		unit = split.unit;
	});

	const open = $derived(user !== null);

	const actions = $derived(
		user
			? [
					{ id: 'save', label: 'Save changes' },
					user.status === 'DISABLED'
						? { id: 'enable', label: 'Enable' }
						: { id: 'disable', label: 'Disable' },
					{ id: 'extend', label: '+30 days' },
					{ id: 'reset', label: 'Reset traffic' },
					{ id: 'revoke', label: 'Revoke sub' }
				]
			: [{ id: 'save', label: 'Save changes' }]
	);

	const menu = $derived(
		user
			? [
					{
						label: 'Copy subscription',
						onclick: () => void copy('Subscription URL', user.subscriptionUrl)
					},
					{ label: 'Copy VLESS UUID', onclick: () => void copy('VLESS UUID', user.vlessUuid) },
					{ label: 'Copy short UUID', onclick: () => void copy('Short UUID', user.shortUuid) },
					{ label: 'Delete', variant: 'destructive' as const, onclick: ondelete }
				]
			: []
	);

	async function copy(label: string, value: string) {
		if (await copyText(value)) appToast.success(`${label} copied`);
		else appToast.error(`Could not copy ${label.toLowerCase()}`);
	}

	async function save() {
		if (!user) return;
		await onupdate({
			id: user.id,
			expireAt: dateInputToExpireIso(expireDate),
			tag: tag.trim() || undefined,
			trafficLimitBytes: trafficToBytes(amount, unit)
		});
	}

	async function runAction() {
		switch (actionId) {
			case 'enable':
				onenable();
				return;
			case 'disable':
				ondisable();
				return;
			case 'extend':
				onextend();
				return;
			case 'reset':
				onreset();
				return;
			case 'revoke':
				onrevoke();
				return;
			default:
				await save();
		}
	}
</script>

<DetailSheet
	{open}
	title={user ? `Edit ${user.username}` : 'Edit user'}
	description={user ? `#${user.id} · ${user.shortUuid}` : ''}
	{menu}
	{actions}
	bind:actionId
	{pending}
	onrun={runAction}
	onOpenChange={(value) => {
		if (!value) {
			user = null;
			loadedId = null;
		}
	}}
>
	{#if user}
		<div class="space-y-6">
			<section class="space-y-3">
				<div class="flex flex-wrap items-center gap-2">
					<StatusBadge value={user.status} />
					<ExpireLabel value={user.expireAt} />
				</div>
				<div class={chrome.tile}>
					<TrafficBar used={user.userTraffic?.usedTrafficBytes} limit={user.trafficLimitBytes} />
				</div>
			</section>

			<section class="grid grid-cols-2 gap-3">
				<div class={chrome.tile}>
					<p class={chrome.hint}>Email</p>
					<p class="mt-1 truncate text-sm font-medium text-[var(--color-text-primary)]">{user.email ?? '—'}</p>
				</div>
				<div class={chrome.tile}>
					<p class={chrome.hint}>Telegram</p>
					<p class="mt-1 truncate text-sm font-medium text-[var(--color-text-primary)]">{user.telegramId ?? '—'}</p>
				</div>
			</section>

			{#if user.activeInternalSquads?.length}
				<section>
					<p class={chrome.section}>Internal squads</p>
					<p class="mt-1.5 text-[13px] leading-5 text-[var(--color-text-secondary)]">
						{user.activeInternalSquads.map((squad) => squad.name).join(', ')}
					</p>
				</section>
			{/if}

			<section class={chrome.stack}>
				<p class={chrome.section}>Record</p>
				<label class={chrome.field}>
					<span class={chrome.label}>Expires on</span>
					<Input type="date" bind:value={expireDate} />
				</label>
				<div class={chrome.field}>
					<span class={chrome.label}>Traffic limit</span>
					<div class="flex gap-2">
						<Input bind:value={amount} inputmode="decimal" />
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
					<Input bind:value={tag} />
				</label>
			</section>
		</div>
	{/if}
</DetailSheet>
