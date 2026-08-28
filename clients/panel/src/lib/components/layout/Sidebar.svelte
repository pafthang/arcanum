<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { cubicOut } from 'svelte/easing';
	import { ArrowLeft, ChevronDown, LogOut, Settings } from 'lucide-svelte';
	import * as Sidebar from '$lib/components/ui/sidebar';
	import * as Popover from '$lib/components/ui/popover';
	import { NAV, SETTINGS_PATH, isSettingsPath } from '$lib/nav';
	import { SETTINGS_NAV } from '$lib/features/settings/nav';
	import { prefetchRoute } from '$lib/features/layout/prefetch';
	import { rw, session } from '$lib/rw';

	let { onnavigate }: { onnavigate?: () => void } = $props();

	const sidebar = Sidebar.useSidebar();
	const currentPath = $derived(page.url.pathname);
	const inSettings = $derived(isSettingsPath(currentPath));
	const iconCollapsed = $derived(sidebar.state === 'collapsed' && !sidebar.isMobile);
	const sections = $derived(
		inSettings
			? SETTINGS_NAV.map((group) => ({
					id: `settings-${group.label}`,
					header: group.label,
					items: group.items.map((item) => ({
						id: item.href,
						label: item.label,
						href: item.href,
						icon: item.icon,
						exact: item.exact
					}))
				}))
			: NAV.map((section) => ({
					...section,
					items: section.items.map((item) => ({ ...item, exact: false as boolean | undefined }))
				}))
	);

	function slideFade(node: HTMLElement, params: { duration?: number } = {}) {
		const duration = params.duration ?? 200;
		const h = node.offsetHeight;
		return {
			duration,
			easing: cubicOut,
			css: (t: number) => `overflow: hidden; height: ${t * h}px; opacity: ${t}`
		};
	}

	function initCollapsed(key: string): boolean {
		if (typeof localStorage === 'undefined') return false;
		return localStorage.getItem(`sidebar_${key}`) === 'collapsed';
	}

	function toggleSection(key: string, current: boolean): boolean {
		const next = !current;
		localStorage.setItem(`sidebar_${key}`, next ? 'collapsed' : 'expanded');
		return next;
	}

	let collapsed = $state<Record<string, boolean>>(
		Object.fromEntries([
			...NAV.map((section) => [section.id, initCollapsed(section.id)] as [string, boolean]),
			...SETTINGS_NAV.map(
				(group) => [`settings-${group.label}`, initCollapsed(`settings-${group.label}`)] as [string, boolean]
			)
		])
	);

	function isActive(href: string, exact?: boolean): boolean {
		if (exact) return currentPath === href;
		return currentPath === href || currentPath.startsWith(href + '/');
	}

	function sectionVisible(id: string): boolean {
		return iconCollapsed || !collapsed[id];
	}

	function navigate() {
		sidebar.setOpenMobile(false);
		onnavigate?.();
	}

	function logout() {
		rw.auth.logout();
		goto('/auth/login');
	}
</script>

<Sidebar.Root
	collapsible="icon"
	class="h-full min-h-0 overflow-hidden border-r border-[var(--app-border)] bg-[var(--color-bg-secondary)]"
>
	<Sidebar.Header class="h-[49px] justify-center border-b border-[var(--app-border)] px-2">
		{#if inSettings}
			<a
				href="/dashboard/home"
				onclick={navigate}
				onpointerenter={() => prefetchRoute('/dashboard/home')}
				title="Back"
				class="flex items-center gap-2 rounded-md px-2 py-1.5 text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]"
			>
				<ArrowLeft class="size-4 shrink-0 text-[var(--color-text-secondary)]" />
				<span class="truncate text-sm font-medium group-data-[collapsible=icon]:hidden">Settings</span>
			</a>
		{:else}
			<a
				href="/dashboard/home"
				onclick={navigate}
				onpointerenter={() => prefetchRoute('/dashboard/home')}
				class="flex items-center gap-2 rounded-md px-2 py-1.5 text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]"
			>
				<span class="flex min-w-0 items-center gap-2">
					<svg
						class="size-4 shrink-0 text-[var(--app-accent)]"
						viewBox="0 0 16 16"
						fill="none"
						xmlns="http://www.w3.org/2000/svg"
						aria-hidden="true"
					>
						<path
							fill-rule="evenodd"
							clip-rule="evenodd"
							d="M8 1a.75.75 0 0 1 .75.75v12.5a.75.75 0 0 1-1.5 0V1.75A.75.75 0 0 1 8 1Zm6 2a.75.75 0 0 1 .75.75v8.5a.75.75 0 0 1-1.5 0v-8.5A.75.75 0 0 1 14 3ZM5 4a.75.75 0 0 1 .75.75v6.5a.75.75 0 0 1-1.5 0v-6.5A.75.75 0 0 1 5 4Zm6 1a.75.75 0 0 1 .75.75v4.5a.75.75 0 0 1-1.5 0v-4.5A.75.75 0 0 1 11 5ZM2 6a.75.75 0 0 1 .75.75v2.5a.75.75 0 0 1-1.5 0v-2.5A.75.75 0 0 1 2 6Z"
							fill="currentColor"
						/>
					</svg>
					<span class="truncate text-sm font-semibold group-data-[collapsible=icon]:hidden">
						<span class="text-[var(--app-accent)]">Optima</span>Wave
					</span>
				</span>
			</a>
		{/if}
	</Sidebar.Header>

	<Sidebar.Content class="px-1 py-2 pb-3">
		{#each sections as section (section.id)}
			<Sidebar.Group class="py-1">
				<button
					type="button"
					onclick={() => (collapsed[section.id] = toggleSection(section.id, collapsed[section.id]))}
					class="flex h-7 w-full items-center px-2 text-[11px] font-semibold uppercase tracking-wide text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] group-data-[collapsible=icon]:hidden"
				>
					<span>{section.header}</span>
					<ChevronDown
						size={12}
						class="ml-auto text-[var(--color-text-tertiary)] transition-transform {collapsed[section.id]
							? '-rotate-90'
							: ''}"
					/>
				</button>
				{#if sectionVisible(section.id)}
					<div transition:slideFade>
						<Sidebar.GroupContent>
							<Sidebar.Menu>
								{#each section.items as item (item.id)}
									{@const Icon = item.icon}
									{@const active = isActive(item.href, item.exact)}
									<Sidebar.MenuItem>
										<Sidebar.MenuButton
											isActive={active}
											tooltipContent={item.label}
											class={active
												? 'bg-[var(--color-bg-hover)]/50 text-[var(--color-text-primary)]'
												: 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-hover)] hover:text-[var(--color-text-primary)]'}
										>
											{#snippet child({ props })}
												<a
													href={item.href}
													{...props}
													onclick={navigate}
													onpointerenter={() => prefetchRoute(item.href)}
													onfocus={() => prefetchRoute(item.href)}
												>
													<Icon />
													<span>{item.label}</span>
												</a>
											{/snippet}
										</Sidebar.MenuButton>
									</Sidebar.MenuItem>
								{/each}
							</Sidebar.Menu>
						</Sidebar.GroupContent>
					</div>
				{/if}
			</Sidebar.Group>
		{/each}
	</Sidebar.Content>

	<Sidebar.Footer class="border-t border-[var(--app-border)]">
		<div class="flex items-center gap-2 px-1 py-1">
			<Popover.Root>
				<Popover.Trigger
					class="flex min-w-0 flex-1 items-center gap-2 rounded-md px-1 py-1 hover:bg-[var(--color-bg-hover)]"
				>
					<div
						class="flex size-7 shrink-0 items-center justify-center rounded-full bg-[var(--app-accent)] text-xs font-bold text-[var(--app-accent-foreground)]"
					>
						A
					</div>
					<div class="min-w-0 flex-1 text-left group-data-[collapsible=icon]:hidden">
						<p class="truncate text-sm font-medium text-[var(--color-text-primary)]">
							{session.isAuthenticated ? 'Admin' : 'Guest'}
						</p>
						<p class="truncate text-xs text-[var(--color-text-secondary)]">Signed in</p>
					</div>
				</Popover.Trigger>
				<Popover.Content side="top" align="start" class="w-52 p-1">
					<a
						href={SETTINGS_PATH}
						onclick={navigate}
						class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]"
					>
						<Settings size={14} />
						Settings
					</a>
					<div class="mt-1 border-t border-[var(--app-border)] pt-1">
						<button
							type="button"
							onclick={logout}
							class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm text-[var(--color-text-primary)] hover:bg-[var(--color-bg-hover)]"
						>
							<LogOut size={14} />
							Log out
						</button>
					</div>
				</Popover.Content>
			</Popover.Root>
		</div>
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
