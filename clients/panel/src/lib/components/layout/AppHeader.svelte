<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		ChevronRight,
		LayoutDashboard,
		Plus,
		RefreshCw,
		RotateCcw,
		Save,
		SlidersHorizontal
	} from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import SidebarToggle from './SidebarToggle.svelte';
	import HeaderPills from './HeaderPills.svelte';
	import HeaderSearch from './HeaderSearch.svelte';
	import { pageChrome } from '$lib/features/layout/page-chrome.svelte';
	import { contentWidth } from '$lib/features/layout/content-width.svelte';
	import { navItemByHref, navTrail, systemHubByPath } from '$lib/nav';

	const current = $derived(navItemByHref(page.url.pathname));
	const crumbs = $derived(navTrail(page.url.pathname, pageChrome.title));
	const CurrentIcon = $derived(current?.icon);
	const heading = $derived(crumbs.at(-1)?.label ?? current?.label ?? 'Home');
	const isHome = $derived(page.url.pathname === '/dashboard/home');
	const homeTab = $derived(page.url.searchParams.get('tab') === 'explore' ? 'explore' : 'overview');
	const hub = $derived(systemHubByPath(page.url.pathname));
	const hubTab = $derived(page.url.searchParams.get('tab') ?? hub?.defaultTab ?? '');
	const showSubbar = $derived(!isHome && Boolean(pageChrome.description || pageChrome.actions));
	const ActionIcon = $derived(pageChrome.action?.icon === 'restart' ? RotateCcw : RefreshCw);

	function setTab(tab: string) {
		const url = new URL(page.url.href);
		if (isHome) {
			if (tab === 'explore') url.searchParams.set('tab', 'explore');
			else url.searchParams.delete('tab');
		} else {
			url.searchParams.set('tab', tab);
			if (tab !== 'sessions') {
				url.searchParams.delete('kind');
				url.searchParams.delete('id');
				url.searchParams.delete('drop');
			}
		}
		void goto(url.pathname + url.search, { replaceState: true, noScroll: true, keepFocus: true });
	}
</script>

<header
	class="relative z-20 grid h-[49px] shrink-0 grid-cols-[minmax(0,1fr)_minmax(14rem,36rem)_minmax(0,1fr)] items-center gap-2 border-b border-[var(--app-border)] bg-[var(--color-bg)] px-4 sm:px-6"
>
	<div class="flex min-w-0 items-center gap-2">
		<SidebarToggle />
		{#if CurrentIcon}
			<CurrentIcon size={16} class="shrink-0 text-[var(--color-text-secondary)]" />
		{/if}
		<nav class="flex min-w-0 items-center gap-1 text-sm">
			{#each crumbs as crumb, i}
				{#if i > 0}
					<ChevronRight size={14} class="shrink-0 text-[var(--color-text-tertiary)]" />
				{/if}
				{#if crumb.href && i < crumbs.length - 1}
					<a
						href={crumb.href}
						class="truncate text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]"
					>
						{crumb.label}
					</a>
				{:else}
					<h1 class="truncate font-medium text-[var(--color-text-primary)]">{crumb.label}</h1>
				{/if}
			{/each}
		</nav>
	</div>

	<div class="flex min-w-0 justify-center">
		<div class="w-full">
			<HeaderSearch />
		</div>
	</div>

	<div class="flex shrink-0 flex-nowrap items-center justify-end gap-1.5">
		{#if pageChrome.action}
			<Button
				variant="outline"
				size="icon-sm"
				title={pageChrome.action.label}
				aria-label={pageChrome.action.label}
				disabled={Boolean(pageChrome.action.disabled?.() || pageChrome.action.pending?.())}
				onclick={() => void pageChrome.action?.onclick()}
			>
				<ActionIcon size={16} class={pageChrome.action.pending?.() ? 'animate-spin' : ''} />
			</Button>
		{/if}
		{#if pageChrome.create}
			<Button
				variant="outline"
				size="icon-sm"
				title={pageChrome.create.label}
				aria-label={pageChrome.create.label}
				onclick={() => pageChrome.create?.onclick()}
			>
				<Plus size={16} />
			</Button>
		{/if}
		{#if isHome}
			<HeaderPills
				value={homeTab}
				onselect={setTab}
				options={[
					{ value: 'overview', label: 'Overview', icon: LayoutDashboard },
					{ value: 'explore', label: 'Explore', icon: SlidersHorizontal }
				]}
			/>
		{:else if hub}
			<HeaderPills value={hubTab} onselect={setTab} options={hub.tabs} />
		{/if}
		{#if pageChrome.save}
			<Button
				variant="outline"
				size="icon-sm"
				title="Save"
				aria-label="Save"
				disabled={Boolean(pageChrome.save.disabled?.() || pageChrome.save.pending?.())}
				onclick={() => void pageChrome.save?.onclick()}
			>
				<Save size={16} />
			</Button>
		{/if}
	</div>
</header>

{#if showSubbar}
	<div class="shrink-0 border-b border-[var(--app-border)] bg-[var(--color-bg-secondary)]/35">
		<div
			class="mx-auto flex w-full {contentWidth.value === 'wide'
				? 'max-w-none'
				: 'max-w-[1440px]'} flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between sm:px-6"
		>
			<div class="min-w-0">
				<p class="truncate text-xs font-medium text-[var(--color-text-primary)]">{heading}</p>
				{#if pageChrome.description}
					<p class="truncate text-[11px] text-[var(--color-text-tertiary)]">{pageChrome.description}</p>
				{/if}
			</div>
			{#if pageChrome.actions}
				<div class="flex shrink-0 flex-wrap items-center gap-2">
					{@render pageChrome.actions()}
				</div>
			{/if}
		</div>
	</div>
{/if}
